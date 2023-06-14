package handler

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

type Metadata struct {
	Project     string        `json:"project_name"`
	Tag         string        `json:"tag"`
	PreviousTag string        `json:"previous_tag"`
	Version     string        `json:"version"`
	Commit      string        `json:"commit"`
	Date        time.Time     `json:"date"`
	Runtime     runtimeConfig `json:"runtime"`
}

type runtimeConfig struct {
	GoOS   string `json:"goos"`
	GoArch string `json:"goarch"`
}

// Load Metadata from file into struct.
func (h *Handler) GetMetadata(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	contents, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, &h.Metadata)
	if err != nil {
		return err
	}

	return nil
}
