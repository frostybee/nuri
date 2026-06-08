package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/core"
	"github.com/frostybee/nuri/transformers"
)

type Block struct {
	Label string
	HTML  template.HTML
}

type Section struct {
	ID          int
	Title       string
	Description string
	Blocks      []Block
	ExtraCSS    template.CSS
	ExtraHTML   template.HTML
}

func must(s string, err error) string {
	if err != nil {
		log.Fatal(err)
	}
	return s
}

// --- Code snippets ---

const goSnippet = `package main

import "fmt"

func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
`

const jsSnippet = `async function fetchUsers(endpoint) {
  const response = await fetch(endpoint);
  const data = await response.json();
  return data.users.filter(u => u.active);
}
`

const goMultiSnippet = `type Server struct {
	Addr    string
	Handler http.Handler
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Addr, s.Handler)
}
`

const jsMultiSnippet = `class EventBus {
  #listeners = new Map();

  on(event, fn) {
    const fns = this.#listeners.get(event) ?? [];
    fns.push(fn);
    this.#listeners.set(event, fns);
  }
}
`

const goHighlightSnippet = `func process(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, strings.ToUpper(item))
	}
	return result
}
`

const jsFocusSnippet = `function middleware(req, res, next) {
  const start = Date.now();
  res.on('finish', () => {
    const ms = Date.now() - start;
    console.log(req.method, req.url, ms + 'ms');
  });
  next();
}
`

const goDiffSnippet = `func Connect(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	db, err := sql.Open("mysql", dsn)
	return db, err
}
`

const jsMetaSnippet = `const config = {
  host: "localhost",
  port: 8080,
  debug: true,
  version: "1.0.0",
};
`

const jsNotationSnippet = `function calculate(a, b) {
  const sum = a + b; // [!code highlight]
  const product = a * b;
  const result = sum + product; // [!code focus]
  console.log("old output"); // [!code --]
  console.log("new output"); // [!code ++]
  return result;
}
`

const dslSnippet = `let name = "world"
set count = 42
if count > 10
  print "hello"
end
`

const goDefaultsSnippet = `func main() {
	fmt.Println("Theme from WithDefaults")
}
`

const goOverrideSnippet = `func main() {
	fmt.Println("Theme overridden per-call")
}
`

const goClassMapSnippet = `var ErrNotFound = errors.New("not found")

func Lookup(id string) (*Record, error) {
	r, ok := store[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}
`

const jsClassMapSnippet = `const PI = 3.14159;

function circleArea(r) {
  return PI * r * r;
}
`

const ansiTestRunnerSnippet = "\x1b[1mRunning tests...\x1b[0m\n" +
	"\x1b[32m  PASS\x1b[0m  TestAuth         \x1b[90m(0.03s)\x1b[0m\n" +
	"\x1b[32m  PASS\x1b[0m  TestLogin        \x1b[90m(0.12s)\x1b[0m\n" +
	"\x1b[31m  FAIL\x1b[0m  TestAPI          \x1b[90m(0.45s)\x1b[0m\n" +
	"\x1b[32m  PASS\x1b[0m  TestWebSocket    \x1b[90m(0.08s)\x1b[0m\n" +
	"\x1b[33m  SKIP\x1b[0m  TestIntegration  \x1b[90m(0.00s)\x1b[0m\n" +
	"\n" +
	"\x1b[1;31m1 failed\x1b[0m, \x1b[32m3 passed\x1b[0m, \x1b[33m1 skipped\x1b[0m"

const ansiBuildSnippet = "\x1b[1;36m  Compiling\x1b[0m nuri v0.1.0\n" +
	"\x1b[1;36m  Compiling\x1b[0m wazero v1.8.2\n" +
	"\x1b[1;31merror\x1b[0m: mismatched types\n" +
	"  \x1b[34m-->\x1b[0m src/engine.go:42:5\n" +
	"   \x1b[90m|\x1b[0m\n" +
	"\x1b[90m42\x1b[0m \x1b[90m|\x1b[0m     return \x1b[38;2;255;128;0m\"invalid\"\x1b[0m\n" +
	"   \x1b[90m|\x1b[0m            \x1b[1;31m^^^^^^^^^\x1b[0m expected int, found string\n" +
	"\n" +
	"\x1b[1;33mwarning\x1b[0m: unused variable \x1b[38;5;214m`count`\x1b[0m\n" +
	"  \x1b[34m-->\x1b[0m src/engine.go:38:6"

