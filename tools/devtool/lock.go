package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/frostybee/nuri/internal/assetfs"
)

const lockfileName = "provenance.lock.json"

type Lockfile struct {
	Version     int                      `json:"version"`
	GeneratedAt string                   `json:"generatedAt"`
	Module      string                   `json:"module"`
	Submodules  map[string]SubmoduleInfo `json:"submodules"`
	Artifacts   map[string]ArtifactInfo  `json:"artifacts"`
	Grammars    map[string]string        `json:"grammars"`
	Themes      map[string]string        `json:"themes"`
}

type SubmoduleInfo struct {
	URL    string `json:"url"`
	Commit string `json:"commit"`
	Ref    string `json:"ref"`
}

type ArtifactInfo struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func generateLockfile(root string) error {
	subs, err := parseSubmoduleStatus(root)
	if err != nil {
		return fmt.Errorf("parse submodule status: %w", err)
	}

	urls, err := parseGitmodules(root)
	if err != nil {
		return fmt.Errorf("parse .gitmodules: %w", err)
	}
	for name, url := range urls {
		if info, ok := subs[name]; ok {
			info.URL = url
			subs[name] = info
		}
	}

	wasmHash, err := hashFile(filepath.Join(root, "resources", "wasm", "onig.wasm"))
	if err != nil {
		return fmt.Errorf("hash onig.wasm: %w", err)
	}

	grammars, err := hashDir(filepath.Join(root, "bundle", "full", "grammars"))
	if err != nil {
		return fmt.Errorf("hash grammars: %w", err)
	}

	themes, err := hashDir(filepath.Join(root, "bundle", "full", "themes"))
	if err != nil {
		return fmt.Errorf("hash themes: %w", err)
	}

	modulePath, err := readModulePath(root)
	if err != nil {
		return fmt.Errorf("read module path: %w", err)
	}

	lf := Lockfile{
		Version:     1,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Module:      modulePath,
		Submodules:  subs,
		Artifacts: map[string]ArtifactInfo{
			"onig.wasm": {
				Path:   "resources/wasm/onig.wasm",
				SHA256: wasmHash,
			},
		},
		Grammars: grammars,
		Themes:   themes,
	}

	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lockfile: %w", err)
	}
	data = append(data, '\n')

	dst := filepath.Join(root, lockfileName)
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("write lockfile: %w", err)
	}

	fmt.Printf("Lockfile generated: %s\n", lockfileName)
	fmt.Printf("  Submodules: %d\n", len(subs))
	fmt.Printf("  Grammars:   %d (sha256)\n", len(grammars))
	fmt.Printf("  Themes:     %d (sha256)\n", len(themes))
	fmt.Printf("  Artifacts:  1 (onig.wasm)\n")
	return nil
}

func verifyLockfile(root string) error {
	failures, err := runVerify(root)
	if err != nil {
		return err
	}

	fmt.Println()
	if failures > 0 {
		fmt.Printf("Verification FAILED: %d issue(s) found\n", failures)
		os.Exit(1)
	}
	fmt.Println("Verification PASSED")
	return nil
}

func loadLockfile(path string) (*Lockfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("malformed lockfile: %w", err)
	}
	return &lf, nil
}

