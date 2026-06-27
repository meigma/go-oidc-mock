package app_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/go-oidc-mock/internal/app"
	"github.com/meigma/go-oidc-mock/internal/config"
	"github.com/meigma/go-oidc-mock/internal/observability"
	"github.com/meigma/go-oidc-mock/internal/oidc"
)

const (
	testIssuer       = "https://issuer.example.test"
	testPKCEVerifier = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
)

type authorizeRequest struct {
	clientID      string
	state         string
	codeVerifier  string
	redirectURI   string
	requestScopes string
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scopes       string `json:"scope"`
}

type authorizationApproval struct {
	ProfileID     string
	Subject       string
	Name          string
	Email         string
	EmailVerified bool
	CustomClaims  string
}

func TestPublicPKCEAuthorizationCodeFlow(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-public",
		codeVerifier:  testPKCEVerifier,
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid profile email",
	}

	code := authorizeCode(t, handler, authReq)
	token := exchangePublicCode(t, handler, code, authReq)

	assert.Equal(t, "Bearer", token.TokenType)
	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.IDToken)
	assert.Empty(t, token.RefreshToken)
	assert.Positive(t, token.ExpiresIn)

	userInfo := userInfo(t, handler, token.AccessToken)
	assert.Equal(t, oidc.DefaultSubject, userInfo["sub"])
	assert.Equal(t, oidc.DefaultName, userInfo["name"])
	assert.Equal(t, oidc.DefaultEmail, userInfo["email"])
	assert.Equal(t, true, userInfo["email_verified"])
}

func TestMountedProfileAuthorizationCodeFlow(t *testing.T) {
	t.Parallel()

	profileDir := t.TempDir()
	writeProfile(t, profileDir, "default.json", `{
		"id": "default",
		"label": "Mounted User",
		"subject": "mounted-user",
		"claims": {
			"name": "Mounted User",
			"email": "mounted@example.test",
			"email_verified": true
		},
		"custom_claims": {
			"roles": ["tester", "admin"]
		}
	}`)
	handler := newFlowHandlerWithProfileDir(t, profileDir)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-mounted-profile",
		codeVerifier:  testPKCEVerifier,
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid profile email",
	}

	page := requestAuthorizationPage(t, handler, authReq)
	assert.Contains(t, page, "Mounted User (default)")
	code := approveAuthorizationPage(t, handler, page, authReq, authorizationApproval{
		ProfileID:     "default",
		Subject:       "mounted-user",
		Name:          "Mounted User",
		Email:         "mounted@example.test",
		EmailVerified: true,
		CustomClaims:  `{"roles":["tester","admin"]}`,
	})
	token := exchangePublicCode(t, handler, code, authReq)

	userInfo := userInfo(t, handler, token.AccessToken)
	assert.Equal(t, "mounted-user", userInfo["sub"])
	assert.Equal(t, "Mounted User", userInfo["name"])
	assert.Equal(t, "mounted@example.test", userInfo["email"])
	assert.Equal(t, true, userInfo["email_verified"])
	assert.Equal(t, []any{"tester", "admin"}, userInfo["roles"])

	idTokenClaims := jwtClaims(t, token.IDToken)
	assert.Equal(t, "mounted-user", idTokenClaims["sub"])
	assert.Equal(t, "Mounted User", idTokenClaims["name"])
	assert.Equal(t, "mounted@example.test", idTokenClaims["email"])
	assert.Equal(t, true, idTokenClaims["email_verified"])
	assert.Equal(t, []any{"tester", "admin"}, idTokenClaims["roles"])
}

