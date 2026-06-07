package theme

import "strings"

// Match resolves the style for a token with the given scope stack.
// scopes[0] is the root scope, scopes[len-1] is the most specific.
// Each property (foreground, background, fontStyle) is taken from the
// highest-scoring matching rule that defines it.
func (t *Theme) Match(scopes []string) TokenSettings {
	result := TokenSettings{FontStyle: FontStyleNotSet}
	var fgScore, bgScore, fsScore matchScore
	var fgSet, bgSet, fsSet bool

	for _, rule := range t.TokenColors {
		score, ok := bestSelectorScore(rule.Scopes, scopes)
		if !ok {
			continue
		}
		s := rule.Settings
		if s.Foreground != "" && (!fgSet || score.greaterThan(fgScore)) {
			result.Foreground = s.Foreground
			fgScore = score
			fgSet = true
		}
		if s.Background != "" && (!bgSet || score.greaterThan(bgScore)) {
			result.Background = s.Background
			bgScore = score
			bgSet = true
		}
		if s.FontStyle != FontStyleNotSet && (!fsSet || score.greaterThan(fsScore)) {
			result.FontStyle = s.FontStyle
			fsScore = score
			fsSet = true
		}
	}
	return result
}

func bestSelectorScore(selectors []string, scopeStack []string) (matchScore, bool) {
	var best matchScore
	var found bool
	for _, sel := range selectors {
		if s, ok := scoreSelector(sel, scopeStack); ok {
			if !found || s.greaterThan(best) {
				best = s
				found = true
			}
		}
	}
	return best, found
}

type matchScore struct {
	depth      int // index of the deepest matched scope in the stack
	scopeDepth int // dot-segments in the rule's target scope (last selector part)
	parents    int // number of parent scope parts in the selector
}

func (s matchScore) greaterThan(other matchScore) bool {
	if s.depth != other.depth {
		return s.depth > other.depth
	}
	if s.scopeDepth != other.scopeDepth {
		return s.scopeDepth > other.scopeDepth
	}
	return s.parents > other.parents
}

// scoreSelector scores a selector against a scope stack.
// A selector may contain spaces for parent scope matching
// (e.g. "source.go keyword" requires "keyword" to appear below "source.go").
func scoreSelector(selector string, scopeStack []string) (matchScore, bool) {
	parts := strings.Fields(selector)
	if len(parts) == 0 {
		return matchScore{}, false
	}

	partIdx := len(parts) - 1
	var depth int
	scopeDepth := strings.Count(parts[len(parts)-1], ".") + 1

	for stackIdx := len(scopeStack) - 1; stackIdx >= 0 && partIdx >= 0; stackIdx-- {
		if scopePrefixMatch(parts[partIdx], scopeStack[stackIdx]) {
			if partIdx == len(parts)-1 {
				depth = stackIdx
			}
			partIdx--
		}
	}

	if partIdx >= 0 {
		return matchScore{}, false
	}

	return matchScore{
		depth:      depth,
		scopeDepth: scopeDepth,
		parents:    len(parts),
	}, true
}

// scopePrefixMatch checks if selector is a prefix of scope on dot boundaries.
// "keyword" matches "keyword", "keyword.control", "keyword.control.go"
// but not "keywordx" or "keywords".
func scopePrefixMatch(selector, scope string) bool {
	if scope == selector {
		return true
	}
	return len(scope) > len(selector) &&
		scope[:len(selector)] == selector &&
		scope[len(selector)] == '.'
}
