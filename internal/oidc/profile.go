package oidc

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

const defaultProfileID = "default"

// Profile is a reusable user template that can be snapshotted onto an
// authorization grant.
type Profile struct {
	// ID is the stable profile identifier used for selection.
	ID string
	// Label is the human-facing profile name. Empty labels default to ID.
	Label string
	// Subject is the OIDC subject claim used for approved grants.
	Subject string
	// Claims are standard editable OIDC claims included in ID token and userinfo responses.
	Claims map[string]any
	// CustomClaims are additional non-reserved claims included in ID token and userinfo responses.
	CustomClaims map[string]any
}

// BuiltInProfile returns the fixed mock user used when no mounted profiles are
// available.
func BuiltInProfile() Profile {
	return Profile{
		ID:      defaultProfileID,
		Label:   "Mock User",
		Subject: DefaultSubject,
		Claims: map[string]any{
			goidc.ClaimName:          DefaultName,
			goidc.ClaimEmail:         DefaultEmail,
			goidc.ClaimEmailVerified: true,
		},
	}
}

// NormalizeProfiles validates and clones profiles for use by the provider.
func NormalizeProfiles(profiles []Profile) ([]Profile, error) {
	if len(profiles) == 0 {
		return nil, nil
	}

	seen := map[string]struct{}{}
	out := make([]Profile, 0, len(profiles))
	for i, profile := range profiles {
		normalized := profile.clone()
		normalized.ID = strings.TrimSpace(normalized.ID)
		normalized.Label = strings.TrimSpace(normalized.Label)
		normalized.Subject = strings.TrimSpace(normalized.Subject)
		if normalized.Label == "" {
			normalized.Label = normalized.ID
		}

		if err := validateProfile(normalized); err != nil {
			return nil, fmt.Errorf("profile %d: %w", i+1, err)
		}
		if _, ok := seen[normalized.ID]; ok {
			return nil, fmt.Errorf("profile %q: duplicate id", normalized.ID)
		}
		seen[normalized.ID] = struct{}{}
		out = append(out, normalized)
	}

	return out, nil
}

// SelectDefaultProfile returns the preferred startup profile.
func SelectDefaultProfile(profiles []Profile) Profile {
	for _, profile := range profiles {
		if profile.ID == defaultProfileID {
			return profile.clone()
		}
	}
	if len(profiles) > 0 {
		return profiles[0].clone()
	}

	return BuiltInProfile()
}

// UserClaims returns the profile claims that should be snapshotted onto a grant.
func (p Profile) UserClaims() map[string]any {
	claims := make(map[string]any, len(p.Claims)+len(p.CustomClaims))
	maps.Copy(claims, p.Claims)
	maps.Copy(claims, p.CustomClaims)

	return claims
}

func validateProfile(profile Profile) error {
	if profile.ID == "" {
		return errors.New("id is required")
	}
	if profile.Subject == "" {
		return errors.New("subject is required")
	}
	if err := validateClaimMap("claims", profile.Claims); err != nil {
		return err
	}
	if err := validateClaimMap("custom_claims", profile.CustomClaims); err != nil {
		return err
	}

	return nil
}

func validateClaimMap(field string, claims map[string]any) error {
	for name, value := range claims {
		if isReservedProfileClaim(name) {
			return fmt.Errorf("%s.%s is reserved", field, name)
		}
		if name == goidc.ClaimEmailVerified {
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("%s.%s must be boolean", field, name)
			}
		}
	}

	return nil
}

func isReservedProfileClaim(name string) bool {
	return slices.Contains(reservedProfileClaims(), name)
}

func reservedProfileClaims() []string {
	return []string{
		goidc.ClaimSubject,
		goidc.ClaimIssuer,
		goidc.ClaimAudience,
		goidc.ClaimExpiry,
		goidc.ClaimIssuedAt,
		goidc.ClaimNotBefore,
		goidc.ClaimTokenID,
		goidc.ClaimNonce,
		"azp",
		"at_hash",
		"c_hash",
	}
}

func (p Profile) clone() Profile {
	return Profile{
		ID:           p.ID,
		Label:        p.Label,
		Subject:      p.Subject,
		Claims:       cloneClaimMap(p.Claims),
		CustomClaims: cloneClaimMap(p.CustomClaims),
	}
}

func cloneClaimMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}

	out := make(map[string]any, len(in))
	maps.Copy(out, in)

	return out
}
