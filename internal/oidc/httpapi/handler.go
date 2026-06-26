// Package httpapi adapts the OIDC mock service to Huma operations.
package httpapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/meigma/go-oidc-mock/internal/oidc"
)

const tagOIDC = "OIDC"

type handlers struct{}

// Register mounts the phase-2 OIDC/OAuth flow placeholders on api.
func Register(api huma.API, _ *oidc.Service) {
	h := &handlers{}

	huma.Register(api, huma.Operation{
		OperationID: "authorize",
		Method:      http.MethodGet,
		Path:        oidc.AuthorizationPath,
		Summary:     "Start OAuth authorization",
		Tags:        []string{tagOIDC},
		Errors:      []int{http.StatusNotImplemented},
	}, h.notImplemented)

	huma.Register(api, huma.Operation{
		OperationID: "token",
		Method:      http.MethodPost,
		Path:        oidc.TokenPath,
		Summary:     "Exchange OAuth token",
		Tags:        []string{tagOIDC},
		Errors:      []int{http.StatusNotImplemented},
	}, h.notImplemented)

	huma.Register(api, huma.Operation{
		OperationID: "userinfo",
		Method:      http.MethodGet,
		Path:        oidc.UserInfoPath,
		Summary:     "Get UserInfo claims",
		Tags:        []string{tagOIDC},
		Errors:      []int{http.StatusNotImplemented},
	}, h.notImplemented)
}

func (h *handlers) notImplemented(_ context.Context, _ *struct{}) (*struct{}, error) {
	return nil, huma.Error501NotImplemented("protocol flow is not implemented in this first-pass shell")
}
