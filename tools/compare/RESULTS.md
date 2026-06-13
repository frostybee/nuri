# Nuri vs Shiki vs Chroma — Benchmark Results

*Generated: 2026-06-13T04:11:48Z | Machine: rsleiman (amd64) | Warm iterations: 10 | Theme: github-dark*
*Versions: chroma v2.18.0, go go1.26.0, node v24.12.0, nuri commit:4a4580b*

## Inputs

| Input | Language | Bytes | Lines |
|---|---|---:|---:|
| Go | go | 117 | 11 |
| HTML | html | 304 | 16 |
| JavaScript | javascript | 309 | 13 |
| Markdown | markdown | 135 | 12 |
| TypeScript | typescript | 203 | 11 |

## Speed (warm median, ms/op)

| Input | Nuri | Nuri (no-interrupt) | Shiki | Chroma |
|---|---:|---:|---:|---:|
| Go | 5.00 (23 KB/s) | 1.26 (91 KB/s) | 0.98 (117 KB/s) | 0.00 |
| HTML | 18.62 (16 KB/s) | 4.00 (74 KB/s) | 1.23 (241 KB/s) | 0.00 |
| JavaScript | 28.57 (11 KB/s) | 7.03 (43 KB/s) | 1.86 (162 KB/s) | 0.00 |
| Markdown | 5.00 (26 KB/s) | 1.57 (84 KB/s) | 0.51 (260 KB/s) | 0.00 |
| TypeScript | 15.02 (13 KB/s) | 4.00 (50 KB/s) | 0.96 (207 KB/s) | 0.00 |

## Cold start (first call, ms)

| Input | Nuri | Shiki | Chroma |
|---|---:|---:|---:|
| Go | 76.60 | 40.27 | 1.00 |
| HTML | 354.34 | 48.15 | 2.55 |
| JavaScript | 556.85 | 74.42 | 1.01 |
| Markdown | 144.86 | 20.62 | 1.00 |
| TypeScript | 676.82 | 87.90 | 1.00 |

## Allocations (per warm call)

| Input | Nuri | Nuri (no-interrupt) | Shiki | Chroma |
|---|---:|---:|---:|---:|
| Go | 1,942 / 205KB | 1,728 / 192KB | 312KB | 1,247 / 73KB |
| HTML | 2,788 / 286KB | 2,281 / 249KB | 662KB | 2,554 / 143KB |
| JavaScript | 3,626 / 566KB | 3,130 / 536KB | 875KB | 2,686 / 169KB |
| Markdown | 2,151 / 266KB | 1,992 / 257KB | 224KB | 1,576 / 90KB |
| TypeScript | 3,121 / 528KB | 2,845 / 512KB | 369KB | 1,733 / 108KB |

## Fidelity

| Input | Engine | Tokens | Distinct scopes/types |
|---|---|---:|---:|
| Go | nuri | 56 | 24 |
|  | shiki | 39 | 24 |
|  | chroma | 55 | 10 |
| HTML | nuri | 110 | 48 |
|  | shiki | 78 | 48 |
|  | chroma | 105 | 11 |
| JavaScript | nuri | 134 | 37 |
|  | shiki | 67 | 37 |
|  | chroma | 120 | 9 |
| Markdown | nuri | 42 | 30 |
|  | shiki | 24 | 30 |
|  | chroma | 31 | 11 |
| TypeScript | nuri | 90 | 51 |
|  | shiki | 63 | 51 |
|  | chroma | 79 | 11 |

*Nuri and Shiki use full TextMate grammars (same Oniguruma engine); Chroma uses Pygments-model lexers with ~80 token types. The fidelity gap is by design, not a bug. Per-engine token dumps are saved alongside the snapshot for detailed diffing.*
