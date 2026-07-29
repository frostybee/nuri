---
title: Contrast Correction
description: "Automatic WCAG 2.1 foreground color adjustment for syntax tokens."
sidebar:
  order: 9
---

Nuri adjusts syntax token foreground colors that fail a WCAG 2.1 contrast check against the theme background. This is enabled by default.

## How it works

At theme load time, each token foreground color is checked against the theme's editor background. Colors that fail the minimum contrast ratio are shifted toward white or black using a binary search. The algorithm tries both directions and picks the one with the smallest color shift, preserving as much of the original hue as possible.

Adjusted themes are cached per name. Each theme is processed exactly once, with zero per-token cost during highlighting.

## Configure the ratio

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithMinContrast(4.5),
)
```

| Ratio | WCAG level | Description |
|---|---|---|
| `4.5` | AA | Minimum for normal text |
| `5.5` | AA enhanced | Default |
| `7.0` | AAA | Enhanced contrast |
| `0` | Disabled | Raw theme colors preserved |

The default is 5.5, which sits between WCAG AA (4.5:1) and AAA (7.0:1). Set to 0 to disable contrast correction entirely and preserve the raw theme colors.

## The WCAG contrast ratio

The contrast ratio is computed per WCAG 2.1:

```
ratio = (L1 + 0.05) / (L2 + 0.05)
```

`L1` and `L2` are the relative luminances of the lighter and darker colors. Relative luminance uses sRGB linearization (gamma 2.4) with the standard coefficients:

```
L = 0.2126 * R + 0.7152 * G + 0.0722 * B
```

A ratio of 1:1 means no contrast (identical colors). A ratio of 21:1 is the maximum (black on white).

## The adjustment algorithm

For a foreground color that fails the ratio:

1. Shift toward white: each RGB channel becomes `c + (1 - c) * step`
2. Shift toward black: each RGB channel becomes `c * (1 - step)`
3. A binary search (32 iterations per direction) finds the smallest `step` (0.0 = no change, 1.0 = pure white or black) that achieves the required ratio
4. The direction with the smaller step wins, producing the minimal color shift
5. If neither direction achieves the ratio, the color falls back to `#ffffff` or `#000000` depending on the background luminance

The result is a color that meets the contrast requirement while staying as close to the original as possible.

## Interaction with `h.GetThemeColors()`

`h.GetThemeColors()` returns contrast-adjusted colors. It loads themes through the same internal path, so the `Background`, `Foreground`, and all `Colors` values reflect the adjusted theme.

## Disable contrast correction

```go
h, err := nuri.New(ctx,
	nuri.WithFS(core.FS()),
	nuri.WithMinContrast(0),
)
```

With a ratio of 0 (or any negative value), themes are returned unmodified from the registry. No color adjustment occurs.

## Next steps

- [Performance & Concurrency](/docs/guides/performance-and-concurrency) for pool sizing, timeouts, and compilation cache
- [Using Themes](/docs/guides/using-themes) for theme selection and available themes
