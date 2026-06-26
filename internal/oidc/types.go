package oidc

// ProviderMetadata is the OpenID Provider Configuration response.
type ProviderMetadata struct {
	// Issuer identifies the mock provider.
	Issuer string `json:"issuer"`
	// AuthorizationEndpoint is the endpoint clients redirect users to.
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	// TokenEndpoint is the endpoint clients exchange authorization grants with.
	TokenEndpoint string `json:"token_endpoint"`
	// JWKSURI is the endpoint clients use to fetch signing keys.
	JWKSURI string `json:"jwks_uri"`
	// UserInfoEndpoint is the endpoint clients use to fetch profile claims.
	UserInfoEndpoint string `json:"userinfo_endpoint"`
	// ScopesSupported lists the scopes this mock intends to support.
	ScopesSupported []string `json:"scopes_supported"`
	// ResponseTypesSupported lists the OAuth response types this mock intends to support.
	ResponseTypesSupported []string `json:"response_types_supported"`
	// GrantTypesSupported lists the OAuth grant types this mock intends to support.
	GrantTypesSupported []string `json:"grant_types_supported"`
	// SubjectTypesSupported lists the subject identifier types this mock supports.
	SubjectTypesSupported []string `json:"subject_types_supported"`
	// IDTokenSigningAlgValuesSupported lists supported ID token signing algorithms.
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	// TokenEndpointAuthMethodsSupported lists supported token endpoint client auth methods.
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	// ClaimsSupported lists the claims this mock intends to issue.
	ClaimsSupported []string `json:"claims_supported"`
	// CodeChallengeMethodsSupported lists supported PKCE code challenge methods.
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

// JWKS is a JSON Web Key Set response.
type JWKS struct {
	// Keys contains the configured JSON Web Keys.
	Keys []JWK `json:"keys"`
}

// JWK is a placeholder JSON Web Key shape for the first-pass empty JWKS.
type JWK map[string]any
