package oidc

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

const (
	authorizationPolicyID = "go-oidc-mock-authorization-page"
	authorizationPageName = "authorization.html"

	authorizationActionApprove = "approve"
	authorizationActionDeny    = "deny"
)

var errAuthorizationDenied = errors.New("authorization denied")
var errInvalidCustomClaims = errors.New("custom_claims must be a JSON object")

type authorizationPolicy struct {
	profiles       []Profile
	defaultProfile Profile
	tmpl           *template.Template
}

type authorizationPage struct {
	ActionURL     string
	ClientID      string
	ClientName    string
	RedirectURI   string
	Scopes        []string
	Profiles      []authorizationProfileOption
	Subject       string
	Name          string
	Email         string
	EmailVerified bool
	CustomClaims  string
	Error         string
}

type authorizationProfileOption struct {
	ID       string
	Label    string
	Selected bool
}

func newAuthorizationPolicy(profiles []Profile) goidc.AuthnPolicy {
	auth := authorizationPolicy{
		profiles:       authorizationProfiles(profiles),
		defaultProfile: SelectDefaultProfile(profiles),
		tmpl:           template.Must(template.New(authorizationPageName).Parse(authorizationPageHTML)),
	}

	return goidc.NewPolicy(
		authorizationPolicyID,
		func(_ *http.Request, session *goidc.AuthnSession, _ *goidc.Client) bool {
			if session.Store == nil {
				session.Store = map[string]any{}
			}

			return true
		},
		auth.authenticate,
	)
}

func authorizationProfiles(profiles []Profile) []Profile {
	if len(profiles) == 0 {
		return []Profile{BuiltInProfile()}
	}

	out := cloneProfiles(profiles)
	if profileByID(out, BuiltInProfile().ID).ID == "" {
		out = append(out, BuiltInProfile())
	}

	return out
}

func (p authorizationPolicy) authenticate(
	w http.ResponseWriter,
	r *http.Request,
	session *goidc.AuthnSession,
	client *goidc.Client,
) (goidc.Status, error) {
	switch r.PostFormValue("action") {
	case "":
		return p.render(w, session, client, authorizationFormFromProfile(p.defaultProfile), "")
	case authorizationActionApprove:
		profile, err := p.approvedProfile(r)
		if err != nil {
			return p.render(w, session, client, authorizationFormFromRequest(r), err.Error())
		}
		approveAuthorization(session, profile)

		return goidc.StatusSuccess, nil
	case authorizationActionDeny:
		return goidc.StatusFailure, errAuthorizationDenied
	default:
		return p.render(w, session, client, authorizationFormFromRequest(r), "unknown authorization action")
	}
}

func (p authorizationPolicy) approvedProfile(r *http.Request) (Profile, error) {
	form := authorizationFormFromRequest(r)
	customClaims, err := parseCustomClaims(form.CustomClaims)
	if err != nil {
		return Profile{}, err
	}

	base := p.defaultProfile
	if selected := profileByID(p.profiles, form.ProfileID); selected.ID != "" {
		base = selected
	}

	profile := Profile{
		ID:      base.ID,
		Label:   base.Label,
		Subject: form.Subject,
		Claims: map[string]any{
			goidc.ClaimName:          form.Name,
			goidc.ClaimEmail:         form.Email,
			goidc.ClaimEmailVerified: form.EmailVerified,
		},
		CustomClaims: customClaims,
	}
	normalized, err := NormalizeProfiles([]Profile{profile})
	if err != nil {
		return Profile{}, err
	}

	return normalized[0], nil
}

func (p authorizationPolicy) render(
	w http.ResponseWriter,
	session *goidc.AuthnSession,
	client *goidc.Client,
	form authorizationForm,
	message string,
) (goidc.Status, error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := p.tmpl.ExecuteTemplate(w, authorizationPageName, p.page(session, client, form, message)); err != nil {
		return goidc.StatusFailure, fmt.Errorf("render authorization page: %w", err)
	}

	return goidc.StatusPending, nil
}

