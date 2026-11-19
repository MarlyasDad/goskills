package middleware

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))

	h := Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		}),
		RequestID(),
		Logging(log),
		Recovery(log),
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, buf.String(), "panic recovered")
	require.Contains(t, buf.String(), "http request")
}

func TestRequestID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		in     string
		want   string
		strict bool
	}{
		{name: "keeps valid", in: "req-123", want: "req-123", strict: true},
		{name: "replaces invalid", in: "bad value", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got string
			h := Chain(
				http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					got = RequestIDFromContext(r.Context())
				}),
				RequestID(),
			)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(requestIDHeader, tt.in)

			h.ServeHTTP(rec, req)

			require.NotEmpty(t, got)
			if tt.strict {
				require.Equal(t, tt.want, got)
			}
			require.Equal(t, got, rec.Header().Get(requestIDHeader))
		})
	}
}

func TestCORS_Preflight(t *testing.T) {
	t.Parallel()

	h := Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("handler must not be called for preflight")
		}),
		CORS(CORSConfig{AllowedOrigins: []string{"https://example.com"}}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")

	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, rec.Header().Values("Vary"), "Origin")
}

type fakeAuthenticator struct {
	principal Principal
	err       error
}

func (f fakeAuthenticator) Authenticate(context.Context, *http.Request) (Principal, error) {
	return f.principal, f.err
}

func TestAuth(t *testing.T) {
	t.Parallel()

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		h := Chain(
			http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Fatal("handler must not be called")
			}),
			Auth(testLogger(), fakeAuthenticator{err: errors.New("invalid token")}),
		)

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.NotEmpty(t, rec.Header().Get("WWW-Authenticate"))
		require.JSONEq(t, `{"errors":[{"message":"unauthorized","code":0}]}`, rec.Body.String())
	})

	t.Run("principal in context", func(t *testing.T) {
		t.Parallel()

		want := Principal{Subject: "user-1"}
		h := Chain(
			http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got, ok := PrincipalFromContext(r.Context())
				require.True(t, ok)
				require.Equal(t, want, got)
			}),
			Auth(testLogger(), fakeAuthenticator{principal: want}),
		)

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusOK, rec.Code)
	})
}

type denyLimiter struct{}

func (denyLimiter) Allow(context.Context, string) (bool, error) { return false, nil }

func TestRateLimit(t *testing.T) {
	t.Parallel()

	h := Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("handler must not be called")
		}),
		RateLimit(testLogger(), denyLimiter{}, func(*http.Request) string { return "client" }),
	)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusTooManyRequests, rec.Code)
	require.Equal(t, "60", rec.Header().Get("Retry-After"))
	require.JSONEq(t, `{"errors":[{"message":"too many requests","code":0}]}`, rec.Body.String())
}
