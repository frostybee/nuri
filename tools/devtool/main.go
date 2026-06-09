package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/devtool <command>")
		fmt.Fprintln(os.Stderr, "commands: sync, generate, lock, verify, all")
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

	var coreCount, fullCount, themeCount int

	entries, err := os.ReadDir(submoduleGrammars)
	if err != nil {
		return fmt.Errorf("read grammars dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		src := filepath.Join(submoduleGrammars, e.Name())

		if err := copyFile(src, filepath.Join(fullGrammarDir, e.Name())); err != nil {
			return fmt.Errorf("copy %s to full: %w", e.Name(), err)
		}
		fullCount++

		baseName := strings.TrimSuffix(e.Name(), ".json")
		if coreSet[baseName] {
			if err := copyFile(src, filepath.Join(coreGrammarDir, e.Name())); err != nil {
				return fmt.Errorf("copy %s to core: %w", e.Name(), err)
			}
			coreCount++
		}
	}

	themeEntries, err := os.ReadDir(submoduleThemes)
	if err != nil {
		return fmt.Errorf("read themes dir: %w", err)
	}
	for _, e := range themeEntries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		src := filepath.Join(submoduleThemes, e.Name())
		if err := copyFile(src, filepath.Join(coreThemeDir, e.Name())); err != nil {
			return fmt.Errorf("copy theme %s to core: %w", e.Name(), err)
		}
		if err := copyFile(src, filepath.Join(fullThemeDir, e.Name())); err != nil {
			return fmt.Errorf("copy theme %s to full: %w", e.Name(), err)
		}
		themeCount++
	}

	fmt.Printf("Synced %d grammars to core, %d to full, %d themes to both\n", coreCount, fullCount, themeCount)
	return nil
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
