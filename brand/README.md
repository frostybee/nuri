# Nuri brand assets

Master copies of the Nuri logo and favicon. This folder is the source of truth.
Files served by the documentation site are copies, because Sarde only serves
static assets from `docs/public/` and its public directory is a hardcoded
constant with no configuration hook.

Copying is manual. After editing a master here, copy it to every destination
listed below.

## Files and their destinations

| Master | Copy to | Used by |
|--------|---------|---------|
| `nuri-logo.svg` | no copy needed | `README.md` header image, referenced here directly |
| `nuri-logo.svg` | `docs/public/images/nuri.svg` | Docs site homepage hero, via `homepage.hero.image` in `docs/sarde.yaml` |
| `nuri-favicon.svg` | `docs/public/favicon.svg` | Docs site favicon, via `site.favicon` in `docs/sarde.yaml` |

`docs/public/images/hero-light.svg` and `hero-dark.svg` are not mastered here.
They are unmodified Sarde scaffold placeholders, byte-identical to the ones every
new Sarde site ships with, and nothing on the site references them. Move them
into this folder only if they are replaced with authored Nuri artwork.

## Palette

Taken from `nuri-logo.svg`. The syntax colours are GitHub Dark's, which is
deliberate: the logo depicts Nuri's own output, so its code lines use the palette
a reader sees when highlighting with `github-dark`.

| Colour | Hex | Role |
|--------|-----|------|
| Panel | `#1a1f35` | Logo tile background |
| Panel border | `#2a3050` | Tile inner stroke |
| Unhighlighted | `#8b949e` | Grey code lines above the brush stroke |
| Red | `#ff7b72` | Brush gradient start, code lines |
| Orange | `#ffa657` | Brush gradient, code lines |
| Yellow | `#e3b341` | Brush gradient, code lines |
| Green | `#7ee787` | Brush gradient, code lines |
| Blue | `#79c0ff` | Brush gradient, code lines, angle brackets |
| Purple | `#d2a8ff` | Brush gradient end, code lines |
| Brush handle | `#5a3d1a` | Fude handle |
| Ferrule | `#d4af5a` | Fude ferrule and ring accents |
| Bristles | `#f0e8dc` | Fude tip |

## Why the favicon is not the logo

The logo is a 512-unit canvas built for display at 160 pixels and up. Scaling it
to a 16 pixel favicon divides every dimension by 32, which puts most of its
detail below one pixel:

| Element | In the logo | At 16px |
|---------|-------------|---------|
| Code lines | `height="10"` | 0.31px |
| Brush stroke | `stroke-width="8"` | 0.25px |
| Angle brackets | `stroke-width="9"` | 0.28px |
| Fude handle | `width="18"` | 0.56px |
| Tile inset | `x="16"` | 0.5px of dead padding per side |

The favicon therefore drops the fude brush, the angle brackets, and 22 of the 26
code lines. What remains is the one idea that survives the scale: two grey bars
above a coloured brush stroke, two coloured bars below it. That is the whole
concept of the logo, which is unhighlighted source becoming highlighted source.

Current favicon values:

| Part | Value |
|------|-------|
| Tile | `#1a1f35`, corner radius `96` (up from the logo's `72`) |
| Bars above | `#8b949e` at 45% opacity, `height="44"` |
| Bars below | `#7ee787` and `#79c0ff`, `height="44"` |
| Brush stroke | `#ff7b72` to `#d2a8ff`, `stroke-width="26"` |

The bars are 4.4x the logo's line height and the stroke is 3.25x its width, so
both still read at 16px. The corner radius grew because a tighter radius reads as
a square once the tile is only 16 pixels across.

## Raster exports

`.gitignore` ignores `*.png` and `*.svg` repository-wide. Negations for
`brand/**/*.png` and `brand/**/*.svg` keep this folder tracked. Verify with
`git check-ignore -v` before assuming a new binary asset was committed.
