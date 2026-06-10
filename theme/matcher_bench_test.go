package theme

import "testing"

// BenchmarkThemeMatch measures scope-stack style resolution against a real
// theme. With pre-compiled selectors the compiled path is 0 allocs/op.
func BenchmarkThemeMatch(b *testing.B) {
	th := loadTestTheme(b, "github-dark")

	stacks := [][]string{
		{"source.go", "keyword.package.go"},
		{"source.go", "string.quoted.double.go"},
		{"source.go", "comment.line.double-slash.go"},
		{"source.go", "meta.function.declaration.go", "entity.name.function.go"},
		{"text.html.markdown", "markup.fenced_code.block.markdown", "source.js", "meta.function.js", "entity.name.function.js"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		th.Match(stacks[i%len(stacks)])
	}
}

// BenchmarkThemeMatchFallback measures the same resolution on a hand-built
// theme (no compiled selectors), i.e. the per-call parsing fallback path.
func BenchmarkThemeMatchFallback(b *testing.B) {
	parsed := loadTestTheme(b, "github-dark")

	// Clone with only the exported fields — compiled stays nil.
	fallback := &Theme{
		Name:              parsed.Name,
		Type:              parsed.Type,
		DefaultForeground: parsed.DefaultForeground,
		DefaultBackground: parsed.DefaultBackground,
	}
	for _, tc := range parsed.TokenColors {
		fallback.TokenColors = append(fallback.TokenColors, TokenColor{
			Scopes:   tc.Scopes,
			Settings: tc.Settings,
		})
	}

	stacks := [][]string{
		{"source.go", "keyword.package.go"},
		{"source.go", "string.quoted.double.go"},
		{"source.go", "comment.line.double-slash.go"},
		{"source.go", "meta.function.declaration.go", "entity.name.function.go"},
		{"text.html.markdown", "markup.fenced_code.block.markdown", "source.js", "meta.function.js", "entity.name.function.js"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		fallback.Match(stacks[i%len(stacks)])
	}
}