func main() {
	outPath := flag.String("o", "cmd/demo/output.html", "output file path")
	toStdout := flag.Bool("stdout", false, "write to stdout instead of file")
	flag.Parse()

	ctx := context.Background()
	h, err := nuri.New(ctx,
		nuri.WithFS(core.FS()),
		nuri.WithPoolSize(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close(ctx)

	themes := map[string]string{
		"dark":  "github-dark-high-contrast",
		"light": "github-light",
	}

	var sections []Section

	// Section 1: Go
	sections = append(sections, Section{
		ID:          1,
		Title:       "Go — github-dark-high-contrast / github-light",
		Description: "Basic syntax highlighting with theme toggle support.",
		Blocks: []Block{{
			HTML: template.HTML(must(h.CodeToHTML(ctx, goSnippet, nuri.CodeToHTMLOptions{
				Lang:   "go",
				Themes: themes,
			}))),
		}},
	})

	// Section 2: JavaScript
	sections = append(sections, Section{
		ID:          2,
		Title:       "JavaScript — github-dark-high-contrast / github-light",
		Description: "Different language, same theme pair.",
		Blocks: []Block{{
			HTML: template.HTML(must(h.CodeToHTML(ctx, jsSnippet, nuri.CodeToHTMLOptions{
				Lang:   "javascript",
				Themes: themes,
			}))),
		}},
	})

	// Section 3: Multi-theme showcase
	sections = append(sections, Section{
		ID:          3,
		Title:       "Multi-Theme — CSS Variables",
		Description: "Two languages rendered with dual-theme CSS variable output.",
		Blocks: []Block{
			{
				Label: "Go",
				HTML: template.HTML(must(h.CodeToHTML(ctx, goMultiSnippet, nuri.CodeToHTMLOptions{
					Lang:   "go",
					Themes: themes,
				}))),
			},
			{
				Label: "JavaScript",
				HTML: template.HTML(must(h.CodeToHTML(ctx, jsMultiSnippet, nuri.CodeToHTMLOptions{
					Lang:   "javascript",
					Themes: themes,
				}))),
			},
		},
	})

	// Section 4: Decorations
	sections = append(sections, Section{
		ID:          4,
		Title:       "Decorations",
		Description: "Highlighted lines, focused lines, and diff markers.",
		Blocks: []Block{
			{
				Label: "Highlighted Lines",
				HTML: template.HTML(must(h.CodeToHTML(ctx, goHighlightSnippet, nuri.CodeToHTMLOptions{
					Lang:           "go",
					Themes:         themes,
					HighlightLines: []nuri.LineRange{nuri.Range(2, 2), nuri.Range(4, 4)},
				}))),
			},
			{
				Label: "Focused Lines",
				HTML: template.HTML(must(h.CodeToHTML(ctx, jsFocusSnippet, nuri.CodeToHTMLOptions{
					Lang:       "javascript",
					Themes:     themes,
					FocusLines: []nuri.LineRange{nuri.Range(3, 5)},
				}))),
			},
			{
				Label: "Diff Lines",
				HTML: template.HTML(must(h.CodeToHTML(ctx, goDiffSnippet, nuri.CodeToHTMLOptions{
					Lang:          "go",
					Themes:        themes,
					InsertedLines: []nuri.LineRange{nuri.Range(2, 2)},
					DeletedLines:  []nuri.LineRange{nuri.Range(3, 3)},
				}))),
			},
		},
	})

	// Section 5: Meta transformer
	sections = append(sections, Section{
		ID:          5,
		Title:       "Transformer — Meta",
		Description: `Fence meta "{1,3-5}" highlights lines 1, 3, 4, 5.`,
		Blocks: []Block{{
			HTML: template.HTML(must(h.CodeToHTML(ctx, jsMetaSnippet, nuri.CodeToHTMLOptions{
				Lang:         "javascript",
				Themes:       themes,
				Transformers: []nuri.Transformer{transformers.Meta("{1,3-5}")},
			}))),
		}},
	})

	// Section 6: Notation transformer
	sections = append(sections, Section{
		ID:          6,
		Title:       "Transformer — Notation",
		Description: "Magic comments [!code ++], [!code highlight], etc. are stripped; their effects appear as line styles.",
		Blocks: []Block{{
			HTML: template.HTML(must(h.CodeToHTML(ctx, jsNotationSnippet, nuri.CodeToHTMLOptions{
				Lang:         "javascript",
				Themes:       themes,
				Transformers: []nuri.Transformer{transformers.Notation()},
			}))),
		}},
	})

	// Section 7: StyleClassMap
	cm := nuri.NewStyleClassMap()
	goClassHTML := must(h.CodeToHTML(ctx, goClassMapSnippet, nuri.CodeToHTMLOptions{
		Lang:     "go",
		Themes:   themes,
		ClassMap:  cm,
	}))
	jsClassHTML := must(h.CodeToHTML(ctx, jsClassMapSnippet, nuri.CodeToHTMLOptions{
		Lang:     "javascript",
		Themes:   themes,
		ClassMap:  cm,
	}))
	extractedCSS := cm.CSS()

	sections = append(sections, Section{
		ID:          7,
		Title:       "StyleClassMap — CSS Class Output",
		Description: "Inline styles replaced with hashed classes. Extracted CSS shown below.",
		Blocks: []Block{
			{Label: "Go", HTML: template.HTML(goClassHTML)},
			{Label: "JavaScript", HTML: template.HTML(jsClassHTML)},
		},
		ExtraCSS:  template.CSS(extractedCSS),
		ExtraHTML: template.HTML(fmt.Sprintf("<pre class=\"css-dump\"><code>%s</code></pre>", template.HTMLEscapeString(extractedCSS))),
	})

	// Section 8: Custom Configuration (constructor options)
	sections = append(sections, buildCustomConfigSection(ctx, 8))

	// Section 9: ANSI Highlighting
	sections = append(sections, Section{
		ID:          9,
		Title:       "ANSI Highlighting",
		Description: `Terminal output with ANSI escape codes rendered as colored HTML (lang: "ansi").`,
		Blocks: []Block{
			{
				Label: "Test Runner Output",
				HTML: template.HTML(must(h.CodeToHTML(ctx, ansiTestRunnerSnippet, nuri.CodeToHTMLOptions{
					Lang:  "ansi",
					Theme: "github-dark-high-contrast",
				}))),
			},
			{
				Label: "Build Output (with truecolor + 256-color)",
				HTML: template.HTML(must(h.CodeToHTML(ctx, ansiBuildSnippet, nuri.CodeToHTMLOptions{
					Lang:  "ansi",
					Theme: "github-dark-high-contrast",
				}))),
			},
		},
	})

	data := struct {
		Sections  []Section
		Generated string
	}{
		Sections:  sections,
		Generated: time.Now().Format(time.RFC3339),
	}

	tmpl, err := template.ParseFiles("cmd/demo/template.html")
	if err != nil {
		log.Fatal(err)
	}

	if *toStdout {
		if err := tmpl.Execute(os.Stdout, data); err != nil {
			log.Fatal(err)
		}
		return
	}

	f, err := os.Create(*outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, data); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote %s\n", *outPath)
}

func buildCustomConfigSection(ctx context.Context, id int) Section {
	// Load an existing bundled theme and re-register it under a custom name.
	bundleFS := core.FS()
	draculaData, err := fs.ReadFile(bundleFS, "themes/dracula.json")
	if err != nil {
		log.Fatalf("read dracula theme: %v", err)
	}

	// A minimal toy DSL grammar.
	dslGrammar := []byte(`{
		"scopeName": "source.mydsl",
		"patterns": [
			{"match": "\\b(let|set|if|end|print)\\b", "name": "keyword.control.mydsl"},
			{"match": "\\b\\d+\\b", "name": "constant.numeric.mydsl"},
			{"begin": "\"", "end": "\"", "name": "string.quoted.double.mydsl"}
		]
	}`)

	h, err := nuri.New(ctx,
		nuri.WithFS(bundleFS),
		nuri.WithPoolSize(1),
		nuri.WithTheme("my-custom-theme", draculaData),
		nuri.WithGrammar("my-dsl", dslGrammar),
		nuri.WithAlias("mydsl", "my-dsl"),
		nuri.WithDefaults(nuri.CodeToHTMLOptions{
			Theme:    "my-custom-theme",
			PreClass: "demo-defaults",
		}),
	)
	if err != nil {
		log.Fatalf("New (custom config): %v", err)
	}
	defer h.Close(ctx)

	// Block 1: Custom grammar via alias — proves WithGrammar + WithAlias.
	dslHTML := must(h.CodeToHTML(ctx, dslSnippet, nuri.CodeToHTMLOptions{
		Lang: "mydsl",
	}))

	// Block 2: Go with no Theme — proves WithDefaults supplies "my-custom-theme".
	goDefaultsHTML := must(h.CodeToHTML(ctx, goDefaultsSnippet, nuri.CodeToHTMLOptions{
		Lang: "go",
	}))

	// Block 3: Go with explicit Theme override — proves per-call wins.
	goOverrideHTML := must(h.CodeToHTML(ctx, goOverrideSnippet, nuri.CodeToHTMLOptions{
		Lang:  "go",
		Theme: "github-dark-high-contrast",
	}))

	return Section{
		ID:          id,
		Title:       "Custom Configuration — Constructor Options",
		Description: "WithTheme, WithGrammar, WithAlias, and WithDefaults exercised in a single New() call.",
		Blocks: []Block{
			{Label: "Custom Grammar (via WithGrammar + WithAlias)", HTML: template.HTML(dslHTML)},
			{Label: "Go — Theme from WithDefaults (dracula)", HTML: template.HTML(goDefaultsHTML)},
			{Label: "Go — Theme overridden per-call (github-dark-high-contrast)", HTML: template.HTML(goOverrideHTML)},
		},
	}
}
