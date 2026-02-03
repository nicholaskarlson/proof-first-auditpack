package auditpack

import (
	"encoding/json"
	"errors"
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

type Options struct {
	Tool    string
	Version string
	// InputLabel is written into manifest/meta instead of the raw inDir path.
	// Use this to keep outputs stable even if inDir is absolute.
	InputLabel string
}

func DefaultOptions() Options {
	return Options{
		Tool:       "proof-first-auditpack",
		Version:    "dev",
		InputLabel: "",
	}
}

func Build(inDir, outDir string, opts Options) error {
	if inDir == "" {
		return errors.New("inDir is required")
	}

	info, err := os.Stat(inDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("input is not a directory: %s", inDir)
	}

	label := opts.InputLabel
	if label == "" {
		label = inDir
	}

	entries := make([]manifest.FileEntry, 0, 64)
	var totalBytes int64

	walkErr := filepath.WalkDir(inDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		// Only regular files.
		if !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(inDir, p)
		if err != nil {
			return err
		}
		// Normalize to forward slashes for cross-platform stability.
		rel = filepath.ToSlash(rel)
		rel = path.Clean(rel)

		h, err := hashing.SHA256File(p)
		if err != nil {
			return err
		}

		entries = append(entries, manifest.FileEntry{
			Path:      rel,
			SizeBytes: h.SizeBytes,
			SHA256:    h.SHA256,
		})
		totalBytes += h.SizeBytes
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	if len(entries) == 0 {
		return fmt.Errorf("no files found under input directory: %s", inDir)
	}

	sort.Slice(entries, func(i, j int) bool {
		return strings.Compare(entries[i].Path, entries[j].Path) < 0
	})

	sum := manifest.Summary{
		FileCount:  len(entries),
		TotalBytes: totalBytes,
	}

	m := manifest.Manifest{
		Version: opts.Version,
		Input:   label,
		Files:   entries,
		Summary: sum,
	}

	meta := manifest.RunMeta{
		Tool:    opts.Tool,
		Version: opts.Version,
		Input:   label,
		Summary: sum,
	}

	manifestBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	manifestBytes = append(manifestBytes, '\n')

	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	metaBytes = append(metaBytes, '\n')

	if err := writeFileAtomic(outDir, "manifest.json", manifestBytes); err != nil {
		return err
	}
	if err := writeFileAtomic(outDir, "run_meta.json", metaBytes); err != nil {
		return err
	}

	// Write manifest.sha256 containing checksums for manifest.json and run_meta.json
	manHash, err := hashing.SHA256File(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		return err
	}
	metaHash, err := hashing.SHA256File(filepath.Join(outDir, "run_meta.json"))
	if err != nil {
		return err
	}

	lines := []string{
		fmt.Sprintf("%s  manifest.json", manHash.SHA256),
		fmt.Sprintf("%s  run_meta.json", metaHash.SHA256),
	}
	sort.Strings(lines)

	shaBytes := []byte(strings.Join(lines, "\n") + "\n")
	if err := writeFileAtomic(outDir, "manifest.sha256", shaBytes); err != nil {
		return err
	}

	return nil
}
