package full

import (
	"embed"
	"io/fs"
)

//go:embed grammars themes
var embedFS embed.FS

// FS returns the embedded filesystem containing grammar and theme JSON files.
// Grammars are under "grammars/" and themes under "themes/".
func FS() fs.FS { return embedFS }