func TestAuthorizationPagePrefillsFirstMountedProfileWithoutDefault(t *testing.T) {
	t.Parallel()

	profileDir := t.TempDir()
	writeProfile(t, profileDir, "alpha.json", `{
		"id": "alpha",
		"label": "Alpha User",
		"subject": "alpha-user",
		"claims": {
			"name": "Alpha User",
			"email": "alpha@example.test",
			"email_verified": true
		}
	}`)
	handler := newFlowHandlerWithProfileDir(t, profileDir)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-alpha-profile",
		codeVerifier:  testPKCEVerifier,
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid profile email",
	}

	page := requestAuthorizationPage(t, handler, authReq)

	assert.Contains(t, page, `value="alpha" selected`)
	assert.Contains(t, page, `value="alpha-user"`)
	assert.Contains(t, page, `value="Alpha User"`)
	assert.Contains(t, page, `value="alpha@example.test"`)
}

func TestAuthorizationPageAllowsEditedClaims(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-edited-claims",
		codeVerifier:  testPKCEVerifier,
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid profile email",
	}
	approval := authorizationApproval{
		ProfileID:     "default",
		Subject:       "edited-user",
		Name:          "Edited User",
		Email:         "edited@example.test",
		EmailVerified: false,
		CustomClaims:  `{"roles":["operator"],"tenant":"alpha"}`,
	}

	code := authorizeCodeWithApproval(t, handler, authReq, approval)
	token := exchangePublicCode(t, handler, code, authReq)

	userInfo := userInfo(t, handler, token.AccessToken)
	assert.Equal(t, "edited-user", userInfo["sub"])
	assert.Equal(t, "Edited User", userInfo["name"])
	assert.Equal(t, "edited@example.test", userInfo["email"])
	assert.Equal(t, false, userInfo["email_verified"])
	assert.Equal(t, []any{"operator"}, userInfo["roles"])
	assert.Equal(t, "alpha", userInfo["tenant"])

	idTokenClaims := jwtClaims(t, token.IDToken)
	assert.Equal(t, "edited-user", idTokenClaims["sub"])
	assert.Equal(t, "Edited User", idTokenClaims["name"])
	assert.Equal(t, "edited@example.test", idTokenClaims["email"])
	assert.Equal(t, false, idTokenClaims["email_verified"])
	assert.Equal(t, []any{"operator"}, idTokenClaims["roles"])
	assert.Equal(t, "alpha", idTokenClaims["tenant"])
}

func TestPublicClientRequiresPKCE(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-missing-pkce",
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid",
	}

	rec := requestAuthorization(t, handler, authReq)
	require.Equal(t, http.StatusSeeOther, rec.Code)

	redirectURL, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)

	assert.Equal(t, "state-missing-pkce", redirectURL.Query().Get("state"))
	assert.Equal(t, "invalid_request", redirectURL.Query().Get("error"))
	assert.Empty(t, redirectURL.Query().Get("code"))
}

func TestConfidentialClientSecretBasicAuthorizationCodeFlow(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultConfidentialClientID,
		state:         "state-confidential",
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid profile email",
	}

	code := authorizeCode(t, handler, authReq)
	token := exchangeConfidentialCode(t, handler, code, authReq)

	assert.Equal(t, "Bearer", token.TokenType)
	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.IDToken)
}

func TestAuthorizationCodeReuseFails(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-reuse",
		codeVerifier:  testPKCEVerifier,
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid",
	}

	code := authorizeCode(t, handler, authReq)
	exchangePublicCode(t, handler, code, authReq)

	form := tokenForm(code, authReq.redirectURI)
	form.Set("client_id", authReq.clientID)
	form.Set("code_verifier", authReq.codeVerifier)
	rec := postForm(t, handler, oidc.TokenPath, form, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_grant")
}

func TestAuthorizationPageDenialRedirectsProtocolError(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	authReq := authorizeRequest{
		clientID:      oidc.DefaultPublicClientID,
		state:         "state-denied",
		codeVerifier:  testPKCEVerifier,
		redirectURI:   oidc.DefaultRedirectURI,
		requestScopes: "openid",
	}

	page := requestAuthorizationPage(t, handler, authReq)
	rec := postForm(t, handler, authorizationActionURL(t, page), url.Values{
		"action": {"deny"},
	}, nil)
	require.Equal(t, http.StatusSeeOther, rec.Code)

	redirectURL, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)

	assert.Equal(t, authReq.state, redirectURL.Query().Get("state"))
	assert.Equal(t, "access_denied", redirectURL.Query().Get("error"))
	assert.Empty(t, redirectURL.Query().Get("code"))
}

