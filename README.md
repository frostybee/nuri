# Nuri

A pure Go port of [Shiki](https://shiki.style), the TextMate grammar-based syntax highlighter used by VS Code. Full TextMate grammar support: 245+ languages, 65+ VS Code themes, and hundreds of hierarchical token scopes. No CGO required.

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
