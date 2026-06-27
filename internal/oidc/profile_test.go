package oidc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/oidc"
)

func TestSelectDefaultProfile(t *testing.T) {
	t.Parallel()

	t.Run("falls back to built-in mock user", func(t *testing.T) {
		t.Parallel()

		profile := oidc.SelectDefaultProfile(nil)

		assert.Equal(t, oidc.DefaultSubject, profile.Subject)
		assert.Equal(t, oidc.DefaultName, profile.Claims["name"])
		assert.Equal(t, oidc.DefaultEmail, profile.Claims["email"])
	})

	t.Run("prefers profile with default id", func(t *testing.T) {
		t.Parallel()

		profiles, err := oidc.NormalizeProfiles([]oidc.Profile{
			{ID: "alpha", Subject: "subject-alpha"},
			{ID: "default", Subject: "subject-default"},
		})
		require.NoError(t, err)

		profile := oidc.SelectDefaultProfile(profiles)

		assert.Equal(t, "default", profile.ID)
		assert.Equal(t, "subject-default", profile.Subject)
	})

	t.Run("uses first profile when default id is absent", func(t *testing.T) {
		t.Parallel()

		profiles, err := oidc.NormalizeProfiles([]oidc.Profile{
			{ID: "alpha", Subject: "subject-alpha"},
			{ID: "beta", Subject: "subject-beta"},
		})
		require.NoError(t, err)

		profile := oidc.SelectDefaultProfile(profiles)

		assert.Equal(t, "alpha", profile.ID)
		assert.Equal(t, "subject-alpha", profile.Subject)
	})
}

func TestNewServiceRejectsInvalidProfiles(t *testing.T) {
	t.Parallel()

	_, err := oidc.NewServiceWithOptions(
		"https://issuer.example.test",
		oidc.WithProfiles(oidc.Profile{ID: "default"}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "subject is required")
}
