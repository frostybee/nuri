# Nuri vs Shiki vs Chroma — Benchmark Results

*Generated: 2026-06-11T01:42:30Z | Machine: rsleiman (amd64) | Warm iterations: 50 | Theme: github-dark*
*Versions: chroma v2.18.0, go go1.26.0, node v24.12.0, nuri commit:88b262d*

## Speed (warm median, ms/op)

| Input | Nuri | Nuri (no-interrupt) | Shiki | Chroma |
|---|---:|---:|---:|---:|
| Go | 5.51 | 2.00 | 0.56 | 0.00 |
| HTML | 20.02 | 4.52 | 0.75 | 0.00 |
| JavaScript | 29.21 | 7.00 | 1.37 | 0.00 |
| Markdown | 3.00 | 1.00 | 0.35 | 0.00 |
| TypeScript | 15.19 | 4.00 | 0.76 | 0.00 |

## Cold start (first call, ms)

| Input | Nuri | Shiki | Chroma |
|---|---:|---:|---:|
| Go | 79.05 | 38.71 | 1.00 |
| HTML | 357.44 | 47.55 | 2.00 |
| JavaScript | 558.27 | 74.18 | 1.00 |
| Markdown | 142.45 | 20.25 | 1.00 |
| TypeScript | 688.77 | 84.23 | 0.00 |

## Allocations (Go engines only, per warm call)

| Input | Nuri | Nuri (no-interrupt) | Chroma |
|---|---:|---:|---:|
| Go | 1,893 / 203KB | 1,711 / 191KB | 1,247 / 73KB |
| HTML | 2,768 / 278KB | 2,279 / 249KB | 2,554 / 143KB |
| JavaScript | 3,618 / 566KB | 3,127 / 539KB | 2,686 / 169KB |
| Markdown | 1,153 / 142KB | 1,069 / 137KB | 1,576 / 90KB |
| TypeScript | 3,119 / 529KB | 2,842 / 514KB | 1,733 / 108KB |

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
| Markdown | nuri | 26 | 14 |
|  | shiki | 24 | 30 |
|  | chroma | 31 | 11 |
| TypeScript | nuri | 90 | 51 |
|  | shiki | 63 | 51 |
|  | chroma | 79 | 11 |

*Nuri and Shiki use full TextMate grammars (same Oniguruma engine); Chroma uses Pygments-model lexers with ~80 token types. The fidelity gap is by design, not a bug. Per-engine token dumps are saved alongside the snapshot for detailed diffing.*