func runVerify(root string) (int, error) {
	lf, err := loadLockfile(filepath.Join(root, lockfileName))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("lockfile not found; run 'go run ./tools/devtool lock' to generate it")
		}
		return 0, err
	}
	if lf.Version != 1 {
		return 0, fmt.Errorf("unsupported lockfile version %d; update your tools", lf.Version)
	}

	var failures int

	// Verify submodules.
	subs, err := parseSubmoduleStatus(root)
	if err != nil {
		return 0, fmt.Errorf("parse submodule status: %w", err)
	}
	urls, err := parseGitmodules(root)
	if err != nil {
		return 0, fmt.Errorf("parse .gitmodules: %w", err)
	}

	for name, expected := range lf.Submodules {
		actual, ok := subs[name]
		if !ok {
			fmt.Printf("  [FAIL] submodule %s: not found\n", name)
			failures++
			continue
		}
		if actual.Commit != expected.Commit {
			fmt.Printf("  [FAIL] submodule %s:\n", name)
			fmt.Printf("    lockfile: %s\n", expected.Commit)
			fmt.Printf("    current:  %s\n", actual.Commit)
			failures++
		} else {
			fmt.Printf("  [PASS] submodule %s: %s\n", name, short(expected.Commit))
		}
		if actualURL, ok := urls[name]; ok && actualURL != expected.URL {
			fmt.Printf("  [FAIL] submodule %s URL:\n", name)
			fmt.Printf("    lockfile: %s\n", expected.URL)
			fmt.Printf("    current:  %s\n", actualURL)
			failures++
		}
	}

	// Verify artifacts.
	for name, expected := range lf.Artifacts {
		actual, err := hashFile(filepath.Join(root, filepath.FromSlash(expected.Path)))
		if err != nil {
			fmt.Printf("  [FAIL] artifact %s: %v\n", name, err)
			failures++
			continue
		}
		if actual != expected.SHA256 {
			fmt.Printf("  [FAIL] artifact %s:\n", name)
			fmt.Printf("    lockfile: %s\n", expected.SHA256)
			fmt.Printf("    current:  %s\n", actual)
			failures++
		} else {
			fmt.Printf("  [PASS] artifact %s: %s\n", name, short(expected.SHA256))
		}
	}

	// Verify grammars.
	failures += verifyAssetDir(root, "grammars", filepath.Join(root, "bundle", "full", "grammars"), lf.Grammars)

	// Verify themes.
	failures += verifyAssetDir(root, "themes", filepath.Join(root, "bundle", "full", "themes"), lf.Themes)

	// Verify core subset consistency.
	failures += verifyCoreSubset(root)

	// Verify the metadata indexes match the asset dirs.
	failures += verifyIndexes(root)

	return failures, nil
}

// verifyIndexes checks that each bundle's grammar metadata index agrees
// with the files actually present. Per file hashes alone cannot catch a
// stale index after a manual asset edit.
func verifyIndexes(root string) int {
	var failures int
	for _, bundle := range []string{"core", "full"} {
		dir := filepath.Join(root, "bundle", bundle, "grammars")
		label := "index (" + bundle + ")"

		compressed, err := os.ReadFile(filepath.Join(dir, assetfs.IndexFileName+".gz"))
		if err != nil {
			fmt.Printf("  [FAIL] %s: %v\n", label, err)
			failures++
			continue
		}
		data, err := gunzipBytes(compressed)
		if err != nil {
			fmt.Printf("  [FAIL] %s: gunzip: %v\n", label, err)
			failures++
			continue
		}
		var idx assetfs.Index
		if err := json.Unmarshal(data, &idx); err != nil {
			fmt.Printf("  [FAIL] %s: parse: %v\n", label, err)
			failures++
			continue
		}
		if idx.Version != assetfs.IndexVersion {
			fmt.Printf("  [FAIL] %s: version %d, want %d\n", label, idx.Version, assetfs.IndexVersion)
			failures++
			continue
		}

		onDisk := make(map[string]bool)
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("  [FAIL] %s: %v\n", label, err)
			failures++
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json.gz") || e.Name() == assetfs.IndexFileName+".gz" {
				continue
			}
			onDisk[strings.TrimSuffix(e.Name(), ".json.gz")] = true
		}

		var bundleFailures int
		for name := range idx.Grammars {
			if !onDisk[name] {
				fmt.Printf("  [FAIL] %s: lists %q but no %s.json.gz on disk\n", label, name, name)
				bundleFailures++
			}
		}
		for name := range onDisk {
			if _, ok := idx.Grammars[name]; !ok {
				fmt.Printf("  [FAIL] %s: %s.json.gz on disk but missing from index\n", label, name)
				bundleFailures++
			}
		}
		if bundleFailures == 0 {
			fmt.Printf("  [PASS] %s: %d entries consistent\n", label, len(idx.Grammars))
		}
		failures += bundleFailures
	}
	return failures
}