func TestAuthorizationPageValidationRerenders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mutate      func(url.Values)
		wantMessage string
	}{
		{
			name: "empty subject",
			mutate: func(form url.Values) {
				form.Set("subject", "")
			},
			wantMessage: "subject is required",
		},
		{
			name: "malformed custom claims",
			mutate: func(form url.Values) {
				form.Set("custom_claims", `{"roles":`)
			},
			wantMessage: "custom_claims must be a JSON object",
		},
		{
			name: "reserved custom claim",
			mutate: func(form url.Values) {
				form.Set("custom_claims", `{"sub":"override"}`)
			},
			wantMessage: "custom_claims.sub is reserved",
		},
		{
			name: "null custom claims",
			mutate: func(form url.Values) {
				form.Set("custom_claims", `null`)
			},
			wantMessage: "custom_claims must be a JSON object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := newFlowHandler(t)
			authReq := authorizeRequest{
				clientID:      oidc.DefaultPublicClientID,
				state:         "state-validation-" + strings.ReplaceAll(tt.name, " ", "-"),
				codeVerifier:  testPKCEVerifier,
				redirectURI:   oidc.DefaultRedirectURI,
				requestScopes: "openid profile email",
			}
			page := requestAuthorizationPage(t, handler, authReq)
			form := authorizationApprovalForm(defaultAuthorizationApproval())
			tt.mutate(form)

			rec := postForm(t, handler, authorizationActionURL(t, page), form, nil)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
			assert.Contains(t, rec.Body.String(), tt.wantMessage)
		})
	}
}

func TestAuthorizationCallbackRouteIsProviderOwned(t *testing.T) {
	t.Parallel()

	handler := newFlowHandler(t)
	req := httptest.NewRequest(http.MethodGet, oidc.AuthorizationPath+"/missing-session", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid_request")
}

func newFlowHandler(t *testing.T) http.Handler {
	t.Helper()

	return newFlowHandlerWithProfileDir(t, t.TempDir())
}

func newFlowHandlerWithProfileDir(t *testing.T, profileDir string) http.Handler {
	t.Helper()

	vp := viper.New()
	vp.Set("issuer-url", testIssuer)
	vp.Set("profile-dir", profileDir)
	vp.Set("rate-limit-enabled", false)
	cfg := config.Load(vp)
	logger := observability.NewLogger(io.Discard, slog.LevelError, "json")

	application, err := app.New(context.Background(), cfg, logger, "test")
	require.NoError(t, err)

	return application.Handler()
}

func authorizeCode(t *testing.T, handler http.Handler, authReq authorizeRequest) string {
	t.Helper()

	return authorizeCodeWithApproval(t, handler, authReq, defaultAuthorizationApproval())
}

func authorizeCodeWithApproval(
	t *testing.T,
	handler http.Handler,
	authReq authorizeRequest,
	approval authorizationApproval,
) string {
	t.Helper()

	page := requestAuthorizationPage(t, handler, authReq)
	return approveAuthorizationPage(t, handler, page, authReq, approval)
}

func approveAuthorizationPage(
	t *testing.T,
	handler http.Handler,
	page string,
	authReq authorizeRequest,
	approval authorizationApproval,
) string {
	t.Helper()

	rec := postForm(t, handler, authorizationActionURL(t, page), authorizationApprovalForm(approval), nil)
	require.Equal(t, http.StatusSeeOther, rec.Code, rec.Body.String())

	redirectURL, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)

	assert.Equal(t, authReq.state, redirectURL.Query().Get("state"))
	require.Empty(t, redirectURL.Query().Get("error"))
	code := redirectURL.Query().Get("code")
	require.NotEmpty(t, code)

	return code
}

