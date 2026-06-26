// Package httpapi adapts the OIDC mock service to Huma operations.
package httpapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/meigma/go-oidc-mock/internal/oidc"
)

const tagOIDC = "OIDC"

type handlers struct {
	service *oidc.Service
}

type metadataOutput struct {
	Body oidc.ProviderMetadata
}

type jwksOutput struct {
	Body oidc.JWKS
}

// Register mounts the OIDC/OAuth protocol endpoints on api.
func Register(api huma.API, service *oidc.Service) {
	h := &handlers{service: service}

	huma.Register(api, huma.Operation{
		OperationID: "openid-configuration",
		Method:      http.MethodGet,
		Path:        oidc.DiscoveryPath,
		Summary:     "Get OpenID provider configuration",
		Tags:        []string{tagOIDC},
	}, h.discovery)

	huma.Register(api, huma.Operation{
		OperationID: "jwks",
		Method:      http.MethodGet,
		Path:        oidc.JWKSPath,
		Summary:     "Get JSON Web Key Set",
		Tags:        []string{tagOIDC},
	}, h.jwks)

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

func (h *handlers) discovery(_ context.Context, _ *struct{}) (*metadataOutput, error) {
	return &metadataOutput{Body: h.service.ProviderMetadata()}, nil
}

func (h *handlers) jwks(_ context.Context, _ *struct{}) (*jwksOutput, error) {
	return &jwksOutput{Body: h.service.JWKS()}, nil
}

func (h *handlers) notImplemented(_ context.Context, _ *struct{}) (*struct{}, error) {
	return nil, huma.Error501NotImplemented("protocol flow is not implemented in this first-pass shell")
}