func verifyAssetDir(root, label, dir string, expected map[string]string) int {
	actual, err := hashDir(dir)
	if err != nil {
		fmt.Printf("  [FAIL] %s: %v\n", label, err)
		return 1
	}

	var failures int
	var matched int

	names := sortedKeys(expected)
	for _, name := range names {
		expectedHash := expected[name]
		actualHash, ok := actual[name]
		if !ok {
			fmt.Printf("  [FAIL] %s MISSING: %s\n", label, name)
			failures++
			continue
		}
		if actualHash != expectedHash {
			fmt.Printf("  [FAIL] %s MISMATCH %s:\n", label, name)
			fmt.Printf("    lockfile: %s\n", expectedHash)
			fmt.Printf("    current:  %s\n", actualHash)
			failures++
		} else {
			matched++
		}
	}

	for _, name := range sortedKeys(actual) {
		if _, ok := expected[name]; !ok {
			fmt.Printf("  [FAIL] %s UNEXPECTED: %s\n", label, name)
			failures++
		}
	}

	if failures == 0 {
		fmt.Printf("  [PASS] %s: %d/%d match\n", label, matched, len(expected))
	} else {
		fmt.Printf("  [FAIL] %s: %d/%d match, %d issue(s)\n", label, matched, len(expected), failures)
	}
	return failures
}

func verifyCoreSubset(root string) int {
	coreDir := filepath.Join(root, "bundle", "core", "grammars")
	fullDir := filepath.Join(root, "bundle", "full", "grammars")

	coreHashes, err := hashDir(coreDir)
	if err != nil {
		fmt.Printf("  [FAIL] core subset: %v\n", err)
		return 1
	}
	fullHashes, err := hashDir(fullDir)
	if err != nil {
		fmt.Printf("  [FAIL] core subset: %v\n", err)
		return 1
	}

	var failures, compared int
	for name, coreHash := range coreHashes {
		// The metadata index legitimately differs between bundles: the
		// core index lists only the core grammar subset.
		if name == assetfs.IndexFileName+".gz" {
			continue
		}
		compared++
		fullHash, ok := fullHashes[name]
		if !ok {
			fmt.Printf("  [FAIL] core subset: %s not in full bundle\n", name)
			failures++
		} else if coreHash != fullHash {
			fmt.Printf("  [FAIL] core subset: %s differs from full bundle\n", name)
			failures++
		}
	}

	if failures == 0 {
		fmt.Printf("  [PASS] core subset: %d/%d consistent\n", compared, compared)
	}
	return failures
}

// --- helpers ---

func parseSubmoduleStatus(root string) (map[string]SubmoduleInfo, error) {
	cmd := exec.Command("git", "submodule", "status")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git submodule status: %w", err)
	}

	result := make(map[string]SubmoduleInfo)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip leading status char (+, -, U, or space).
		if line[0] == '+' || line[0] == '-' || line[0] == 'U' {
			line = line[1:]
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		commit := parts[0]
		path := parts[1]

		var ref string
		if len(parts) >= 3 {
			ref = strings.Trim(parts[2], "()")
		}

		result[path] = SubmoduleInfo{
			Commit: commit,
			Ref:    ref,
		}
	}
	return result, nil
}

func parseGitmodules(root string) (map[string]string, error) {
	f, err := os.Open(filepath.Join(root, ".gitmodules"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	urls := make(map[string]string)
	var currentPath string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "path = ") {
			currentPath = strings.TrimPrefix(line, "path = ")
		} else if strings.HasPrefix(line, "url = ") && currentPath != "" {
			urls[currentPath] = strings.TrimPrefix(line, "url = ")
			currentPath = ""
		}
	}
	return urls, scanner.Err()
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

func hashDir(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.IsDir() || (!strings.HasSuffix(e.Name(), ".json") && !strings.HasSuffix(e.Name(), ".json.gz")) {
			continue
		}
		h, err := hashFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("hash %s: %w", e.Name(), err)
		}
		result[e.Name()] = "sha256:" + h
	}
	return result, nil
}

func readModulePath(root string) (string, error) {
	f, err := os.Open(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

func short(s string) string {
	if len(s) > 12 {
		return s[:12] + "..."
	}
	return s
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
