package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "hello world"
	writeFile(t, path, content)

	got, err := hashFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := sha256Hex([]byte(content))
	if got != want {
		t.Errorf("hashFile = %s, want %s", got, want)
	}
}

func TestHashFile_NotFound(t *testing.T) {
	_, err := hashFile(filepath.Join(t.TempDir(), "nonexistent"))
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestHashDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.json"), `{"name":"a"}`)
	writeFile(t, filepath.Join(dir, "b.json"), `{"name":"b"}`)
	writeFile(t, filepath.Join(dir, "skip.txt"), "not json")
	os.Mkdir(filepath.Join(dir, "subdir"), 0o755)

	result, err := hashDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	for _, name := range []string{"a.json", "b.json"} {
		hash, ok := result[name]
		if !ok {
			t.Errorf("missing entry for %s", name)
			continue
		}
		data, _ := os.ReadFile(filepath.Join(dir, name))
		want := "sha256:" + sha256Hex(data)
		if hash != want {
			t.Errorf("%s hash = %s, want %s", name, hash, want)
		}
	}
	if _, ok := result["skip.txt"]; ok {
		t.Error("skip.txt should not be included (not .json)")
	}
}

func TestHashDir_Empty(t *testing.T) {
	dir := t.TempDir()
	result, err := hashDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}

func TestParseGitmodules(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".gitmodules"), `[submodule "grammars-themes"]
	path = grammars-themes
	url = https://github.com/shikijs/textmate-grammars-themes.git
[submodule "vscode-textmate"]
	path = vscode-textmate
	url = https://github.com/microsoft/vscode-textmate.git
	ignore = dirty
`)

	urls, err := parseGitmodules(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(urls))
	}
	if urls["grammars-themes"] != "https://github.com/shikijs/textmate-grammars-themes.git" {
		t.Errorf("grammars-themes URL = %q", urls["grammars-themes"])
	}
	if urls["vscode-textmate"] != "https://github.com/microsoft/vscode-textmate.git" {
		t.Errorf("vscode-textmate URL = %q", urls["vscode-textmate"])
	}
}

func TestReadModulePath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), `module github.com/frostybee/nuri

go 1.25

require (
	github.com/tetratelabs/wazero v1.12.0
)
`)

	got, err := readModulePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "github.com/frostybee/nuri" {
		t.Errorf("readModulePath = %q", got)
	}
}

func TestReadModulePath_Missing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "go 1.25\n")

	_, err := readModulePath(dir)
	if err == nil {
		t.Fatal("expected error when module directive missing")
	}
}

func TestVerifyAssetDir_AllMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.json"), `{"a":1}`)
	writeFile(t, filepath.Join(dir, "b.json"), `{"b":2}`)

	hashes, _ := hashDir(dir)
	failures := verifyAssetDir("", "test", dir, hashes)
	if failures != 0 {
		t.Errorf("expected 0 failures, got %d", failures)
	}
}

func TestVerifyAssetDir_Mismatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.json"), `{"a":1}`)

	expected := map[string]string{
		"a.json": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
	}
	failures := verifyAssetDir("", "test", dir, expected)
	if failures != 1 {
		t.Errorf("expected 1 failure (mismatch), got %d", failures)
	}
}

func TestVerifyAssetDir_Missing(t *testing.T) {
	dir := t.TempDir()

	expected := map[string]string{
		"gone.json": "sha256:abc",
	}
	failures := verifyAssetDir("", "test", dir, expected)
	if failures != 1 {
		t.Errorf("expected 1 failure (missing), got %d", failures)
	}
}

func TestVerifyAssetDir_Unexpected(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "extra.json"), `{"extra":true}`)

	failures := verifyAssetDir("", "test", dir, map[string]string{})
	if failures != 1 {
		t.Errorf("expected 1 failure (unexpected), got %d", failures)
	}
}

