---
title: Bundled Languages & Themes
description: "Lookup tables for all bundled languages, aliases, and themes."
sidebar:
  order: 7
---

Nuri ships two bundles. `bundle/core` includes 38 grammars (32 primary languages + 6 injection helpers) and 65 themes. `bundle/full` includes all 257 grammars and the same 65 themes.

Both packages export a single function:

```go
func FS() fs.FS
```

```go
import "github.com/frostybee/nuri/bundle/core"   // 38 grammars, 65 themes
import "github.com/frostybee/nuri/bundle/full"   // 257 grammars, 65 themes

h, err := nuri.New(ctx, nuri.WithFS(core.FS()))
```

## Core languages

The 32 primary languages in `bundle/core`. Use the canonical name in the `Lang` field of all options structs.

| Language | Canonical name | Aliases | Extensions |
|---|---|---|---|
| Batch | `bat` | `batch` | `.bat`, `.cmd` |
| C | `c` | | `.c`, `.h` |
| C++ | `cpp` | `c++` | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.hh` |
| C# | `csharp` | `c#`, `cs` | `.cs` |
| CSS | `css` | | `.css` |
| Docker | `docker` | `dockerfile` | `Dockerfile` |
| Go | `go` | | `.go` |
| GraphQL | `graphql` | `gql` | |
| HTML | `html` | | `.html`, `.htm` |
| Java | `java` | | `.java` |
| JavaScript | `javascript` | `js`, `cjs`, `mjs` | `.js`, `.mjs`, `.cjs` |
| JSON | `json` | | `.json` |
| JSONC | `jsonc` | | `.jsonc` |
| JSX | `jsx` | | `.jsx` |
| Kotlin | `kotlin` | `kt`, `kts` | `.kt`, `.kts` |
| Lua | `lua` | | `.lua` |
| Markdown | `markdown` | `md` | `.md` |
| PHP | `php` | | `.php` |
| Python | `python` | `py` | `.py`, `.pyi`, `.pyw` |
| Ruby | `ruby` | `rb` | `.rb`, `Gemfile`, `Rakefile` |
| Rust | `rust` | `rs` | `.rs` |
| SCSS | `scss` | | `.scss` |
| Shell | `shellscript` | `bash`, `sh`, `shell`, `zsh` | `.sh`, `.bash`, `.zsh` |
| SQL | `sql` | | `.sql` |
| Svelte | `svelte` | | `.svelte` |
| Swift | `swift` | | `.swift` |
| TOML | `toml` | | `.toml` |
| TSX | `tsx` | | `.tsx` |
| TypeScript | `typescript` | `ts`, `cts`, `mts` | `.ts`, `.mts`, `.cts` |
| Vue | `vue` | | `.vue` |
| XML | `xml` | | `.xml`, `.xsl`, `.svg` |
| YAML | `yaml` | `yml` | `.yaml`, `.yml` |

### Injection helpers

The core bundle also includes 6 injection helper grammars. These are loaded automatically when their parent grammar requires them. Do not use them directly in `Lang` fields.

`cpp-macro`, `html-derivative`, `vue-directives`, `vue-html`, `vue-interpolations`, `vue-sfc-style-variable-injection`

## Full bundle additions

The full bundle includes all 32 core languages plus 219 additional grammars. The canonical name is what to pass in the `Lang` field.

