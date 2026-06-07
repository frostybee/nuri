package theme

import "testing"

func TestScopePrefixMatch(t *testing.T) {
	tests := []struct {
		selector, scope string
		want            bool
	}{
		{"keyword", "keyword", true},
		{"keyword", "keyword.control", true},
		{"keyword", "keyword.control.go", true},
		{"keyword", "keywordx", false},
		{"keyword", "keywords", false},
		{"keyword", "key", false},
		{"entity.name", "entity.name", true},
		{"entity.name", "entity.name.function", true},
		{"entity.name", "entity.name.function.go", true},
		{"entity.name", "entity.namex", false},
		{"entity.name", "entity", false},
		{"source.go", "source.go", true},
		{"source.go", "source.goo", false},
		{"source.go", "source.golang", false},
	}
	for _, tt := range tests {
		got := scopePrefixMatch(tt.selector, tt.scope)
		if got != tt.want {
			t.Errorf("scopePrefixMatch(%q, %q) = %v, want %v",
				tt.selector, tt.scope, got, tt.want)
		}
	}
}

func TestScoreSelector(t *testing.T) {
	stack := []string{"source.go", "meta.function.declaration.go", "entity.name.function.go"}

	tests := []struct {
		selector   string
		wantOK     bool
		depth      int
		scopeDepth int
		parents    int
	}{
		{"entity.name", true, 2, 2, 1},
		{"entity.name.function", true, 2, 3, 1},
		{"entity", true, 2, 1, 1},
		{"meta.function", true, 1, 2, 1},
		{"source.go entity.name", true, 2, 2, 2},
		{"source.go meta.function entity.name", true, 2, 2, 3},
		{"notfound", false, 0, 0, 0},
		{"entity.name source.go", false, 0, 0, 0}, // wrong order
	}
	for _, tt := range tests {
		score, ok := scoreSelector(tt.selector, stack)
		if ok != tt.wantOK {
			t.Errorf("scoreSelector(%q) ok=%v, want %v", tt.selector, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if score.depth != tt.depth || score.scopeDepth != tt.scopeDepth || score.parents != tt.parents {
			t.Errorf("scoreSelector(%q) = {depth:%d, sd:%d, par:%d}, want {%d, %d, %d}",
				tt.selector, score.depth, score.scopeDepth, score.parents,
				tt.depth, tt.scopeDepth, tt.parents)
		}
	}
}

func TestMatchGitHubDarkKeyword(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	result := th.Match([]string{"source.go", "keyword.package.go"})
	if result.Foreground != "#f97583" {
		t.Errorf("keyword.package.go foreground = %q, want %q", result.Foreground, "#f97583")
	}
}

func TestMatchGitHubDarkString(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	result := th.Match([]string{"source.go", "string.quoted.double.go"})
	if result.Foreground != "#9ecbff" {
		t.Errorf("string foreground = %q, want %q", result.Foreground, "#9ecbff")
	}
}

func TestMatchGitHubDarkComment(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	result := th.Match([]string{"source.go", "comment.line.double-slash.go"})
	if result.Foreground != "#6a737d" {
		t.Errorf("comment foreground = %q, want %q", result.Foreground, "#6a737d")
	}
}

func TestMatchGitHubDarkEntity(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	result := th.Match([]string{"source.go", "entity.name.function.go"})
	if result.Foreground != "#b392f0" {
		t.Errorf("entity.name.function foreground = %q, want %q", result.Foreground, "#b392f0")
	}
}

func TestMatchSpecificityMoreSegmentsWins(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "entity", "settings": {"foreground": "#aaaaaa"}},
			{"scope": "entity.name", "settings": {"foreground": "#bbbbbb"}},
			{"scope": "entity.name.function", "settings": {"foreground": "#cccccc"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	result := th.Match([]string{"source.go", "entity.name.function.go"})
	if result.Foreground != "#cccccc" {
		t.Errorf("foreground = %q, want #cccccc (most specific)", result.Foreground)
	}
}

func TestMatchSpecificityDeeperScopeWins(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "source", "settings": {"foreground": "#aaaaaa"}},
			{"scope": "keyword", "settings": {"foreground": "#bbbbbb"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	result := th.Match([]string{"source.go", "keyword.control"})
	if result.Foreground != "#bbbbbb" {
		t.Errorf("foreground = %q, want #bbbbbb (deeper match wins)", result.Foreground)
	}
}

func TestMatchParentScopeSelector(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "variable", "settings": {"foreground": "#aaaaaa"}},
			{"scope": "string variable", "settings": {"foreground": "#bbbbbb"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// Without parent context: "variable" wins.
	r1 := th.Match([]string{"source.go", "variable.other"})
	if r1.Foreground != "#aaaaaa" {
		t.Errorf("variable without string parent = %q, want #aaaaaa", r1.Foreground)
	}

	// With string parent: "string variable" wins (more parents = higher specificity).
	r2 := th.Match([]string{"source.go", "string.quoted.double", "variable.other"})
	if r2.Foreground != "#bbbbbb" {
		t.Errorf("variable inside string = %q, want #bbbbbb", r2.Foreground)
	}
}

func TestMatchFontStyle(t *testing.T) {
	th := loadTestTheme(t, "github-dark")

	// "markup.bold" has fontStyle: "bold"
	result := th.Match([]string{"text.html", "markup.bold"})
	if !result.FontStyle.Has(FontStyleBold) {
		t.Errorf("markup.bold fontStyle = %v, want bold", result.FontStyle)
	}

	// "markup.italic" has fontStyle: "italic"
	result = th.Match([]string{"text.html", "markup.italic"})
	if !result.FontStyle.Has(FontStyleItalic) {
		t.Errorf("markup.italic fontStyle = %v, want italic", result.FontStyle)
	}
}

func TestMatchFontStyleMergeWithForeground(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "entity", "settings": {"foreground": "#ff0000", "fontStyle": "italic"}},
			{"scope": "entity.name.function", "settings": {"foreground": "#00ff00"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	result := th.Match([]string{"source", "entity.name.function.go"})

	// Foreground from more specific "entity.name.function" rule.
	if result.Foreground != "#00ff00" {
		t.Errorf("foreground = %q, want #00ff00", result.Foreground)
	}
	// FontStyle from less specific "entity" rule (first-non-null-wins).
	if result.FontStyle != FontStyleItalic {
		t.Errorf("fontStyle = %v, want italic", result.FontStyle)
	}
}

func TestMatchFontStyleExplicitNoneResets(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "entity", "settings": {"fontStyle": "italic"}},
			{"scope": "entity.name", "settings": {"fontStyle": ""}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	result := th.Match([]string{"source", "entity.name.function"})
	if result.FontStyle != FontStyleNone {
		t.Errorf("fontStyle = %v, want none (explicit reset)", result.FontStyle)
	}
}

func TestMatchNoMatch(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#ff0000"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	result := th.Match([]string{"source.go", "string.quoted"})
	if result.Foreground != "" {
		t.Errorf("expected empty foreground for non-matching scope, got %q", result.Foreground)
	}
	if result.FontStyle != FontStyleNotSet {
		t.Errorf("expected FontStyleNotSet, got %v", result.FontStyle)
	}
}

func TestMatchFirstRuleWinsOnTie(t *testing.T) {
	data := []byte(`{
		"name": "test",
		"tokenColors": [
			{"scope": "keyword", "settings": {"foreground": "#ff0000"}},
			{"scope": "keyword", "settings": {"foreground": "#00ff00"}}
		]
	}`)
	th, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	result := th.Match([]string{"source", "keyword.control"})
	if result.Foreground != "#ff0000" {
		t.Errorf("foreground = %q, want #ff0000 (first rule wins on tie)", result.Foreground)
	}
}

func TestMatchGitHubDarkVariableInString(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	// variable.other (scopeDepth=2) beats string variable (scopeDepth=1),
	// matching vscode-textmate's trie behavior where deeper nodes shadow shallower ones.
	result := th.Match([]string{"source.ruby", "string.quoted.double", "variable.other"})
	if result.Foreground != "#e1e4e8" {
		t.Errorf("variable.other foreground = %q, want #e1e4e8", result.Foreground)
	}
}

func TestMatchVariableOtherBeatsStringVariable(t *testing.T) {
	th := loadTestTheme(t, "github-dark")
	result := th.Match([]string{
		"source.js",
		"meta.block.js",
		"string.template.js",
		"meta.template.expression.js",
		"meta.embedded.line.js",
		"variable.other.readwrite.js",
	})
	if result.Foreground != "#e1e4e8" {
		t.Errorf("variable.other.readwrite.js foreground = %q, want #e1e4e8", result.Foreground)
	}
}
