package documents

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"<module>/internal/domain"
	"<module>/internal/shared"

	"github.com/stretchr/testify/require"
)

type fakeCreator struct {
	out shared.CreateDocumentOutput
	err error
}

func (f fakeCreator) Create(context.Context, shared.CreateDocumentInput) (shared.CreateDocumentOutput, error) {
	return f.out, f.err
}

func TestCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "created",
			body:       `{"title":"Акт","kind":"inspection"}`,
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"1","status":"created"}`,
		},
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid request","code":"invalid_request"}`,
		},
		{
			name:       "body too large",
			body:       `{"title":"` + strings.Repeat("a", maxBodyBytes) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody:   `{"error":"request body too large","code":"payload_too_large"}`,
		},
		{
			name:       "conflict",
			body:       `{"title":"Акт","kind":"inspection"}`,
			err:        domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":"conflict","code":"conflict"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := New(
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				fakeCreator{
					out: shared.CreateDocumentOutput{ID: "1", Status: "created"},
					err: tt.err,
				},
			)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/v1/documents", strings.NewReader(tt.body))

			h.Create(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			require.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
