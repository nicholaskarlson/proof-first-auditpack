package auditpack

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/hashing"
	"github.com/nicholaskarlson/proof-first-auditpack/internal/manifest"
)

func VerifyPack(outDir string) error {
	shaPath := filepath.Join(outDir, "manifest.sha256")
	lines, err := readLines(shaPath)
	if err != nil {
		return fmt.Errorf("read manifest.sha256: %w", err)
	}
	if len(lines) == 0 {
		return fmt.Errorf("manifest.sha256 is empty: %s", shaPath)
	}

	// Parse expected checksums.
	type exp struct {
		hash string
		file string
	}
	exps := make([]exp, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		fields := strings.Fields(ln)
		if len(fields) < 2 {
			return fmt.Errorf("invalid manifest.sha256 line: %q", ln)
		}
		h := fields[0]
		f := fields[1]
		if !isSHA256Hex(h) {
			return fmt.Errorf("invalid sha256: %q", h)
		}
		if strings.Contains(f, "..") || strings.Contains(f, "\\") || strings.HasPrefix(f, "/") {
			return fmt.Errorf("invalid filename in manifest.sha256: %q", f)
		}
		exps = append(exps, exp{hash: h, file: f})
	}

	if len(exps) == 0 {
		return fmt.Errorf("no checksum entries found in manifest.sha256")
	}

	// Verify each referenced file.
	for _, e := range exps {
		p := filepath.Join(outDir, e.file)
		got, err := hashing.SHA256File(p)
		if err != nil {
			return fmt.Errorf("hash %s: %w", e.file, err)
		}
		if got.SHA256 != e.hash {
			return fmt.Errorf("sha256 mismatch for %s: expected %s got %s", e.file, e.hash, got.SHA256)
		}
	}

	// Also validate manifest.json internal consistency.
	_, err = VerifyManifestSummary(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		return err
	}

	return nil
}

func VerifyInput(inDir, outDir string, strict bool) error {
	manPath := filepath.Join(outDir, "manifest.json")
	b, err := os.ReadFile(manPath)
	if err != nil {
		return fmt.Errorf("read manifest.json: %w", err)
	}

	var m manifest.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("parse manifest.json: %w", err)
	}

	if len(m.Files) == 0 {
		return fmt.Errorf("manifest.json has no files")
	}

	// Validate paths are unique + clean, and verify each file.
	expected := make(map[string]manifest.FileEntry, len(m.Files))
	paths := make([]string, 0, len(m.Files))

	var totalBytes int64
	for _, fe := range m.Files {
		if err := validateRelPath(fe.Path); err != nil {
			return fmt.Errorf("manifest path invalid (%q): %w", fe.Path, err)
		}
		if !isSHA256Hex(fe.SHA256) {
			return fmt.Errorf("manifest sha256 invalid for %q: %q", fe.Path, fe.SHA256)
		}
		if fe.SizeBytes < 0 {
			return fmt.Errorf("manifest size invalid for %q: %d", fe.Path, fe.SizeBytes)
		}
		if _, ok := expected[fe.Path]; ok {
			return fmt.Errorf("duplicate manifest path: %q", fe.Path)
		}
		expected[fe.Path] = fe
		paths = append(paths, fe.Path)
		totalBytes += fe.SizeBytes
	}

	// Determinism invariant: paths should be sorted.
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	for i := range paths {
		if paths[i] != sorted[i] {
			return fmt.Errorf("manifest files are not sorted by path (determinism invariant)")
		}
	}

	// Summary invariant.
	if m.Summary.FileCount != len(m.Files) {
		return fmt.Errorf("summary.file_count mismatch: expected %d got %d", len(m.Files), m.Summary.FileCount)
	}
	if m.Summary.TotalBytes != totalBytes {
		return fmt.Errorf("summary.total_bytes mismatch: expected %d got %d", totalBytes, m.Summary.TotalBytes)
	}

	// Verify actual input tree matches manifest entries.
	for _, p := range sorted {
		fe := expected[p]
		full := filepath.Join(inDir, filepath.FromSlash(p))

		info, err := os.Stat(full)
		if err != nil {
			return fmt.Errorf("input missing %q: %w", p, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("input not a regular file %q", p)
		}

		h, err := hashing.SHA256File(full)
		if err != nil {
			return fmt.Errorf("hash input %q: %w", p, err)
		}
		if h.SHA256 != fe.SHA256 {
			return fmt.Errorf("input sha256 mismatch for %q: expected %s got %s", p, fe.SHA256, h.SHA256)
		}
		if h.SizeBytes != fe.SizeBytes {
			return fmt.Errorf("input size mismatch for %q: expected %d got %d", p, fe.SizeBytes, h.SizeBytes)
		}
	}

	if strict {
		actual, err := walkInputRegularFiles(inDir)
		if err != nil {
			return err
		}
		for ap := range actual {
			if _, ok := expected[ap]; !ok {
				return fmt.Errorf("strict: extra input file not in manifest: %q", ap)
			}
		}
	}

	return nil
}

// VerifyManifestSummary validates manifest.json internal invariants (paths sorted/unique, summary counts/totals).
func VerifyManifestSummary(manifestPath string) (manifest.Manifest, error) {
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read manifest.json: %w", err)
	}
	var m manifest.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return manifest.Manifest{}, fmt.Errorf("parse manifest.json: %w", err)
	}
	if len(m.Files) == 0 {
		return manifest.Manifest{}, fmt.Errorf("manifest.json has no files")
	}

	paths := make([]string, 0, len(m.Files))
	seen := map[string]bool{}
	var totalBytes int64

	for _, fe := range m.Files {
		if err := validateRelPath(fe.Path); err != nil {
			return manifest.Manifest{}, fmt.Errorf("manifest path invalid (%q): %w", fe.Path, err)
		}
		if seen[fe.Path] {
			return manifest.Manifest{}, fmt.Errorf("duplicate manifest path: %q", fe.Path)
		}
		seen[fe.Path] = true
		paths = append(paths, fe.Path)
		totalBytes += fe.SizeBytes
	}

	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	for i := range paths {
		if paths[i] != sorted[i] {
			return manifest.Manifest{}, fmt.Errorf("manifest files are not sorted by path (determinism invariant)")
		}
	}

	if m.Summary.FileCount != len(m.Files) {
		return manifest.Manifest{}, fmt.Errorf("summary.file_count mismatch: expected %d got %d", len(m.Files), m.Summary.FileCount)
	}
	if m.Summary.TotalBytes != totalBytes {
		return manifest.Manifest{}, fmt.Errorf("summary.total_bytes mismatch: expected %d got %d", totalBytes, m.Summary.TotalBytes)
	}

	return m, nil
}

func walkInputRegularFiles(inDir string) (map[string]struct{}, error) {
	out := make(map[string]struct{}, 64)
	err := filepath.WalkDir(inDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(inDir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		rel = path.Clean(rel)
		if err := validateRelPath(rel); err != nil {
			return fmt.Errorf("input path invalid (%q): %w", rel, err)
		}
		out[rel] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func validateRelPath(p string) error {
	if p == "" {
		return fmt.Errorf("empty path")
	}
	if strings.Contains(p, "\\") {
		return fmt.Errorf("must use forward slashes")
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("must be relative")
	}
	clean := path.Clean(p)
	if clean != p {
		return fmt.Errorf("not clean: expected %q got %q", clean, p)
	}
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("path traversal not allowed")
	}
	return nil
}

func isSHA256Hex(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
