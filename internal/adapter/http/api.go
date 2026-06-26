// Package http assembles the generic, resource-agnostic HTTP transport: the chi
// router and middleware, the infrastructure routes (/healthz, /readyz, /metrics),
// the Huma API, and server-less OpenAPI export. Resource operations are mounted by
// their own adapter packages (for example, internal/oidc/httpapi) through
// the Registrar seam.
package http

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

// apiTitle is the OpenAPI document title for this service.
const apiTitle = "go-oidc-mock"

// Registrar mounts resource operations onto a Huma API. Each resource's HTTP
// adapter package provides one, and the composition root composes them.
type Registrar func(huma.API)

// RawRoute mounts a non-Huma HTTP handler on an exact router path.
type RawRoute struct {
	// Method is the HTTP method accepted by the route.
	Method string
	// Path is the exact route path.
	Path string
	// Handler serves the route.
	Handler http.Handler
}

// NewAPI wraps mux with Huma and returns the API. It registers no operations;
// callers register resource handlers onto the returned API via a Registrar.
func NewAPI(mux chi.Router, version string) huma.API {
	return humachi.New(mux, huma.DefaultConfig(apiTitle, version))
}

// SpecYAML builds the API on a throwaway router, applies register, and returns the
// OpenAPI 3.0.3 specification as YAML, without binding a network listener.
//
// finalize, when non-nil, runs after the operations are registered and before
// the document is serialized.
func SpecYAML(version string, register Registrar, finalize func(huma.API)) ([]byte, error) {
	api := NewAPI(chi.NewMux(), version)
	if register != nil {
		register(api)
	}
	if finalize != nil {
		finalize(api)
	}

	spec, err := api.OpenAPI().DowngradeYAML()
	if err != nil {
		return nil, fmt.Errorf("downgrade openapi spec to yaml: %w", err)
	}

	return spec, nil
}
