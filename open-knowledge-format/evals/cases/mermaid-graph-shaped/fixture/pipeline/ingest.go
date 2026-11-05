package pipeline

import "os"

// Ingest reads a raw file from disk. Its output feeds Transform.
func Ingest(path string) ([]byte, error) {
	return os.ReadFile(path)
}