func (p authorizationPolicy) page(
	session *goidc.AuthnSession,
	client *goidc.Client,
	form authorizationForm,
	message string,
) authorizationPage {
	profiles := make([]authorizationProfileOption, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, authorizationProfileOption{
			ID:       profile.ID,
			Label:    profile.Label,
			Selected: profile.ID == form.ProfileID,
		})
	}

	return authorizationPage{
		ActionURL:     AuthorizationPath + "/" + session.ID,
		ClientID:      client.ID,
		ClientName:    authorizationClientName(client),
		RedirectURI:   session.RedirectURI,
		Scopes:        strings.Fields(session.Scopes),
		Profiles:      profiles,
		Subject:       form.Subject,
		Name:          form.Name,
		Email:         form.Email,
		EmailVerified: form.EmailVerified,
		CustomClaims:  form.CustomClaims,
		Error:         message,
	}
}

func approveAuthorization(session *goidc.AuthnSession, profile Profile) {
	claims := profile.UserClaims()
	session.Subject = profile.Subject
	session.Username = profile.Subject
	session.GrantedScopes = session.Scopes
	session.GrantedResources = session.Resources
	session.GrantedAuthDetails = session.AuthDetails
	if session.Store == nil {
		session.Store = map[string]any{}
	}
	session.Store[idClaimsStoreKey] = claims
	session.Store[infoClaimsStoreKey] = claims
}

func parseCustomClaims(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}, nil
	}

	var claims map[string]any
	if err := json.Unmarshal([]byte(raw), &claims); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidCustomClaims, err)
	}
	if claims == nil {
		return nil, errInvalidCustomClaims
	}

	return claims, nil
}

func profileByID(profiles []Profile, id string) Profile {
	for _, profile := range profiles {
		if profile.ID == id {
			return profile
		}
	}

	return Profile{}
}

func authorizationClientName(client *goidc.Client) string {
	if client.Name != "" {
		return client.Name
	}

	return client.ID
}

type authorizationForm struct {
	ProfileID     string
	Subject       string
	Name          string
	Email         string
	EmailVerified bool
	CustomClaims  string
}

func authorizationFormFromProfile(profile Profile) authorizationForm {
	customClaims, err := json.MarshalIndent(profile.CustomClaims, "", "  ")
	if err != nil {
		customClaims = []byte("{}")
	}
	if len(customClaims) == 0 || string(customClaims) == "null" {
		customClaims = []byte("{}")
	}

	return authorizationForm{
		ProfileID:     profile.ID,
		Subject:       profile.Subject,
		Name:          stringClaim(profile.Claims, goidc.ClaimName),
		Email:         stringClaim(profile.Claims, goidc.ClaimEmail),
		EmailVerified: boolClaim(profile.Claims, goidc.ClaimEmailVerified),
		CustomClaims:  string(customClaims),
	}
}

func authorizationFormFromRequest(r *http.Request) authorizationForm {
	return authorizationForm{
		ProfileID:     strings.TrimSpace(r.PostFormValue("profile_id")),
		Subject:       strings.TrimSpace(r.PostFormValue("subject")),
		Name:          strings.TrimSpace(r.PostFormValue("name")),
		Email:         strings.TrimSpace(r.PostFormValue("email")),
		EmailVerified: r.PostFormValue("email_verified") == "true",
		CustomClaims:  strings.TrimSpace(r.PostFormValue("custom_claims")),
	}
}

func stringClaim(claims map[string]any, name string) string {
	value, ok := claims[name].(string)
	if !ok {
		return ""
	}

	return value
}

func boolClaim(claims map[string]any, name string) bool {
	value, ok := claims[name].(bool)
	if !ok {
		return false
	}

	return value
}

