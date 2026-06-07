package tokenizer

import (
	"context"

	"github.com/frostybee/nuri/internal/grammar"
	"github.com/frostybee/nuri/internal/oniguruma"
)

func tokenizeLine(
	ctx context.Context,
	line []byte,
	g *grammar.Grammar,
	onigLib oniguruma.OnigLib,
	state *StateStack,
	resolver grammar.GrammarResolver,
	injections []grammar.Injection,
	cache *scannerCache,
	startPos int,
	isFirstLine bool,
) ([]Token, *StateStack, error) {

	builder := newLineTokenBuilder(startPos)
	lineLen := len(line)
	anchorPosition := -1

	cc := &captureContext{
		ctx:        ctx,
		line:       line,
		g:          g,
		onigLib:    onigLib,
		resolver:   resolver,
		injections: injections,
		cache:      cache,
	}

	// Strip trailing newline for the purpose of token text,
	// but keep the full line for regex matching (anchors need it).
	tokenEnd := lineLen
	if tokenEnd > 0 && line[tokenEnd-1] == '\n' {
		tokenEnd--
	}

	// Phase A: check while conditions (only for top-level calls, not retokenization)
	if startPos == 0 {
		wcr := checkWhileConditions(ctx, line, isFirstLine, g, onigLib, state, builder, cache, cc)
		state = wcr.state
		startPos = wcr.linePos
		anchorPosition = wcr.anchorPosition
		isFirstLine = wcr.isFirstLine
	}

	pos := startPos
loop:
	for pos < lineLen {
		activeRules, endRule := getActivePatterns(state, g)

		compileG := g
		if top := state.top(); top.ContentGrammar != nil {
			compileG = top.ContentGrammar
		}
		compiled, err := grammar.CompilePatterns(activeRules, compileG, compileG.Repository, nil, resolver)
		if err != nil {
			break
		}

		if endRule != nil && endRule.EndPattern != "" {
			endCR := grammar.CompiledRule{
				Pattern: []byte(endRule.EndPattern),
				Rule:    endRule,
			}
			if endRule.Parent != nil && endRule.Parent.ApplyEndPatternLast {
				compiled.Rules = append(compiled.Rules, endCR)
			} else {
				compiled.Rules = append([]grammar.CompiledRule{endCR}, compiled.Rules...)
			}
		}

		searchOpts := computeSearchOptions(isFirstLine, pos, anchorPosition)
		mr, err := findNextMatch(ctx, onigLib, compiled, line, pos, cache, searchOpts)
		if err != nil {
			break
		}

		injMatch, injPriority, injErr := matchInjections(ctx, onigLib, injections, g, state, line, pos, resolver, cache, searchOpts)
		if injErr != nil {
			break
		}

		mr = pickBestMatch(mr, injMatch, injPriority)

		if mr == nil {
			if pos < tokenEnd {
				builder.produce(tokenEnd, state.scopeSlice())
			}
			break
		}

		matchStart := mr.match.Captures[0].Start
		matchEnd := mr.match.Captures[0].End

		if matchStart > pos {
			gapEnd := matchStart
			if gapEnd > tokenEnd {
				gapEnd = tokenEnd
			}
			builder.produce(gapEnd, state.scopeSlice())
		}

		hasAdvanced := matchEnd > pos

		switch rule := mr.rule.(type) {
		case *grammar.EndRule:
			poppedFrame := *state.top()
			handleEndRule(rule, mr.match, state, builder, cc)
			anchorPosition = poppedFrame.AnchorPosition

			// Check [1]: EndRule popped without advancing, frame entered at same pos
			// (vscode-textmate lines 148-163).
			if !hasAdvanced && poppedFrame.EnterPosition == pos {
				state.push(poppedFrame)
				builder.produce(tokenEnd, state.scopeSlice())
				break loop
			}

		case *grammar.MatchRule:
			handleMatchRule(rule, mr.match, state, builder, cc)

			// Check [4]: MatchRule without advancement (vscode-textmate lines 317-328).
			if !hasAdvanced {
				state.safePop()
				builder.produce(tokenEnd, state.scopeSlice())
				break loop
			}

		case *grammar.BeginEndRule:
			handleBeginRule(rule, mr.match, line, state, builder, pos, cc, mr.ruleGrammar)
			anchorPosition = matchEnd

			// Check [2]: BeginEndRule pushed same rule without advancing
			// (vscode-textmate lines 230-241).
			if !hasAdvanced && state.topHasSameRuleBelow() {
				state.pop()
				builder.produce(tokenEnd, state.scopeSlice())
				break loop
			}

		case *grammar.BeginWhileRule:
			handleBeginWhileRule(rule, mr.match, line, state, builder, pos, cc, mr.ruleGrammar)
			anchorPosition = matchEnd

			// Check [3]: BeginWhileRule pushed same rule without advancing
			// (vscode-textmate lines 279-290).
			if !hasAdvanced && state.topHasSameRuleBelow() {
				state.pop()
				builder.produce(tokenEnd, state.scopeSlice())
				break loop
			}
		}

		if matchEnd > pos {
			pos = matchEnd
			isFirstLine = false
		}
	}
	if len(line) > 0 && line[len(line)-1] == '\n' {
		builder.finalize(len(line))
	}
	return builder.finish(), state, nil
}

