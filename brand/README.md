# Nuri brand assets

Master copies of the Nuri logo and favicon. This folder is the source of truth.
Files served by the documentation site are copies, because Sarde only serves
static assets from `docs/public/` and its public directory is a hardcoded
constant with no configuration hook.

Copying is manual. After editing a master here, copy it to every destination
listed below.

## Social card masters

Sarde's `social_cards` plugin brands generated Open Graph images with a corner mark and a
large low-opacity watermark. It composites raster images only, so the SVG masters are
rasterized offline with oksvg, which cannot handle several features of `nuri-logo.svg`.
The `-src` files are flattened, oksvg-adapted copies of the logo that compensate for each:

- clip paths are ignored by oksvg, so the top-half/bottom-half clipping is resolved by
  hand: grey lines above the brush stroke, coloured lines below it, clipped-away
  duplicates omitted
- `rx` on `rect` is dropped, so the rounded tile and border are explicit paths and the
  pill code lines are round-cap strokes
- stroke widths do not scale, so the watermark file pre-multiplies every coordinate and
  stroke width by 2 for a 1024 px raster
- percent-based gradients are converted to explicit `userSpaceOnUse` coordinates
- element opacities are pre-blended into flat colours against the `#1a1f35` tile (grey
  lines `#3a4052`, angle brackets `#4e78a4`, fude detail tints)
- the fude brush `translate + rotate` becomes an explicit `matrix`; in the watermark the
  matrix also carries the doubled scale, which is safe because the brush interior uses
  fills only

Files:

- `social-card-mark-src.svg`: flattened `nuri-logo.svg`, rasterized at 512 px
- `social-card-mark.png`: the corner mark drawn on social cards
- `social-card-watermark-2x.svg`: the bare logo (no tile or border) at 2x, rasterized at 1024 px
- `social-card-watermark.png`: the card watermark. On the navy card background the grey
  lines recede, leaving the highlighted lines, brush stroke, brackets, and fude

Edit `nuri-logo.svg` first, mirror the change into both `-src` files, re-rasterize, and
copy the PNGs to the destinations below.

## Files and their destinations

| Master | Copy to | Used by |
|--------|---------|---------|
| `nuri-logo.svg` | no copy needed | `README.md` header image, referenced here directly |
| `nuri-logo.svg` | `docs/public/images/nuri.svg` | Docs site homepage hero, via `homepage.hero.image` in `docs/sarde.yaml` |
| `nuri-favicon.svg` | `docs/public/favicon.svg` | Docs site favicon, via `site.favicon` in `docs/sarde.yaml` |
| `nuri-favicon.svg` | `docs/public/favicon.svg` | Docs site header logo, via `site.logo` in `docs/sarde.yaml`. Same copy as the favicon, no second file. |
| `social-card-mark.png` | `docs/public/images/nuri-mark.png` | Social card corner mark, via `social_cards.logo` in `docs/sarde.yaml` |
| `social-card-watermark.png` | `docs/public/images/nuri-watermark.png` | Social card watermark, via `social_cards.watermark_image` in `docs/sarde.yaml` |

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

## Why the header logo is the favicon art

The docs header renders its logo at Sarde's `logo-height` token, 1.75rem or 28px
by default. That is an 18x reduction of the 512-unit canvas, which puts the logo's
code lines at 0.55px, its brush stroke at 0.44px, its angle brackets at 0.49px,
and its fude handle at 0.98px. Sizing up is not a fix: the logo needs 160px and
the header is 64px tall.

The favicon art at 28px puts its bars at 2.4px and its stroke at 1.4px, so it
holds. `site.logo` therefore points at `/favicon.svg`, the same copy the favicon
uses. A dedicated header asset is only worth authoring if the header should carry
more of the logo's character than four bars and a stroke, and that means new
artwork tuned for roughly 24 to 40 pixels, not a rescale of either existing file.

## Raster exports

`.gitignore` ignores `*.png` and `*.svg` repository-wide. Negations for
`brand/**/*.png` and `brand/**/*.svg` keep this folder tracked. Verify with
`git check-ignore -v` before assuming a new binary asset was committed.
