package assetfs

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func gz(t *testing.T, data string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write([]byte(data)); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func testFS(t *testing.T) fs.FS {
	t.Helper()
	return New(fstest.MapFS{
		"grammars/go.json.gz":    &fstest.MapFile{Data: gz(t, `{"scopeName":"source.go"}`)},
		"grammars/index.json.gz": &fstest.MapFile{Data: gz(t, `{"version":1,"grammars":{}}`)},
		"themes/dark.json.gz":    &fstest.MapFile{Data: gz(t, `{"name":"dark"}`)},
		"plain.txt":              &fstest.MapFile{Data: []byte("plain bytes")},
	})
}

func TestReadFileVirtual(t *testing.T) {
	fsys := testFS(t)
	data, err := fs.ReadFile(fsys, "grammars/go.json")
	if err != nil {
		t.Fatalf("ReadFile virtual: %v", err)
	}
	if string(data) != `{"scopeName":"source.go"}` {
		t.Errorf("decompressed content = %q", data)
	}
}

func TestReadFileLiteralGz(t *testing.T) {
	fsys := testFS(t)
	data, err := fs.ReadFile(fsys, "grammars/go.json.gz")
	if err != nil {
		t.Fatalf("ReadFile literal .gz: %v", err)
	}
	if _, err := gzip.NewReader(bytes.NewReader(data)); err != nil {
		t.Errorf("literal .gz read should return compressed bytes: %v", err)
	}
}

func TestOpenVirtualStat(t *testing.T) {
	fsys := testFS(t)
	f, err := fsys.Open("grammars/go.json")
	if err != nil {
		t.Fatalf("Open virtual: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	want := `{"scopeName":"source.go"}`
	if string(data) != want {
		t.Errorf("content = %q, want %q", data, want)
	}

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Name() != "go.json" {
		t.Errorf("Stat name = %q, want go.json", info.Name())
	}
	if info.Size() != int64(len(want)) {
		t.Errorf("Stat size = %d, want %d (decompressed)", info.Size(), len(want))
	}
	if info.IsDir() {
		t.Error("Stat IsDir = true")
	}
}

func TestReadDirVirtual(t *testing.T) {
	fsys := testFS(t)
	entries, err := fs.ReadDir(fsys, "grammars")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	want := []string{"go.json", "index.json"}
	if len(names) != len(want) {
		t.Fatalf("ReadDir names = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("ReadDir names = %v, want %v (sorted virtual names)", names, want)
			break
		}
	}

	info, err := entries[0].Info()
	if err != nil {
		t.Fatalf("DirEntry.Info: %v", err)
	}
	if got, want := info.Size(), int64(len(`{"scopeName":"source.go"}`)); got != want {
		t.Errorf("DirEntry.Info size = %d, want %d (decompressed)", got, want)
	}
}

func TestPlainPassthrough(t *testing.T) {
	fsys := testFS(t)
	data, err := fs.ReadFile(fsys, "plain.txt")
	if err != nil {
		t.Fatalf("ReadFile plain: %v", err)
	}
	if string(data) != "plain bytes" {
		t.Errorf("plain content = %q", data)
	}
}

func TestPlainWins(t *testing.T) {
	fsys := New(fstest.MapFS{
		"x.json":    &fstest.MapFile{Data: []byte(`{"plain":true}`)},
		"x.json.gz": &fstest.MapFile{Data: gz(t, `{"compressed":true}`)},
	})

	data, err := fs.ReadFile(fsys, "x.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != `{"plain":true}` {
		t.Errorf("content = %q, want plain file to win", data)
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	count := 0
	for _, e := range entries {
		if e.Name() == "x.json" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("x.json listed %d times, want 1", count)
	}
}

func TestNotExist(t *testing.T) {
	fsys := testFS(t)

	_, err := fs.ReadFile(fsys, "grammars/missing.json")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile missing: err = %v, want fs.ErrNotExist", err)
	}
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) && pathErr.Path != "grammars/missing.json" {
		t.Errorf("error path = %q, want requested name", pathErr.Path)
	}

	_, err = fsys.Open("grammars/missing.json")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Open missing: err = %v, want fs.ErrNotExist", err)
	}
}

func TestCorruptGzip(t *testing.T) {
	fsys := New(fstest.MapFS{
		"bad.json.gz": &fstest.MapFile{Data: []byte("not gzip at all")},
	})

	_, err := fsys.Open("bad.json")
	var pathErr *fs.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("Open corrupt: err = %v, want *fs.PathError", err)
	}
	if pathErr.Path != "bad.json" {
		t.Errorf("PathError path = %q, want bad.json", pathErr.Path)
	}

	if _, err := fs.ReadFile(fsys, "bad.json"); err == nil {
		t.Error("ReadFile corrupt: want error")
	}
}

func TestSub(t *testing.T) {
	fsys := testFS(t)
	sub, err := fs.Sub(fsys, "grammars")
	if err != nil {
		t.Fatalf("Sub: %v", err)
	}

	data, err := fs.ReadFile(sub, "go.json")
	if err != nil {
		t.Fatalf("Sub ReadFile: %v", err)
	}
	if string(data) != `{"scopeName":"source.go"}` {
		t.Errorf("Sub content = %q", data)
	}

	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		t.Fatalf("Sub ReadDir: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Sub ReadDir count = %d, want 2", len(entries))
	}
}

func TestFSConformance(t *testing.T) {
	if err := fstest.TestFS(testFS(t),
		"grammars/go.json",
		"grammars/index.json",
		"themes/dark.json",
		"plain.txt",
	); err != nil {
		t.Fatalf("fstest.TestFS: %v", err)
	}
}
