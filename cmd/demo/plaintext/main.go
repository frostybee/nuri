package main

import (
	"context"
	"fmt"
	"log"

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

	text, err := h.CodeToPlainText(ctx, goSnippet, nuri.CodeToPlainTextOptions{
		Lang: "go", Theme: "github-dark",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== PlainText Output ===")
	fmt.Println(text)
	fmt.Println("========================")
	fmt.Printf("Length: %d bytes\n", len(text))
}
