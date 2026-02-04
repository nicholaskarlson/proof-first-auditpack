package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/manifest"
)

func buildAuditpackBinary(t *testing.T) (repoRoot string, binPath string) {
	t.Helper()

	// Package tests is in ./tests, so repo root is ..
	repoRoot = filepath.Clean(filepath.Join(".."))
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("abs repo root: %v", err)
	}
	repoRoot = absRoot

	binName := "auditpack"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath = filepath.Join(t.TempDir(), binName)

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/auditpack")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}
	return repoRoot, binPath
}

func runCmdOK(t *testing.T, bin string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\nargs=%v\n%s", err, args, string(out))
	}
}

func TestCLI_VerifyFlagAliasAndLabel(t *testing.T) {
	repoRoot, bin := buildAuditpackBinary(t)

	inDir := filepath.Join(repoRoot, "fixtures", "input", "case01")
	outDir := filepath.Join(t.TempDir(), "out")

	// Run with a stable label so manifest/meta are portable even if --in is absolute.
	runCmdOK(t, bin,
		"run",
		"--in", inDir,
		"--out", outDir,
		"--label", "fixtures/input/case01",
	)

	// Confirm label was written into manifest.json.
	manBytes, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest.json: %v", err)
	}
	var m manifest.Manifest
	if err := json.Unmarshal(manBytes, &m); err != nil {
		t.Fatalf("parse manifest.json: %v", err)
	}
	if m.Input != "fixtures/input/case01" {
		t.Fatalf("manifest input label mismatch: got %q", m.Input)
	}

	// Verify with the preferred flag.
	runCmdOK(t, bin, "verify", "--pack", outDir)
	// Verify with the legacy alias.
	runCmdOK(t, bin, "verify", "--out", outDir)
}
