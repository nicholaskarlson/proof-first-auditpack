package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/auditpack"
)

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func TestGoldenCase01(t *testing.T) {
	t.Parallel()

	inDir := filepath.Join("..", "fixtures", "input", "case01")
	outDir := t.TempDir()

	opts := auditpack.DefaultOptions()
	opts.Version = "dev"
	opts.InputLabel = "fixtures/input/case01"

	if err := auditpack.Build(inDir, outDir, opts); err != nil {
		t.Fatalf("build: %v", err)
	}

	gotManifest := mustRead(t, filepath.Join(outDir, "manifest.json"))
	gotMeta := mustRead(t, filepath.Join(outDir, "run_meta.json"))
	gotSHA := mustRead(t, filepath.Join(outDir, "manifest.sha256"))

	expDir := filepath.Join("..", "fixtures", "expected", "case01")
	expManifest := mustRead(t, filepath.Join(expDir, "manifest.json"))
	expMeta := mustRead(t, filepath.Join(expDir, "run_meta.json"))
	expSHA := mustRead(t, filepath.Join(expDir, "manifest.sha256"))

	if string(gotManifest) != string(expManifest) {
		t.Fatalf("manifest.json mismatch")
	}
	if string(gotMeta) != string(expMeta) {
		t.Fatalf("run_meta.json mismatch")
	}
	if string(gotSHA) != string(expSHA) {
		t.Fatalf("manifest.sha256 mismatch")
	}
}
