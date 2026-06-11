package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	iters := flag.Int("n", 50, "warm iterations per input per engine")
	save := flag.Bool("save", false, "write results/ snapshot + regenerate RESULTS.md")
	skipShiki := flag.Bool("skip-shiki", false, "skip Shiki benchmark")
	skipChroma := flag.Bool("skip-chroma", false, "skip Chroma benchmark")
	skipNuri := flag.Bool("skip-nuri", false, "skip Nuri benchmark")
	comparePath := flag.String("compare", "", "compare against a previous snapshot JSON file")
	themeName := flag.String("theme", "github-dark", "theme name for all engines")
	flag.Parse()

	snap := &Snapshot{
		Timestamp: nowTimestamp(),
		Machine:   machineID(),
		Versions:  captureVersions(),
		Iters:     *iters,
		Theme:     *themeName,
		Results:   make(map[string]map[string]EngineResult),
	}
	tokenDumps := make(map[string]map[string]string)
	inputs := defaultInputs

	if !*skipNuri {
		fmt.Println("Running Nuri (default)...")
		nuriResults, err := benchNuri(inputs, *iters, *themeName, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nuri: %v\n", err)
			os.Exit(1)
		}
		mergeResults(snap, nuriResults, "nuri")

		fmt.Println("Running Nuri (no-interrupt)...")
		nuriNoInt, err := benchNuri(inputs, *iters, *themeName, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nuri (no-interrupt): %v\n", err)
			os.Exit(1)
		}
		mergeResults(snap, nuriNoInt, "nuri-no-interrupt")

		fmt.Println("Dumping Nuri tokens...")
		nuriDumps, err := dumpNuriTokens(inputs, *themeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "nuri token dump: %v\n", err)
		} else {
			tokenDumps["nuri"] = nuriDumps
		}
	}

	if !*skipChroma {
		fmt.Println("Running Chroma...")
		chromaResults, err := benchChroma(inputs, *iters, *themeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chroma: %v\n", err)
			os.Exit(1)
		}
		mergeResults(snap, chromaResults, "chroma")

		fmt.Println("Dumping Chroma tokens...")
		chromaDumps, err := dumpChromaTokens(inputs, *themeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "chroma token dump: %v\n", err)
		} else {
			tokenDumps["chroma"] = chromaDumps
		}
	}

	if !*skipShiki {
		fmt.Println("Running Shiki...")
		shikiResults, shikiDumps, err := benchShiki(inputs, *iters, *themeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "shiki: %v\n", err)
			os.Exit(1)
		}
		mergeResults(snap, shikiResults, "shiki")
		if shikiDumps != nil {
			tokenDumps["shiki"] = shikiDumps
		}
	}

	printResults(snap)

	if *save {
		dir, err := saveSnapshot(snap, tokenDumps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "save: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nSnapshot saved to %s/\n", dir)

		var prev *Snapshot
		if *comparePath != "" {
			prev, _ = loadSnapshot(*comparePath)
		}

		md := generateMarkdown(snap, prev)
		if err := os.WriteFile("RESULTS.md", []byte(md), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write RESULTS.md: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("RESULTS.md regenerated.")
	} else if *comparePath != "" {
		prev, err := loadSnapshot(*comparePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load previous snapshot: %v\n", err)
			os.Exit(1)
		}
		md := generateMarkdown(snap, prev)
		fmt.Println()
		fmt.Println(md)
	}
}

func mergeResults(snap *Snapshot, engineResults map[string]EngineResult, engine string) {
	for name, result := range engineResults {
		if snap.Results[name] == nil {
			snap.Results[name] = make(map[string]EngineResult)
		}
		snap.Results[name][engine] = result
	}
}

func machineID() string {
	host, _ := os.Hostname()
	return fmt.Sprintf("%s (%s)", host, runtime.GOARCH)
}

func captureVersions() map[string]string {
	v := map[string]string{
		"go": runtime.Version(),
	}

	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		v["nuri"] = "commit:" + strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("node", "--version").Output(); err == nil {
		v["node"] = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("go", "list", "-m", "github.com/alecthomas/chroma/v2").Output(); err == nil {
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) >= 2 {
			v["chroma"] = parts[1]
		}
	}

	return v
}
