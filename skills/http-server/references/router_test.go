package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRouter_SystemRoutes(t *testing.T) {
	t.Parallel()

	router := NewRouter(Deps{})

	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, path)
	}
}

func TestNewRouter_OptionalRoutesDisabledByDefault(t *testing.T) {
	t.Parallel()

	router := NewRouter(Deps{})

	for _, path := range []string{"/metrics", "/docs/"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code, path)
	}
}
