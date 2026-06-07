package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
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
