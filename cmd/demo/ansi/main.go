package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/frostybee/nuri"
	"github.com/frostybee/nuri/bundle/core"
	"golang.org/x/term"
)

var goSnippet = `package main

import "fmt"

// Greet returns a greeting for the given name.
func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	msg := Greet("world")
	fmt.Println(msg) // prints: Hello, world!
}
`

var jsSnippet = `import { readFile } from "fs/promises";

// Parse a JSON config file.
const loadConfig = async (path) => {
  const data = await readFile(path, "utf-8");
  const config = JSON.parse(data);
  return { ...config, loadedAt: Date.now() };
};

export default loadConfig;
`

var pySnippet = `from dataclasses import dataclass
from typing import Optional

@dataclass
class User:
    """A user with an optional bio."""
    name: str
    age: int
    bio: Optional[str] = None

    def greeting(self) -> str:
        return f"Hi, I'm {self.name} ({self.age})"

users = [User("Alice", 30), User("Bob", 25, "dev")]
names = [u.name for u in users if u.age > 20]
`

type snippet struct {
	lang string
	code string
}

var snippets = []snippet{
	{"go", goSnippet},
	{"javascript", jsSnippet},
	{"python", pySnippet},
}

type depthInfo struct {
	name  string
	depth nuri.ColorDepth
}

var depths = []depthInfo{
	{"Truecolor (24-bit)", nuri.ColorDepthTruecolor},
	{"256-color", nuri.ColorDepth256},
	{"16-color", nuri.ColorDepth16},
	{"8-color", nuri.ColorDepth8},
}

var interactive bool

func main() {
	ctx := context.Background()
	h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "nuri.New: %v\n", err)
		os.Exit(1)
	}
	defer h.Close(ctx)

	fd := int(os.Stdin.Fd())
	interactive = term.IsTerminal(fd)
	if interactive {
		oldState, err := term.MakeRaw(fd)
		if err == nil {
			defer term.Restore(fd, oldState)
		} else {
			interactive = false
		}
	}

	printHeader()

	// --- Section 1: Multiple languages ---
	printSection("Languages", "github-dark, truecolor")
	for _, s := range snippets {
		printLabel(s.lang)
		out, err := h.CodeToANSI(ctx, s.code, nuri.CodeToANSIOptions{
			Lang:  s.lang,
			Theme: "github-dark",
		})
		if err != nil {
			writef("  %serror:%s %v\n", red, reset, err)
			continue
		}
		writef("%s\n", out)
	}
	if !waitForKey() {
		return
	}

	// --- Section 2: Color depth comparison ---
	printSection("Color Depths", "Go, github-dark")
	for _, d := range depths {
		printLabel(d.name)
		out, err := h.CodeToANSI(ctx, goSnippet, nuri.CodeToANSIOptions{
			Lang:       "go",
			Theme:      "github-dark",
			ColorDepth: d.depth,
		})
		if err != nil {
			writef("  %serror:%s %v\n", red, reset, err)
			continue
		}
		writef("%s\n", out)
	}
	if !waitForKey() {
		return
	}

	// --- Section 3: Theme comparison ---
	printSection("Themes", "Go, truecolor")
	themes := []string{"github-dark", "github-light", "dracula", "nord", "catppuccin-mocha", "one-dark-pro"}
	for _, theme := range themes {
		printLabel(theme)
		out, err := h.CodeToANSI(ctx, goSnippet, nuri.CodeToANSIOptions{
			Lang:  "go",
			Theme: theme,
		})
		if err != nil {
			writef("  %serror:%s %v\n", red, reset, err)
			continue
		}
		writef("%s\n", out)
	}
	if !waitForKey() {
		return
	}

	// --- Section 4: Language detection ---
	printSection("Language Detection", "auto-detect from filename")
	detectFiles := []struct {
		filename string
		snippet  string
	}{
		{"main.go", goSnippet},
		{"app.py", pySnippet},
		{"index.js", jsSnippet},
		{"Makefile", "all: build\n\nbuild:\n\tgo build ./...\n"},
		{"style.css", "body {\n  margin: 0;\n  font-family: sans-serif;\n}\n"},
	}
	for _, df := range detectFiles {
		lang, ok := h.DetectLanguage(df.filename)
		if !ok {
			writef("  %s%s%s %s→ not detected%s\n", bold, df.filename, reset, dim, reset)
			continue
		}
		printLabel(fmt.Sprintf("%s → %s", df.filename, lang))
		out, err := h.CodeToANSI(ctx, df.snippet, nuri.CodeToANSIOptions{
			Lang:  lang,
			Theme: "github-dark",
		})
		if err != nil {
			writef("  %serror:%s %v\n", red, reset, err)
			continue
		}
		writef("%s\n", out)
	}

	writef("\n")
}

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	italic  = "\033[3m"
	cyan    = "\033[36m"
	magenta = "\033[35m"
	blue    = "\033[34m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	white   = "\033[97m"
)

// writef writes formatted text, converting \n to \r\n in raw terminal mode.
func writef(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	if interactive {
		s = strings.ReplaceAll(s, "\n", "\r\n")
	}
	fmt.Print(s)
}

func printHeader() {
	writef("\n")
	writef("  %s%sNuri%s  %sANSI Output Demo%s\n", bold, cyan, reset, dim, reset)
	writef("  %s%s%s\n", dim, strings.Repeat("─", 40), reset)
	writef("  %sTextMate syntax highlighting for the terminal%s\n", dim, reset)
	writef("\n")
}

func printSection(title, subtitle string) {
	writef("\n  %s%s%s %s  %s%s%s\n", bold, white, title, reset, dim, subtitle, reset)
	writef("  %s%s%s\n\n", dim, strings.Repeat("─", 50), reset)
}

func printLabel(label string) {
	writef("  %s%s▸%s %s%s%s\n", dim, magenta, reset, bold, label, reset)
}

func waitForKey() bool {
	if !interactive {
		return true
	}
	writef("  %s%s[space]%s%s next  %s·%s  %s[q]%s%s quit%s", dim, white, reset, dim, dim, dim, white, reset, dim, reset)
	var buf [1]byte
	os.Stdin.Read(buf[:])
	writef("\r  %s\r", strings.Repeat(" ", 40))
	switch buf[0] {
	case 'q', 'Q', 0x1b, 0x03:
		return false
	}
	return true
}