func TestVerifyCoreSubset_Consistent(t *testing.T) {
	root := t.TempDir()
	fullDir := filepath.Join(root, "bundle", "full", "grammars")
	coreDir := filepath.Join(root, "bundle", "core", "grammars")

	writeFile(t, filepath.Join(fullDir, "go.json"), `{"name":"go"}`)
	writeFile(t, filepath.Join(coreDir, "go.json"), `{"name":"go"}`)

	failures := verifyCoreSubset(root)
	if failures != 0 {
		t.Errorf("expected 0 failures, got %d", failures)
	}
}

func TestVerifyCoreSubset_Differs(t *testing.T) {
	root := t.TempDir()
	fullDir := filepath.Join(root, "bundle", "full", "grammars")
	coreDir := filepath.Join(root, "bundle", "core", "grammars")

	writeFile(t, filepath.Join(fullDir, "go.json"), `{"name":"go"}`)
	writeFile(t, filepath.Join(coreDir, "go.json"), `{"name":"go","extra":true}`)

	failures := verifyCoreSubset(root)
	if failures != 1 {
		t.Errorf("expected 1 failure, got %d", failures)
	}
}

func TestVerifyCoreSubset_NotInFull(t *testing.T) {
	root := t.TempDir()
	fullDir := filepath.Join(root, "bundle", "full", "grammars")
	coreDir := filepath.Join(root, "bundle", "core", "grammars")

	os.MkdirAll(fullDir, 0o755)
	writeFile(t, filepath.Join(coreDir, "orphan.json"), `{}`)

	failures := verifyCoreSubset(root)
	if failures != 1 {
		t.Errorf("expected 1 failure, got %d", failures)
	}
}

func TestLoadLockfile_Roundtrip(t *testing.T) {
	dir := t.TempDir()

	original := Lockfile{
		Version:     1,
		GeneratedAt: "2026-06-09T15:30:00Z",
		Module:      "github.com/frostybee/nuri",
		Submodules: map[string]SubmoduleInfo{
			"grammars-themes": {
				URL:    "https://github.com/shikijs/textmate-grammars-themes.git",
				Commit: "022eed00a8dd29481123f08e1cccf5a5bfee23f9",
				Ref:    "heads/main",
			},
		},
		Artifacts: map[string]ArtifactInfo{
			"onig.wasm": {Path: "resources/wasm/onig.wasm", SHA256: "abc123"},
		},
		Grammars: map[string]string{"go.json": "sha256:def456"},
		Themes:   map[string]string{"dark.json": "sha256:789abc"},
	}

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, lockfileName)
	os.WriteFile(path, data, 0o644)

	loaded, err := loadLockfile(path)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version = %d, want %d", loaded.Version, original.Version)
	}
	if loaded.Module != original.Module {
		t.Errorf("module = %q, want %q", loaded.Module, original.Module)
	}
	if loaded.Submodules["grammars-themes"].Commit != original.Submodules["grammars-themes"].Commit {
		t.Error("submodule commit mismatch")
	}
	if loaded.Grammars["go.json"] != original.Grammars["go.json"] {
		t.Error("grammar hash mismatch")
	}
	if loaded.Themes["dark.json"] != original.Themes["dark.json"] {
		t.Error("theme hash mismatch")
	}
}

func TestLoadLockfile_NotFound(t *testing.T) {
	_, err := loadLockfile(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for nonexistent lockfile")
	}
}

func TestLoadLockfile_Malformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, lockfileName)
	os.WriteFile(path, []byte("not json"), 0o644)

	_, err := loadLockfile(path)
	if err == nil {
		t.Fatal("expected error for malformed lockfile")
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]string{"c": "3", "a": "1", "b": "2"}
	got := sortedKeys(m)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestShort(t *testing.T) {
	if got := short("abcdefghijklmnop"); got != "abcdefghijkl..." {
		t.Errorf("short(long) = %q", got)
	}
	if got := short("abc"); got != "abc" {
		t.Errorf("short(short) = %q", got)
	}
}