func getActivePatterns(state *StateStack, g *grammar.Grammar) ([]grammar.Rule, *grammar.EndRule) {
	top := state.top()
	if top.Rule == nil {
		return g.Patterns, nil
	}

	switch r := top.Rule.(type) {
	case *grammar.BeginEndRule:
		return r.Patterns, top.EndRule
	case *grammar.BeginWhileRule:
		return r.Patterns, nil
	case *grammar.CollectionRule:
		return r.Patterns, nil
	default:
		return g.Patterns, nil
	}
}

func computeSearchOptions(isFirstLine bool, pos, anchorPosition int) oniguruma.SearchOptions {
	opts := oniguruma.SearchOptionNone
	if !isFirstLine {
		opts |= oniguruma.SearchOptionNotBeginString
	}
	if pos != anchorPosition {
		opts |= oniguruma.SearchOptionNotBeginPosition
	}
	return opts
}

func resolveScopeName(name string, captures []oniguruma.Capture, line []byte) string {
	if name == "" {
		return name
	}
	captureTexts := extractCaptureTexts(captures, line)
	return grammar.ResolveScopeBackrefs(name, captureTexts)
}

func handleMatchRule(
	rule *grammar.MatchRule,
	match *oniguruma.Match,
	state *StateStack,
	builder *lineTokenBuilder,
	cc *captureContext,
) {
	scopes := state.scopeSlice()
	if rule.Name != "" {
		resolved := resolveScopeName(rule.Name, match.Captures, cc.line)
		scopes = appendScopes(scopes, resolved)
	}

	if len(rule.Captures) > 0 {
		handleCaptures(match.Captures, rule.Captures, scopes, builder, cc)
	} else {
		builder.produce(match.Captures[0].End, scopes)
	}
}

func handleBeginRule(
	rule *grammar.BeginEndRule,
	match *oniguruma.Match,
	line []byte,
	state *StateStack,
	builder *lineTokenBuilder,
	pos int,
	cc *captureContext,
	ruleGrammar *grammar.Grammar,
) {
	captureTexts := extractCaptureTexts(match.Captures, line)

	// Resolve scope names against captures
	resolvedName := grammar.ResolveScopeBackrefs(rule.Name, captureTexts)
	resolvedContentName := grammar.ResolveScopeBackrefs(rule.ContentName, captureTexts)

	// Build scopes with resolved name pushed
	scopes := state.scopeSlice()
	if resolvedName != "" {
		scopes = appendScopes(scopes, resolvedName)
	}

	// Handle begin captures
	if len(rule.BeginCaptures) > 0 {
		handleCaptures(match.Captures, rule.BeginCaptures, scopes, builder, cc)
	} else {
		builder.produce(match.Captures[0].End, scopes)
	}

	// Create EndRule with backref substitution
	endPattern := grammar.ResolveBackrefs(rule.End, captureTexts)
	endRule := &grammar.EndRule{
		ID:          rule.ID + 1000000, // synthetic ID
		Parent:      rule,
		EndPattern:  endPattern,
		EndCaptures: rule.EndCaptures,
	}

	matchEnd := match.Captures[0].End
	state.pushBeginEnd(rule, endRule, matchEnd, pos, resolvedName, resolvedContentName, ruleGrammar)
	state.top().BeginCapturedEOL = matchEnd == len(line)
}

func handleEndRule(
	rule *grammar.EndRule,
	match *oniguruma.Match,
	state *StateStack,
	builder *lineTokenBuilder,
	cc *captureContext,
) {
	top := state.top()

	// Pop content scope by building scopes without it
	scopes := state.scopeSlice()
	if top.ContentScope != "" {
		scopes = scopes[:len(scopes)-countScopes(top.ContentScope)]
	}

	// Handle end captures
	if len(rule.EndCaptures) > 0 {
		handleCaptures(match.Captures, rule.EndCaptures, scopes, builder, cc)
	} else {
		builder.produce(match.Captures[0].End, scopes)
	}

	state.pop()
}

