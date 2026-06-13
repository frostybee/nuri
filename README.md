# Nuri

A pure Go port of [Shiki](https://shiki.style), the TextMate grammar-based syntax highlighter used by VS Code. Full TextMate grammar support: 257 languages, 65+ VS Code themes, and hundreds of hierarchical token scopes. No CGO required.

> 227 of 234 tested grammars (97%) produce output byte-identical to Shiki, verified against [vscode-textmate](https://github.com/microsoft/vscode-textmate); the core bundle's 32 fidelity-tested grammars are at 100%.

## Development

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/frostybee/nuri.git

# Or initialize submodules after cloning
git submodule update --init

# Sync grammars and themes from the submodule into bundle/
go run ./tools/devtool sync

# Generate fidelity test fixtures (requires Node.js 22+)
go run ./tools/devtool generate

# Run tests
go test ./...
```

See [tools/devtool/README.md](tools/devtool/README.md) for full details on managing grammars, themes, and test fixtures.
