package handler

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

// Load Metadata from file into struct.
func (h *Handler) GetArtifacts(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	contents, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, &h.Artifacts)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetShaSums() *string {
	for _, a := range *h.Artifacts {
		if a.Type == "Checksum" {
			return &a.Path
		}
	}
	return nil
}

func (h *Handler) GetSumsSig() *string {
	for _, a := range *h.Artifacts {
		if a.Type == "Signature" {
			return &a.Path
		}
	}
	return nil
}
