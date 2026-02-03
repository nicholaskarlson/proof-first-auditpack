package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type FileHash struct {
	SHA256    string
	SizeBytes int64
}

func SHA256File(path string) (FileHash, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileHash{}, err
	}
	defer f.Close()

	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return FileHash{}, err
	}

	sum := hex.EncodeToString(h.Sum(nil))
	return FileHash{SHA256: sum, SizeBytes: n}, nil
}
