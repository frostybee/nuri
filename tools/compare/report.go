package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

// EngineResult holds benchmark results for one engine on one input.
type EngineResult struct {
	ColdMs float64 `json:"coldMs"`
	WarmMs float64 `json:"warmMs"`
	AllocB int64   `json:"allocB"`
	Allocs int64   `json:"allocs"`
	Tokens int     `json:"tokens"`
	Scopes int     `json:"scopes"`
}

// Snapshot captures a complete benchmark run.
type Snapshot struct {
	Timestamp string                          `json:"timestamp"`
	Machine   string                          `json:"machine"`
	Versions  map[string]string               `json:"versions"`
	Iters     int                             `json:"iters"`
	Theme     string                          `json:"theme"`
	Results   map[string]map[string]EngineResult `json:"results"`
}

func saveSnapshot(s *Snapshot, tokenDumps map[string]map[string]string) (string, error) {
	ts := strings.ReplaceAll(s.Timestamp[:19], ":", "-")
	dir := filepath.Join("results", ts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	snapPath := filepath.Join(dir, "snapshot.json")
	if err := os.WriteFile(snapPath, data, 0o644); err != nil {
		return "", err
	}

	for engine, langs := range tokenDumps {
		for lang, dump := range langs {
			fname := filepath.Join(dir, fmt.Sprintf("%s-%s.tokens", engine, lang))
			if err := os.WriteFile(fname, []byte(dump), 0o644); err != nil {
				return "", err
			}
		}
	}

	return dir, nil
}

func loadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Snapshot
	return &s, json.Unmarshal(data, &s)
}

