package profilefile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/adapter/profilefile"
)

func TestLoadReturnsNoProfilesForMissingOrEmptyDirectory(t *testing.T) {
	t.Parallel()

	t.Run("missing directory", func(t *testing.T) {
		t.Parallel()

		profiles, err := profilefile.Load(filepath.Join(t.TempDir(), "missing"))

		require.NoError(t, err)
		assert.Empty(t, profiles)
	})

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()

		profiles, err := profilefile.Load(t.TempDir())

		require.NoError(t, err)
		assert.Empty(t, profiles)
	})
}

func TestLoadReadsJSONProfilesInFilenameOrder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "b.json"), `{"id":"second","subject":"subject-second"}`)
	writeFile(t, filepath.Join(dir, "a.json"), `{"id":"first","subject":"subject-first","claims":{"name":"First"}}`)
	writeFile(t, filepath.Join(dir, "ignored.txt"), `{"id":"ignored","subject":"ignored"}`)
	require.NoError(t, os.Mkdir(filepath.Join(dir, "nested"), 0o700))
	writeFile(t, filepath.Join(dir, "nested", "c.json"), `{"id":"nested","subject":"nested"}`)

	profiles, err := profilefile.Load(dir)

	require.NoError(t, err)
	require.Len(t, profiles, 2)
	assert.Equal(t, "first", profiles[0].ID)
	assert.Equal(t, "first", profiles[0].Label)
	assert.Equal(t, "subject-first", profiles[0].Subject)
	assert.Equal(t, "First", profiles[0].Claims["name"])
	assert.Equal(t, "second", profiles[1].ID)
}

func TestLoadRejectsMalformedProfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		write     func(t *testing.T, dir string)
		wantError string
	}{
		{
			name: "invalid json",
			write: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "default.json"), `{`)
			},
			wantError: "decode profile",
		},
		{
			name: "unknown field",
			write: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "default.json"), `{"id":"default","subject":"subject","extra":true}`)
			},
			wantError: "unknown field",
		},
		{
			name: "missing id",
			write: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "default.json"), `{"subject":"subject"}`)
			},
			wantError: "id is required",
		},
		{
			name: "missing subject",
			write: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "default.json"), `{"id":"default"}`)
			},
			wantError: "subject is required",
		},
		{
			name: "claims must be object",
			write: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "default.json"), `{"id":"default","subject":"subject","claims":[]}`)
			},
			wantError: "cannot unmarshal array",
		},
		{
			name: "custom claims must be object",
			write: func(t *testing.T, dir string) {
				writeFile(
					t,
					filepath.Join(dir, "default.json"),
					`{"id":"default","subject":"subject","custom_claims":true}`,
				)
			},
			wantError: "cannot unmarshal bool",
		},
		{
			name: "duplicate ids",
			write: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a.json"), `{"id":"same","subject":"one"}`)
				writeFile(t, filepath.Join(dir, "b.json"), `{"id":"same","subject":"two"}`)
			},
			wantError: "duplicate id",
		},
		{
			name: "reserved standard claim",
			write: func(t *testing.T, dir string) {
				writeFile(
					t,
					filepath.Join(dir, "default.json"),
					`{"id":"default","subject":"subject","claims":{"sub":"other"}}`,
				)
			},
			wantError: "claims.sub is reserved",
		},
		{
			name: "reserved custom claim",
			write: func(t *testing.T, dir string) {
				writeFile(
					t,
					filepath.Join(dir, "default.json"),
					`{"id":"default","subject":"subject","custom_claims":{"iss":"issuer"}}`,
				)
			},
			wantError: "custom_claims.iss is reserved",
		},
		{
			name: "email verified must be boolean",
			write: func(t *testing.T, dir string) {
				writeFile(
					t,
					filepath.Join(dir, "default.json"),
					`{"id":"default","subject":"subject","claims":{"email_verified":"true"}}`,
				)
			},
			wantError: "email_verified must be boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			tt.write(t, dir)

			_, err := profilefile.Load(dir)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}