func requestAuthorizationPage(t *testing.T, handler http.Handler, authReq authorizeRequest) string {
	t.Helper()

	rec := requestAuthorization(t, handler, authReq)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")

	body := rec.Body.String()
	assert.Contains(t, body, "Authorize")
	assert.Contains(t, body, authReq.clientID)
	assert.Contains(t, body, authReq.redirectURI)

	return body
}

func authorizationActionURL(t *testing.T, page string) string {
	t.Helper()

	matches := regexp.MustCompile(`action="([^"]+)"`).FindStringSubmatch(page)
	require.Len(t, matches, 2)

	return matches[1]
}

func defaultAuthorizationApproval() authorizationApproval {
	return authorizationApproval{
		ProfileID:     "default",
		Subject:       oidc.DefaultSubject,
		Name:          oidc.DefaultName,
		Email:         oidc.DefaultEmail,
		EmailVerified: true,
		CustomClaims:  `{}`,
	}
}

func authorizationApprovalForm(approval authorizationApproval) url.Values {
	form := url.Values{
		"action":        {"approve"},
		"profile_id":    {approval.ProfileID},
		"subject":       {approval.Subject},
		"name":          {approval.Name},
		"email":         {approval.Email},
		"custom_claims": {approval.CustomClaims},
	}
	if approval.EmailVerified {
		form.Set("email_verified", "true")
	}

	return form
}

func requestAuthorization(t *testing.T, handler http.Handler, authReq authorizeRequest) *httptest.ResponseRecorder {
	t.Helper()

	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", authReq.clientID)
	query.Set("redirect_uri", authReq.redirectURI)
	query.Set("scope", authReq.requestScopes)
	query.Set("state", authReq.state)
	query.Set("nonce", "nonce-"+authReq.state)
	if authReq.codeVerifier != "" {
		query.Set("code_challenge", codeChallenge(authReq.codeVerifier))
		query.Set("code_challenge_method", "S256")
	}

	req := httptest.NewRequest(http.MethodGet, oidc.AuthorizationPath+"?"+query.Encode(), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func exchangePublicCode(
	t *testing.T,
	handler http.Handler,
	code string,
	authReq authorizeRequest,
) tokenResponse {
	t.Helper()

	form := tokenForm(code, authReq.redirectURI)
	form.Set("client_id", authReq.clientID)
	form.Set("code_verifier", authReq.codeVerifier)
	rec := postForm(t, handler, oidc.TokenPath, form, nil)

	return decodeTokenResponse(t, rec, http.StatusOK)
}

func exchangeConfidentialCode(
	t *testing.T,
	handler http.Handler,
	code string,
	authReq authorizeRequest,
) tokenResponse {
	t.Helper()

	form := tokenForm(code, authReq.redirectURI)
	rec := postForm(t, handler, oidc.TokenPath, form, func(req *http.Request) {
		req.SetBasicAuth(authReq.clientID, oidc.DefaultConfidentialClientSecret)
	})

	return decodeTokenResponse(t, rec, http.StatusOK)
}

func tokenForm(code, redirectURI string) url.Values {
	return url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}
}

func postForm(
	t *testing.T,
	handler http.Handler,
	path string,
	form url.Values,
	mutate func(*http.Request),
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if mutate != nil {
		mutate(req)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func decodeTokenResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) tokenResponse {
	t.Helper()

	require.Equal(t, wantStatus, rec.Code, rec.Body.String())

	var token tokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &token))

	return token
}

func userInfo(t *testing.T, handler http.Handler, accessToken string) map[string]any {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, oidc.UserInfoPath, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var claims map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &claims))

	return claims
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))

	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func jwtClaims(t *testing.T, token string) map[string]any {
	t.Helper()

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))

	return claims
}

func writeProfile(t *testing.T, dir, name, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}
