package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/auditpack"
)

func TestVerifyPackAndInput_OK(t *testing.T) {
	t.Parallel()

	inDir := filepath.Join("..", "fixtures", "input", "case01")
	outDir := t.TempDir()

	opts := auditpack.DefaultOptions()
	opts.Version = "dev"
	opts.InputLabel = "fixtures/input/case01"

	if err := auditpack.Build(inDir, outDir, opts); err != nil {
		t.Fatalf("build: %v", err)
	}

	if err := auditpack.VerifyPack(outDir); err != nil {
		t.Fatalf("verify pack: %v", err)
	}

	if err := auditpack.VerifyInput(inDir, outDir, true); err != nil {
		t.Fatalf("verify input: %v", err)
	}
}

func TestVerifyPack_FailsOnTamper(t *testing.T) {
	t.Parallel()

	inDir := filepath.Join("..", "fixtures", "input", "case01")
	outDir := t.TempDir()

	opts := auditpack.DefaultOptions()
	opts.Version = "dev"
	opts.InputLabel = "fixtures/input/case01"

	if err := auditpack.Build(inDir, outDir, opts); err != nil {
		t.Fatalf("build: %v", err)
	}

	// Tamper with manifest.json after build (manifest.sha256 should catch it).
	manPath := filepath.Join(outDir, "manifest.json")
	if err := os.WriteFile(manPath, []byte("{\"tampered\":true}\n"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}

	if err := auditpack.VerifyPack(outDir); err == nil {
		t.Fatalf("expected verify failure after tamper, got nil")
	}
}
