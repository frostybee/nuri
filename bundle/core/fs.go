package core

import (
	"embed"
	"io/fs"

	"github.com/frostybee/nuri/internal/assetfs"
)

//go:embed grammars themes
var embedFS embed.FS

// FS returns the bundle filesystem containing grammar and theme JSON files.
// Grammars are under "grammars/" and themes under "themes/". Assets are
// stored gzip compressed as .json.gz and presented as virtual .json files
// through a transparent decompressing wrapper.
func FS() fs.FS { return assetfs.New(embedFS) }
