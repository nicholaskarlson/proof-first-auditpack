package auditpack

import (
	"fmt"
	"os"
	"path/filepath"
)

type SelfCheckOptions struct {
	Strict bool // if true, VerifyInput uses strict mode (fails on extra files)
	Keep   bool // if true, do not delete the temp dir; print where it is
}

// SelfCheck creates a tiny deterministic input tree, builds an audit pack, then verifies:
//  1. pack integrity via manifest.sha256 + manifest.json invariants
//  2. (optionally strict) input tree integrity against manifest.json
//
// Returns the temp root directory (useful if Keep==true).
func SelfCheck(opts SelfCheckOptions) (string, error) {
	root, err := os.MkdirTemp("", "proof-first-auditpack-selfcheck-*")
	if err != nil {
		return "", fmt.Errorf("mktemp: %w", err)
	}

	cleanup := func() {
		if !opts.Keep {
			_ = os.RemoveAll(root)
		}
	}

	inDir := filepath.Join(root, "input")
	outDir := filepath.Join(root, "pack")

	if err := os.MkdirAll(filepath.Join(inDir, "nested"), 0o755); err != nil {
		cleanup()
		return "", fmt.Errorf("mkdir input: %w", err)
	}

	// Deterministic content (stable across runs/platforms).
	if err := os.WriteFile(filepath.Join(inDir, "a.txt"), []byte("alpha\n"), 0o644); err != nil {
		cleanup()
		return "", fmt.Errorf("write a.txt: %w", err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "nested", "b.txt"), []byte("bravo\n"), 0o644); err != nil {
		cleanup()
		return "", fmt.Errorf("write b.txt: %w", err)
	}

	apOpts := DefaultOptions()
	apOpts.Version = "self-check"
	apOpts.InputLabel = "self-check-input"

	if err := Build(inDir, outDir, apOpts); err != nil {
		cleanup()
		return "", fmt.Errorf("build: %w", err)
	}

	if err := VerifyPack(outDir); err != nil {
		cleanup()
		return "", err
	}

	if err := VerifyInput(inDir, outDir, opts.Strict); err != nil {
		cleanup()
		return "", err
	}

	// Success. Keep or cleanup based on opts.
	if !opts.Keep {
		_ = os.RemoveAll(root)
		return "", nil
	}

	return root, nil
}
