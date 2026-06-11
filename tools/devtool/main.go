package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/frostybee/nuri/internal/assetfs"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/devtool <command>")
		fmt.Fprintln(os.Stderr, "commands: sync, generate, lock, verify, notices, all")
		os.Exit(1)
	}

	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "sync":
		if err := syncGrammars(root); err != nil {
			fmt.Fprintf(os.Stderr, "sync: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Regenerating lockfile...")
		if err := generateLockfile(root); err != nil {
			fmt.Fprintf(os.Stderr, "lock: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Regenerating NOTICE files...")
		if err := generateNotices(root); err != nil {
			fmt.Fprintf(os.Stderr, "notices: %v\n", err)
			os.Exit(1)
		}
	case "generate":
		if err := generate(root); err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
	case "lock":
		if err := generateLockfile(root); err != nil {
			fmt.Fprintf(os.Stderr, "lock: %v\n", err)
			os.Exit(1)
		}
	case "notices":
		if err := generateNotices(root); err != nil {
			fmt.Fprintf(os.Stderr, "notices: %v\n", err)
			os.Exit(1)
		}
	case "verify":
		fmt.Println("Verifying provenance.lock.json...")
		if err := verifyLockfile(root); err != nil {
			fmt.Fprintf(os.Stderr, "verify: %v\n", err)
			os.Exit(1)
		}
	case "all":
		if err := syncGrammars(root); err != nil {
			fmt.Fprintf(os.Stderr, "sync: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Regenerating lockfile...")
		if err := generateLockfile(root); err != nil {
			fmt.Fprintf(os.Stderr, "lock: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Regenerating NOTICE files...")
		if err := generateNotices(root); err != nil {
			fmt.Fprintf(os.Stderr, "notices: %v\n", err)
			os.Exit(1)
		}
		if err := generate(root); err != nil {
			fmt.Fprintf(os.Stderr, "generate: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

func syncGrammars(root string) error {
	submoduleGrammars := filepath.Join(root, "grammars-themes", "packages", "tm-grammars", "grammars")
	submoduleThemes := filepath.Join(root, "grammars-themes", "packages", "tm-themes", "themes")

	if _, err := os.Stat(submoduleGrammars); os.IsNotExist(err) {
		fmt.Println("grammars-themes submodule not initialized, running git submodule update...")
		if err := runCmd(root, "git", "submodule", "update", "--init", "--recursive"); err != nil {
			return fmt.Errorf("git submodule update: %w", err)
		}
	}

	coreList, err := readGrammarList(filepath.Join(root, "bundle", "core", "grammars.txt"))
	if err != nil {
		return fmt.Errorf("read core grammar list: %w", err)
	}
	coreSet := make(map[string]bool, len(coreList))
	for _, name := range coreList {
		coreSet[name] = true
	}

	coreGrammarDir := filepath.Join(root, "bundle", "core", "grammars")
	fullGrammarDir := filepath.Join(root, "bundle", "full", "grammars")
	coreThemeDir := filepath.Join(root, "bundle", "core", "themes")
	fullThemeDir := filepath.Join(root, "bundle", "full", "themes")

	// Cleaning first makes sync idempotent and removes assets that
	// disappeared upstream.
	for _, dir := range []string{coreGrammarDir, fullGrammarDir, coreThemeDir, fullThemeDir} {
		if err := cleanAssetDir(dir); err != nil {
			return fmt.Errorf("clean %s: %w", dir, err)
		}
	}

	var coreCount, fullCount, themeCount int
	coreIndex := assetfs.Index{Version: assetfs.IndexVersion, Grammars: make(map[string]assetfs.GrammarMeta)}
	fullIndex := assetfs.Index{Version: assetfs.IndexVersion, Grammars: make(map[string]assetfs.GrammarMeta)}

	entries, err := os.ReadDir(submoduleGrammars)
	if err != nil {
		return fmt.Errorf("read grammars dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		baseName := strings.TrimSuffix(e.Name(), ".json")
		if e.Name() == assetfs.IndexFileName {
			return fmt.Errorf("upstream grammar %q collides with the reserved index file name", e.Name())
		}

		raw, err := os.ReadFile(filepath.Join(submoduleGrammars, e.Name()))
		if err != nil {
			return fmt.Errorf("read %s: %w", e.Name(), err)
		}
		minified, err := minifyStripTop(raw, grammarDropFields...)
		if err != nil {
			return fmt.Errorf("minify %s: %w", e.Name(), err)
		}
		var meta assetfs.GrammarMeta
		if err := json.Unmarshal(minified, &meta); err != nil {
			return fmt.Errorf("probe %s: %w", e.Name(), err)
		}
		compressed, err := gzipBytes(minified)
		if err != nil {
			return fmt.Errorf("gzip %s: %w", e.Name(), err)
		}

		outName := e.Name() + ".gz"
		if err := writeAsset(filepath.Join(fullGrammarDir, outName), compressed); err != nil {
			return fmt.Errorf("write %s to full: %w", outName, err)
		}
		fullIndex.Grammars[baseName] = meta
		fullCount++

		if coreSet[baseName] {
			// The identical byte slice goes to core so the lockfile's
			// core subset check stays byte exact.
			if err := writeAsset(filepath.Join(coreGrammarDir, outName), compressed); err != nil {
				return fmt.Errorf("write %s to core: %w", outName, err)
			}
			coreIndex.Grammars[baseName] = meta
			coreCount++
		}
	}

	var missing []string
	for _, name := range coreList {
		if _, ok := coreIndex.Grammars[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("core grammar list entries missing from submodule: %v", missing)
	}

	if err := writeIndex(filepath.Join(fullGrammarDir, assetfs.IndexFileName+".gz"), fullIndex); err != nil {
		return fmt.Errorf("write full index: %w", err)
	}
	if err := writeIndex(filepath.Join(coreGrammarDir, assetfs.IndexFileName+".gz"), coreIndex); err != nil {
		return fmt.Errorf("write core index: %w", err)
	}

	themeEntries, err := os.ReadDir(submoduleThemes)
	if err != nil {
		return fmt.Errorf("read themes dir: %w", err)
	}
	for _, e := range themeEntries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(submoduleThemes, e.Name()))
		if err != nil {
			return fmt.Errorf("read theme %s: %w", e.Name(), err)
		}
		minified, err := minifyStripTop(raw, themeDropFields...)
		if err != nil {
			return fmt.Errorf("minify theme %s: %w", e.Name(), err)
		}
		compressed, err := gzipBytes(minified)
		if err != nil {
			return fmt.Errorf("gzip theme %s: %w", e.Name(), err)
		}
		outName := e.Name() + ".gz"
		if err := writeAsset(filepath.Join(coreThemeDir, outName), compressed); err != nil {
			return fmt.Errorf("write theme %s to core: %w", outName, err)
		}
		if err := writeAsset(filepath.Join(fullThemeDir, outName), compressed); err != nil {
			return fmt.Errorf("write theme %s to full: %w", outName, err)
		}
		themeCount++
	}

	fmt.Printf("Synced %d grammars to core, %d to full, %d themes to both (minified, pruned, gzipped)\n",
		coreCount, fullCount, themeCount)
	return nil
}

// cleanAssetDir removes all .json and .json.gz files from dir, creating
// the directory if it does not exist.
func cleanAssetDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0o755)
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".json") || strings.HasSuffix(e.Name(), ".json.gz") {
			if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeIndex(dst string, idx assetfs.Index) error {
	data, err := marshalCompact(idx)
	if err != nil {
		return err
	}
	compressed, err := gzipBytes(data)
	if err != nil {
		return err
	}
	return writeAsset(dst, compressed)
}

func generate(root string) error {
	fmt.Println("Ensuring submodules are initialized...")
	if err := runCmd(root, "git", "submodule", "update", "--init", "--recursive"); err != nil {
		return fmt.Errorf("git submodule update: %w", err)
	}

	vsctm := filepath.Join(root, "vscode-textmate")
	if _, err := os.Stat(filepath.Join(vsctm, "out", "src", "main.js")); os.IsNotExist(err) {
		fmt.Println("Building vscode-textmate...")
		if err := runCmd(vsctm, "npm", "install"); err != nil {
			return fmt.Errorf("npm install (vscode-textmate): %w", err)
		}
		if err := runCmd(vsctm, "npm", "run", "compile"); err != nil {
			return fmt.Errorf("npm run compile (vscode-textmate): %w", err)
		}
	} else {
		fmt.Println("vscode-textmate already built, skipping...")
	}

	genfixtures := filepath.Join(root, "tools", "genfixtures")
	fmt.Println("Installing fixture generator dependencies...")
	if err := runCmd(genfixtures, "npm", "ci"); err != nil {
		return fmt.Errorf("npm ci (genfixtures): %w", err)
	}

	fmt.Println("Generating fixtures...")
	args := []string{"generate.mjs"}
	if len(os.Args) > 2 {
		args = append(args, os.Args[2:]...)
	}
	if err := runCmd(genfixtures, "node", args...); err != nil {
		return fmt.Errorf("node generate.mjs: %w", err)
	}

	fmt.Println("Done.")
	return nil
}

func readGrammarList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			names = append(names, line)
		}
	}
	return names, scanner.Err()
}

func writeAsset(dst string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
