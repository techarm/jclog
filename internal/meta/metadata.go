package meta

import (
	_ "embed"
	"encoding/json"
	"sync"
)

//go:embed metadata.json
var raw []byte

var (
	metadata     Metadata
	metadataOnce sync.Once
)

type Metadata struct {
	Version string `json:"version"`
}

func GetMetadata() *Metadata {
	metadataOnce.Do(func() {
		if err := json.Unmarshal(raw, &metadata); err != nil {
			panic(err)
		}
	})
	return &metadata
}
