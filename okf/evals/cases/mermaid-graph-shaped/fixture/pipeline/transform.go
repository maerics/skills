package pipeline

// Record is the parsed form of a raw ingested file.
type Record struct {
	Fields map[string]string
}

// Transform parses raw bytes from Ingest into a Record. Its output feeds Load.
func Transform(raw []byte) (Record, error) {
	return Record{Fields: map[string]string{"raw": string(raw)}}, nil
}
