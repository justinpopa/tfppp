package artifacts

import (
	"encoding/json"
	"io"
	"os"
)

type Artifact struct {
	Name         string        `json:"name"`
	Path         string        `json:"path"`
	GoOS         string        `json:"goos"`
	GoArch       string        `json:"goarch"`
	InternalType int           `json:"internal_type"`
	Type         string        `json:"type"`
	Extra        ArtifactExtra `json:"extra"`
}

type ArtifactExtra struct {
	Binary   string   `json:"Binary"`
	Binaries []string `json:"Binaries"`
	Ext      string   `json:"Ext"`
	ID       string   `json:"ID"`
	Checksum string   `json:"Checksum"`
	Format   string   `json:"Format"`
}

// Load Artifacts from file into struct.
func Get(path string) (*[]Artifact, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	artifacts := make([]Artifact, 0)

	err = json.Unmarshal(contents, &artifacts)
	if err != nil {
		return nil, err
	}

	return &artifacts, nil
}
