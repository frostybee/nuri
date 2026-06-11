package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/core"
)

const goSnippet = `package main

import "fmt"

func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(Greet("world"))
}
`

const jsSnippet = `async function fetchUsers(endpoint) {
  const response = await fetch(endpoint);
  const data = await response.json();
  return data.users.filter(u => u.active);
}
`

func main() {
	outDir := flag.String("o", ".", "output directory for SVG files")
	flag.Parse()

	ctx := context.Background()
	h, err := nuri.New(ctx, nuri.WithFS(core.FS()), nuri.WithPoolSize(1))
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close(ctx)

	snippets := []struct {
		name  string
		lang  string
		theme string
		code  string
	}{
		{"go-github-dark", "go", "github-dark", goSnippet},
		{"js-dracula", "javascript", "dracula", jsSnippet},
		{"go-nord", "go", "nord", goSnippet},
	}

	for _, s := range snippets {
		svg, err := h.CodeToSVG(ctx, s.code, nuri.CodeToSVGOptions{
			Lang:  s.lang,
			Theme: s.theme,
		})
		if err != nil {
			log.Fatalf("%s: %v", s.name, err)
		}

		path := fmt.Sprintf("%s/%s.svg", *outDir, s.name)
		if err := os.WriteFile(path, []byte(svg), 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Wrote %s (%d bytes)\n", path, len(svg))
	}
}
