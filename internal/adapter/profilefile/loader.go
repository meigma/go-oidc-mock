// Package profilefile loads OIDC mock profiles from mounted JSON files.
package profilefile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/meigma/go-oidc-mock/internal/oidc"
)

type profileDocument struct {
	ID           string         `json:"id"`
	Label        string         `json:"label"`
	Subject      string         `json:"subject"`
	Claims       map[string]any `json:"claims"`
	CustomClaims map[string]any `json:"custom_claims"`
}

// Load reads non-recursive *.json profile files from dir.
func Load(dir string) ([]oidc.Profile, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read profile dir %q: %w", dir, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	slices.Sort(paths)

	profiles := make([]oidc.Profile, 0, len(paths))
	for _, path := range paths {
		profile, loadErr := loadProfile(path)
		if loadErr != nil {
			return nil, loadErr
		}
		profiles = append(profiles, profile)
	}

	normalized, err := oidc.NormalizeProfiles(profiles)
	if err != nil {
		return nil, fmt.Errorf("validate profiles in %q: %w", dir, err)
	}

	return normalized, nil
}

func loadProfile(path string) (oidc.Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return oidc.Profile{}, fmt.Errorf("open profile %q: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var doc profileDocument
	if err := decoder.Decode(&doc); err != nil {
		return oidc.Profile{}, fmt.Errorf("decode profile %q: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return oidc.Profile{}, fmt.Errorf("decode profile %q: trailing JSON content", path)
	}

	return oidc.Profile{
		ID:           doc.ID,
		Label:        doc.Label,
		Subject:      doc.Subject,
		Claims:       doc.Claims,
		CustomClaims: doc.CustomClaims,
	}, nil
}