`abap`, `actionscript-3`, `ada`, `ahk`, `ahk2`, `angular-expression`, `angular-html`, `angular-inline-style`, `angular-let-declaration`, `angular-template`, `angular-template-blocks`, `angular-ts`, `apache`, `apex`, `apl`, `applescript`, `ara`, `asciidoc`, `asm`, `astro`, `awk`, `ballerina`, `beancount`, `berry`, `bibtex`, `bicep`, `bird2`, `blade`, `bsl`, `c3`, `cadence`, `cairo`, `chapel`, `clarity`, `clojure`, `cmake`, `cobol`, `codeowners`, `codeql`, `coffee`, `common-lisp`, `coq`, `crystal`, `csv`, `cue`, `cypher`, `d`, `dart`, `dax`, `desktop`, `diff`, `dotenv`, `dream-maker`, `edge`, `elixir`, `elm`, `emacs-lisp`, `erb`, `erlang`, `es-tag-css`, `es-tag-glsl`, `es-tag-html`, `es-tag-sql`, `es-tag-xml`, `fennel`, `fish`, `fluent`, `fortran-fixed-form`, `fortran-free-form`, `fsharp`, `gdresource`, `gdscript`, `gdshader`, `genie`, `gherkin`, `git-commit`, `git-rebase`, `gleam`, `glimmer-js`, `glimmer-ts`, `glsl`, `gn`, `gnuplot`, `groovy`, `hack`, `haml`, `handlebars`, `haskell`, `haxe`, `hcl`, `hjson`, `hlsl`, `http`, `hurl`, `hxml`, `hy`, `imba`, `ini`, `jinja`, `jinja-html`, `jison`, `json5`, `jsonl`, `jsonnet`, `jssm`, `julia`, `just`, `kdl`, `kusto`, `latex`, `lean`, `less`, `liquid`, `llvm`, `log`, `logo`, `luau`, `make`, `markdown-nix`, `markdown-vue`, `marko`, `matlab`, `mdc`, `mdx`, `mermaid`, `mipsasm`, `mojo`, `moonbit`, `move`, `narrat`, `nextflow`, `nextflow-groovy`, `nginx`, `nim`, `nix`, `nushell`, `objective-c`, `objective-cpp`, `ocaml`, `odin`, `openscad`, `org`, `pascal`, `perl`, `pkl`, `plsql`, `po`, `polar`, `postcss`, `powerquery`, `powershell`, `prisma`, `prolog`, `proto`, `pug`, `puppet`, `purescript`, `qml`, `qmldir`, `qss`, `r`, `racket`, `raku`, `razor`, `reg`, `regexp`, `rel`, `riscv`, `ron`, `rosmsg`, `rst`, `sas`, `sass`, `scala`, `scheme`, `sdbl`, `shaderlab`, `shellsession`, `smalltalk`, `solidity`, `soy`, `sparql`, `splunk`, `ssh-config`, `stata`, `stylus`, `surrealql`, `systemd`, `system-verilog`, `talonscript`, `tasl`, `tcl`, `templ`, `terraform`, `tex`, `ts-tags`, `tsv`, `turtle`, `twig`, `typespec`, `typst`, `v`, `vala`, `vb`, `verilog`, `vhdl`, `viml`, `vue-vine`, `vyper`, `wasm`, `wenyan`, `wgsl`, `wikitext`, `wit`, `wolfram`, `xsl`, `zenscript`, `zig`

## Themes

All 65 themes are included in both `bundle/core` and `bundle/full`.

### Dark themes

| Theme name |
|---|
| `andromeeda` |
| `aurora-x` |
| `ayu-dark` |
| `ayu-mirage` |
| `catppuccin-frappe` |
| `catppuccin-macchiato` |
| `catppuccin-mocha` |
| `dark-plus` |
| `dracula` |
| `dracula-soft` |
| `everforest-dark` |
| `github-dark` |
| `github-dark-default` |
| `github-dark-dimmed` |
| `github-dark-high-contrast` |
| `gruvbox-dark-hard` |
| `gruvbox-dark-medium` |
| `gruvbox-dark-soft` |
| `horizon` |
| `houston` |
| `kanagawa-dragon` |
| `kanagawa-wave` |
| `laserwave` |
| `material-theme` |
| `material-theme-darker` |
| `material-theme-ocean` |
| `material-theme-palenight` |
| `min-dark` |
| `monokai` |
| `night-owl` |
| `nord` |
| `one-dark-pro` |
| `plastic` |
| `poimandres` |
| `red` |
| `rose-pine` |
| `rose-pine-moon` |
| `slack-dark` |
| `solarized-dark` |
| `synthwave-84` |
| `tokyo-night` |
| `vesper` |
| `vitesse-black` |
| `vitesse-dark` |

### Light themes

| Theme name |
|---|
| `ayu-light` |
| `catppuccin-latte` |
| `everforest-light` |
| `github-light` |
| `github-light-default` |
| `github-light-high-contrast` |
| `gruvbox-light-hard` |
| `gruvbox-light-medium` |
| `gruvbox-light-soft` |
| `horizon-bright` |
| `kanagawa-lotus` |
| `light-plus` |
| `material-theme-lighter` |
| `min-light` |
| `night-owl-light` |
| `one-light` |
| `rose-pine-dawn` |
| `slack-ochin` |
| `snazzy-light` |
| `solarized-light` |
| `vitesse-light` |

## Language detection

`h.DetectLanguage(filename)` resolves filenames and extensions to canonical language names. `h.DetectLanguageByContent(firstLine)` resolves by shebang. Both return the canonical name from the tables above. See [Highlighter API](/docs/reference/highlighter-api) for method details.

Register additional aliases and extensions with `nuri.WithAlias()`, `nuri.WithExtension()`, `h.RegisterAlias()`, or `h.RegisterExtension()`.
