package manifest

type FileEntry struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type Summary struct {
	FileCount  int   `json:"file_count"`
	TotalBytes int64 `json:"total_bytes"`
}

type Manifest struct {
	Version string      `json:"version"`
	Input   string      `json:"input"`
	Files   []FileEntry `json:"files"`
	Summary Summary     `json:"summary"`
}

type RunMeta struct {
	Tool    string  `json:"tool"`
	Version string  `json:"version"`
	Input   string  `json:"input"`
	Summary Summary `json:"summary"`
}