func inputNames(s *Snapshot) []string {
	set := make(map[string]struct{})
	for name := range s.Results {
		set[name] = struct{}{}
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func generateMarkdown(s *Snapshot, prev *Snapshot, inputs []Input) string {
	var sb strings.Builder
	sb.WriteString("# Nuri vs Shiki vs Chroma — Benchmark Results\n\n")
	sb.WriteString(fmt.Sprintf("*Generated: %s | Machine: %s | Warm iterations: %d | Theme: %s*\n", s.Timestamp, s.Machine, s.Iters, s.Theme))

	var verParts []string
	for k, v := range s.Versions {
		verParts = append(verParts, fmt.Sprintf("%s %s", k, v))
	}
	sort.Strings(verParts)
	sb.WriteString(fmt.Sprintf("*Versions: %s*\n\n", strings.Join(verParts, ", ")))

	names := inputNames(s)

	hasNuri := hasEngine(s, "nuri")
	hasNuriNoInt := hasEngine(s, "nuri-no-interrupt")
	hasShiki := hasEngine(s, "shiki")
	hasChroma := hasEngine(s, "chroma")

	// Inputs metadata table.
	sb.WriteString("## Inputs\n\n")
	writeInputsTable(&sb, inputs, names)

	// Speed table.
	sb.WriteString("\n## Speed (warm median, ms/op)\n\n")
	writeSpeedTable(&sb, s, prev, names, inputs, hasNuri, hasNuriNoInt, hasShiki, hasChroma)

	// Cold start table.
	sb.WriteString("\n## Cold start (first call, ms)\n\n")
	writeColdTable(&sb, s, prev, names, hasNuri, hasShiki, hasChroma)

	// Allocations table.
	if hasNuri || hasNuriNoInt || hasShiki || hasChroma {
		sb.WriteString("\n## Allocations (per warm call)\n\n")
		writeAllocTable(&sb, s, names, hasNuri, hasNuriNoInt, hasShiki, hasChroma)
	}

	// Fidelity table.
	sb.WriteString("\n## Fidelity\n\n")
	writeFidelityTable(&sb, s, names, hasNuri, hasShiki, hasChroma)

	sb.WriteString("\n*Nuri and Shiki use full TextMate grammars (same Oniguruma engine); Chroma uses Pygments-model lexers with ~80 token types. The fidelity gap is by design, not a bug. Per-engine token dumps are saved alongside the snapshot for detailed diffing.*\n")

	return sb.String()
}

func hasEngine(s *Snapshot, engine string) bool {
	for _, engines := range s.Results {
		if _, ok := engines[engine]; ok {
			return true
		}
	}
	return false
}

func writeInputsTable(sb *strings.Builder, inputs []Input, names []string) {
	sb.WriteString("| Input | Language | Bytes | Lines |\n")
	sb.WriteString("|---|---|---:|---:|\n")

	byName := make(map[string]Input, len(inputs))
	for _, inp := range inputs {
		byName[inp.Name] = inp
	}
	for _, name := range names {
		inp := byName[name]
		lines := strings.Count(inp.Code, "\n")
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d |\n", inp.Name, inp.Lang, len(inp.Code), lines))
	}
}

func writeSpeedTable(sb *strings.Builder, s, prev *Snapshot, names []string, inputs []Input, hasNuri, hasNuriNoInt, hasShiki, hasChroma bool) {
	header := "| Input |"
	sep := "|---|"
	if hasNuri {
		header += " Nuri |"
		sep += "---:|"
	}
	if hasNuriNoInt {
		header += " Nuri (no-interrupt) |"
		sep += "---:|"
	}
	if hasShiki {
		header += " Shiki |"
		sep += "---:|"
	}
	if hasChroma {
		header += " Chroma |"
		sep += "---:|"
	}
	sb.WriteString(header + "\n")
	sb.WriteString(sep + "\n")

	bytesByName := make(map[string]int, len(inputs))
	for _, inp := range inputs {
		bytesByName[inp.Name] = len(inp.Code)
	}

	for _, name := range names {
		row := fmt.Sprintf("| %s |", name)
		engines := s.Results[name]
		b := bytesByName[name]
		if hasNuri {
			row += fmtSpeedCell(engines, "nuri", prev, name, b) + " |"
		}
		if hasNuriNoInt {
			row += fmtSpeedCell(engines, "nuri-no-interrupt", prev, name, b) + " |"
		}
		if hasShiki {
			row += fmtSpeedCell(engines, "shiki", prev, name, b) + " |"
		}
		if hasChroma {
			row += fmtSpeedCell(engines, "chroma", prev, name, b) + " |"
		}
		sb.WriteString(row + "\n")
	}
}

func fmtSpeedCell(engines map[string]EngineResult, engine string, prev *Snapshot, name string, inputBytes int) string {
	r, ok := engines[engine]
	if !ok {
		return " —"
	}
	cell := fmt.Sprintf(" %.2f", r.WarmMs)
	if r.WarmMs > 0 && inputBytes > 0 {
		kbPerSec := (float64(inputBytes) / 1024.0) / (r.WarmMs / 1000.0)
		cell += fmt.Sprintf(" (%.0f KB/s)", kbPerSec)
	}
	if prev != nil {
		if pe, ok := prev.Results[name]; ok {
			if pr, ok := pe[engine]; ok && pr.WarmMs > 0 {
				pct := (r.WarmMs - pr.WarmMs) / pr.WarmMs * 100
				if pct > 0 {
					cell += fmt.Sprintf(" [+%.0f%%]", pct)
				} else {
					cell += fmt.Sprintf(" [%.0f%%]", pct)
				}
			}
		}
	}
	return cell
}

func writeColdTable(sb *strings.Builder, s, prev *Snapshot, names []string, hasNuri, hasShiki, hasChroma bool) {
	header := "| Input |"
	sep := "|---|"
	if hasNuri {
		header += " Nuri |"
		sep += "---:|"
	}
	if hasShiki {
		header += " Shiki |"
		sep += "---:|"
	}
	if hasChroma {
		header += " Chroma |"
		sep += "---:|"
	}
	sb.WriteString(header + "\n")
	sb.WriteString(sep + "\n")

	for _, name := range names {
		row := fmt.Sprintf("| %s |", name)
		engines := s.Results[name]
		if hasNuri {
			if r, ok := engines["nuri"]; ok {
				row += fmt.Sprintf(" %.2f |", r.ColdMs)
			} else {
				row += " — |"
			}
		}
		if hasShiki {
			if r, ok := engines["shiki"]; ok {
				row += fmt.Sprintf(" %.2f |", r.ColdMs)
			} else {
				row += " — |"
			}
		}
		if hasChroma {
			if r, ok := engines["chroma"]; ok {
				row += fmt.Sprintf(" %.2f |", r.ColdMs)
			} else {
				row += " — |"
			}
		}
		sb.WriteString(row + "\n")
	}
}

func writeAllocTable(sb *strings.Builder, s *Snapshot, names []string, hasNuri, hasNuriNoInt, hasShiki, hasChroma bool) {
	header := "| Input |"
	sep := "|---|"
	if hasNuri {
		header += " Nuri |"
		sep += "---:|"
	}
	if hasNuriNoInt {
		header += " Nuri (no-interrupt) |"
		sep += "---:|"
	}
	if hasShiki {
		header += " Shiki |"
		sep += "---:|"
	}
	if hasChroma {
		header += " Chroma |"
		sep += "---:|"
	}
	sb.WriteString(header + "\n")
	sb.WriteString(sep + "\n")

	for _, name := range names {
		row := fmt.Sprintf("| %s |", name)
		engines := s.Results[name]
		if hasNuri {
			row += fmtAllocCell(engines, "nuri") + " |"
		}
		if hasNuriNoInt {
			row += fmtAllocCell(engines, "nuri-no-interrupt") + " |"
		}
		if hasShiki {
			row += fmtAllocCell(engines, "shiki") + " |"
		}
		if hasChroma {
			row += fmtAllocCell(engines, "chroma") + " |"
		}
		sb.WriteString(row + "\n")
	}
}

func fmtAllocCell(engines map[string]EngineResult, engine string) string {
	r, ok := engines[engine]
	if !ok {
		return " —"
	}
	if r.Allocs == 0 && r.AllocB == 0 {
		return " —"
	}
	if r.Allocs == 0 {
		return fmt.Sprintf(" %s", fmtBytes(r.AllocB))
	}
	return fmt.Sprintf(" %s / %s", fmtAllocs(r.Allocs), fmtBytes(r.AllocB))
}

func fmtAllocs(n int64) string {
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}

func fmtBytes(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(b)/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.0fKB", float64(b)/1024)
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func writeFidelityTable(sb *strings.Builder, s *Snapshot, names []string, hasNuri, hasShiki, hasChroma bool) {
	sb.WriteString("| Input | Engine | Tokens | Distinct scopes/types |\n")
	sb.WriteString("|---|---|---:|---:|\n")

	for _, name := range names {
		engines := s.Results[name]
		first := true
		for _, eng := range []string{"nuri", "shiki", "chroma"} {
			switch eng {
			case "nuri":
				if !hasNuri {
					continue
				}
			case "shiki":
				if !hasShiki {
					continue
				}
			case "chroma":
				if !hasChroma {
					continue
				}
			}
			r, ok := engines[eng]
			if !ok {
				continue
			}
			label := name
			if !first {
				label = ""
			}
			first = false
			sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d |\n", label, eng, r.Tokens, r.Scopes))
		}
	}
}

func printResults(s *Snapshot, inputs []Input) {
	names := inputNames(s)

	fmt.Println()
	fmt.Printf("Benchmark results — %s | %s | %d warm iters | theme: %s\n", s.Timestamp[:19], s.Machine, s.Iters, s.Theme)
	for k, v := range s.Versions {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Println()

	bytesByName := make(map[string]int, len(inputs))
	linesByName := make(map[string]int, len(inputs))
	for _, inp := range inputs {
		bytesByName[inp.Name] = len(inp.Code)
		linesByName[inp.Name] = strings.Count(inp.Code, "\n")
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "INPUT\tBYTES\tLINES\tENGINE\tCOLD (ms)\tWARM (ms)\tKB/s\tALLOCS\tMEMORY\tTOKENS\tSCOPES\t")
	fmt.Fprintln(tw, "-----\t-----\t-----\t------\t---------\t---------\t----\t------\t------\t------\t------\t")

	for _, name := range names {
		engines := s.Results[name]
		engineNames := make([]string, 0, len(engines))
		for e := range engines {
			engineNames = append(engineNames, e)
		}
		sort.Strings(engineNames)
		for i, eng := range engineNames {
			r := engines[eng]
			allocStr := "—"
			memStr := "—"
			if r.AllocB > 0 || r.Allocs > 0 {
				allocStr = fmtAllocs(r.Allocs)
				memStr = fmtBytes(r.AllocB)
			}
			kbps := "—"
			if r.WarmMs > 0 && bytesByName[name] > 0 {
				kbps = fmt.Sprintf("%.0f", (float64(bytesByName[name])/1024.0)/(r.WarmMs/1000.0))
			}
			label, bStr, lStr := name, fmt.Sprintf("%d", bytesByName[name]), fmt.Sprintf("%d", linesByName[name])
			if i > 0 {
				label, bStr, lStr = "", "", ""
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%.2f\t%.2f\t%s\t%s\t%s\t%d\t%d\t\n", label, bStr, lStr, eng, r.ColdMs, r.WarmMs, kbps, allocStr, memStr, r.Tokens, r.Scopes)
		}
	}
	tw.Flush()
}

func nowTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
