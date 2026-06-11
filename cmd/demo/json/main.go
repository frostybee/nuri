package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/core"
)

const goSnippet = `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`

func main() {
	ctx := context.Background()
	h, err := nuri.New(ctx, nuri.WithFS(core.FS()), nuri.WithPoolSize(1))
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close(ctx)

	fmt.Println("=== Compact JSON ===")
	data, err := h.CodeToJSON(ctx, goSnippet, nuri.CodeToJSONOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)
	fmt.Println()

	fmt.Println("\n=== Indented JSON ===")
	data, err = h.CodeToJSON(ctx, goSnippet, nuri.CodeToJSONOptions{
		Lang: "go", Theme: "github-dark", Indent: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)
	fmt.Println()

	fmt.Println("\n=== Multi-theme JSON ===")
	data, err = h.CodeToJSON(ctx, goSnippet, nuri.CodeToJSONOptions{
		Themes: map[string]string{"dark": "github-dark", "light": "github-light"},
		Indent: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(data)
	fmt.Println()
}
