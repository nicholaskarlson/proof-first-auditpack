package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/auditpack"
	"github.com/nicholaskarlson/proof-first-auditpack/internal/manifest"
)

func mustWrite(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

func TestBuildExcludesOutDirWhenInsideInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	inDir := filepath.Join(root, "in")
	outDir := filepath.Join(inDir, "out")

	// Real input file.
	mustWrite(t, filepath.Join(inDir, "a.txt"), []byte("hello\n"))

	// Stale pack junk under outDir (the footgun we want to eliminate).
	mustWrite(t, filepath.Join(outDir, "old.txt"), []byte("stale"))
	mustWrite(t, filepath.Join(outDir, "sub", "older.txt"), []byte("staler"))

	opts := auditpack.DefaultOptions()
	opts.Version = "dev"
	opts.InputLabel = "test/input"

	if err := auditpack.Build(inDir, outDir, opts); err != nil {
		t.Fatalf("build: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest.json: %v", err)
	}

	var m manifest.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal manifest.json: %v", err)
	}

	if m.Summary.FileCount != 1 {
		t.Fatalf("expected 1 file in manifest, got %d", m.Summary.FileCount)
	}
	if len(m.Files) != 1 {
		t.Fatalf("expected 1 file entry, got %d", len(m.Files))
	}
	if m.Files[0].Path != "a.txt" {
		t.Fatalf("expected only a.txt, got %q", m.Files[0].Path)
	}
}