const authorizationPageHTML = `{{ define "authorization.html" }}
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Authorize {{ .ClientName }}</title>
  <style>
    :root {
      color-scheme: light;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #f6f7f9;
      color: #1f2933;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 32px 16px;
    }
    main {
      width: min(760px, 100%);
      background: #ffffff;
      border: 1px solid #d9e0e8;
      border-radius: 8px;
      box-shadow: 0 16px 48px rgba(15, 23, 42, 0.10);
      overflow: hidden;
    }
    header {
      padding: 24px 28px 16px;
      border-bottom: 1px solid #e5eaf0;
    }
    h1 {
      margin: 0 0 8px;
      font-size: 24px;
      line-height: 1.2;
      font-weight: 700;
      letter-spacing: 0;
    }
    .meta {
      display: grid;
      gap: 6px;
      color: #52616f;
      font-size: 14px;
      overflow-wrap: anywhere;
    }
    form {
      display: grid;
      gap: 18px;
      padding: 24px 28px 28px;
    }
    fieldset {
      border: 0;
      padding: 0;
      margin: 0;
      display: grid;
      gap: 12px;
    }
    legend {
      font-weight: 700;
      margin-bottom: 2px;
    }
    label {
      display: grid;
      gap: 6px;
      font-size: 14px;
      font-weight: 600;
    }
    input, select, textarea {
      width: 100%;
      border: 1px solid #c8d1dc;
      border-radius: 6px;
      padding: 10px 12px;
      color: #1f2933;
      font: inherit;
      font-weight: 400;
      background: #ffffff;
    }
    textarea {
      min-height: 132px;
      resize: vertical;
      font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
      font-size: 13px;
      line-height: 1.45;
    }
    .check {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 600;
    }
    .check input {
      width: 18px;
      height: 18px;
      margin: 0;
    }
    .scopes {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-top: 4px;
    }
    .scope {
      border: 1px solid #d9e0e8;
      border-radius: 999px;
      padding: 4px 10px;
      background: #f8fafc;
      color: #334e68;
      font-size: 13px;
    }
    .error {
      border: 1px solid #f5b5b8;
      border-radius: 6px;
      padding: 10px 12px;
      background: #fff1f2;
      color: #9f1239;
      font-size: 14px;
      font-weight: 600;
    }
    .actions {
      display: flex;
      gap: 10px;
      justify-content: flex-end;
      flex-wrap: wrap;
      border-top: 1px solid #e5eaf0;
      padding-top: 18px;
    }
    button {
      border: 1px solid #b8c4d0;
      border-radius: 6px;
      padding: 10px 16px;
      font: inherit;
      font-weight: 700;
      cursor: pointer;
      background: #ffffff;
      color: #1f2933;
    }
    button[value="approve"] {
      border-color: #2563eb;
      background: #2563eb;
      color: #ffffff;
    }
    @media (max-width: 560px) {
      body { padding: 0; place-items: stretch; }
      main { border-radius: 0; border-left: 0; border-right: 0; box-shadow: none; }
      header, form { padding-left: 18px; padding-right: 18px; }
      .actions { justify-content: stretch; }
      button { flex: 1; }
    }
  </style>
</head>
<body>
<main>
  <header>
    <h1>Authorize {{ .ClientName }}</h1>
    <div class="meta">
      <div>Client ID: {{ .ClientID }}</div>
      <div>Redirect URI: {{ .RedirectURI }}</div>
      <div>
        Requested scopes:
        <span class="scopes">{{ range .Scopes }}<span class="scope">{{ . }}</span>{{ else }}<span class="scope">none</span>{{ end }}</span>
      </div>
    </div>
  </header>
  <form method="post" action="{{ .ActionURL }}">
    {{ if .Error }}<div class="error">{{ .Error }}</div>{{ end }}
    <fieldset>
      <legend>User profile</legend>
      <label>
        Profile template
        <select name="profile_id">
          {{ range .Profiles }}
          <option value="{{ .ID }}"{{ if .Selected }} selected{{ end }}>{{ .Label }} ({{ .ID }})</option>
          {{ end }}
        </select>
      </label>
      <label>
        Subject
        <input name="subject" value="{{ .Subject }}" autocomplete="off" required>
      </label>
      <label>
        Name
        <input name="name" value="{{ .Name }}" autocomplete="off">
      </label>
      <label>
        Email
        <input name="email" type="email" value="{{ .Email }}" autocomplete="off">
      </label>
      <label class="check">
        <input name="email_verified" type="checkbox" value="true"{{ if .EmailVerified }} checked{{ end }}>
        Email verified
      </label>
      <label>
        Custom claims JSON
        <textarea name="custom_claims" spellcheck="false">{{ .CustomClaims }}</textarea>
      </label>
    </fieldset>
    <div class="actions">
      <button type="submit" name="action" value="deny">Deny</button>
      <button type="submit" name="action" value="approve">Approve</button>
    </div>
  </form>
</main>
</body>
</html>
{{ end }}
`