func handleBeginWhileRule(
	rule *grammar.BeginWhileRule,
	match *oniguruma.Match,
	line []byte,
	state *StateStack,
	builder *lineTokenBuilder,
	pos int,
	cc *captureContext,
	ruleGrammar *grammar.Grammar,
) {
	captureTexts := extractCaptureTexts(match.Captures, line)

	resolvedName := grammar.ResolveScopeBackrefs(rule.Name, captureTexts)
	resolvedContentName := grammar.ResolveScopeBackrefs(rule.ContentName, captureTexts)

	scopes := state.scopeSlice()
	if resolvedName != "" {
		scopes = appendScopes(scopes, resolvedName)
	}

	if len(rule.BeginCaptures) > 0 {
		handleCaptures(match.Captures, rule.BeginCaptures, scopes, builder, cc)
	} else {
		builder.produce(match.Captures[0].End, scopes)
	}

	whilePattern := grammar.ResolveBackrefs(rule.While, captureTexts)
	whileRule := &grammar.WhileRule{
		ID:            rule.ID + 2000000,
		Parent:        rule,
		WhilePattern:  whilePattern,
		WhileCaptures: rule.WhileCaptures,
	}

	matchEnd := match.Captures[0].End
	state.pushBeginWhile(rule, whileRule, matchEnd, pos, resolvedName, resolvedContentName, ruleGrammar)
	state.top().BeginCapturedEOL = matchEnd == len(line)
}

type whileCheckResult struct {
	state          *StateStack
	linePos        int
	anchorPosition int
	isFirstLine    bool
}

// checkWhileConditions checks while-rule conditions at the start of a line.
// Matches vscode-textmate's _checkWhileConditions (tokenizeString.ts:345-403).
func checkWhileConditions(
	ctx context.Context,
	line []byte,
	isFirstLine bool,
	g *grammar.Grammar,
	onigLib oniguruma.OnigLib,
	state *StateStack,
	builder *lineTokenBuilder,
	cache *scannerCache,
	cc *captureContext,
) whileCheckResult {
	linePos := 0
	anchorPosition := -1
	if state.top().BeginCapturedEOL {
		anchorPosition = 0
	}

	// vscode-textmate appends "\n" before tokenizing, so while-condition
	// patterns like ^[\t ]*$ can match after the content newline. We do
	// the same for the while-check scanner input only.
	whileScanLine := append(line[:len(line):len(line)], '\n')

	// Collect while-rule frame indices from bottom to top
	var whileIndices []int
	for i := range state.frames {
		if state.frames[i].WhileRule != nil {
			whileIndices = append(whileIndices, i)
		}
	}

	// Check in outermost-first order (matching vscode-textmate's reversed pop loop)
	for _, idx := range whileIndices {
		wr := state.frames[idx].WhileRule

		compiled := &grammar.CompileResult{
			Rules: []grammar.CompiledRule{{
				Pattern: []byte(wr.WhilePattern),
				Rule:    wr,
			}},
		}

		whileOpts := computeSearchOptions(isFirstLine, linePos, anchorPosition)
		mr, err := findNextMatch(ctx, onigLib, compiled, whileScanLine, linePos, cache, whileOpts)
		if err != nil || mr == nil {
			// While condition failed — pop back to this frame's parent
			for state.depth() > idx {
				state.pop()
			}
			break
		}

		// While matched — emit tokens around captures (vscode-textmate lines 383-393)
		if len(mr.match.Captures) > 0 {
			captureStart := mr.match.Captures[0].Start
			captureEnd := mr.match.Captures[0].End

			// Scope for this while-rule's frame
			scopes := state.scopeSliceTo(idx)

			builder.produce(captureStart, scopes)
			if len(wr.WhileCaptures) > 0 {
				handleCaptures(mr.match.Captures, wr.WhileCaptures, scopes, builder, cc)
			}
			builder.produce(captureEnd, scopes)

			anchorPosition = captureEnd
			if captureEnd > linePos {
				linePos = captureEnd
				isFirstLine = false
			}
		}
	}

	return whileCheckResult{
		state:          state,
		linePos:        linePos,
		anchorPosition: anchorPosition,
		isFirstLine:    isFirstLine,
	}
}

// pickBestMatch returns the winning match between a grammar match and an
// injection match. Earlier start position wins. On a tie, L: (PriorityLeft)
// lets the injection win; otherwise the grammar match wins.
func pickBestMatch(grammarMatch, injMatch *matchResult, injPriority grammar.Priority) *matchResult {
	if injMatch == nil {
		return grammarMatch
	}
	if grammarMatch == nil {
		return injMatch
	}

	gStart := grammarMatch.match.Captures[0].Start
	iStart := injMatch.match.Captures[0].Start

	if iStart < gStart {
		return injMatch
	}
	if gStart < iStart {
		return grammarMatch
	}

	if injPriority == grammar.PriorityLeft {
		return injMatch
	}
	return grammarMatch
}

// extractCaptureTexts extracts the text of each capture group.
func extractCaptureTexts(captures []oniguruma.Capture, line []byte) []string {
	texts := make([]string, len(captures))
	for i, c := range captures {
		if c.Start >= 0 && c.End >= 0 && line != nil {
			texts[i] = string(line[c.Start:c.End])
		}
	}
	return texts
}
