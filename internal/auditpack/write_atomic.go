package auditpack

import (
	"os"
	"path/filepath"
)

func writeFileAtomic(outDir, name string, data []byte) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(outDir, name+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// Ensure cleanup on error.
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	finalPath := filepath.Join(outDir, name)
	if err := os.Rename(tmpName, finalPath); err != nil {
		return err
	}
	return os.Chmod(finalPath, 0o644)

}
