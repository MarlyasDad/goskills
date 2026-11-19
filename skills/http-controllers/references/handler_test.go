package documents

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"<module>/internal/domain"
	"<module>/internal/shared"
	"<module>/internal/transport/<appname>/http/responses"

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
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"invalid request","code":%d}]}`,
				responses.CodeUnassigned,
			),
		},
		{
			name:       "body too large",
			body:       `{"title":"` + strings.Repeat("a", maxBodyBytes) + `"}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"request body too large","code":%d}]}`,
				responses.CodeUnassigned,
			),
		},
		{
			name:       "missing title",
			body:       `{"kind":"inspection"}`,
			wantStatus: http.StatusBadRequest,
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"title: value is required","code":%d}]}`,
				responses.CodeUnassigned,
			),
		},
		{
			name:       "conflict",
			body:       `{"title":"Акт","kind":"inspection"}`,
			err:        domain.ErrConflict,
			wantStatus: http.StatusConflict,
			wantBody: fmt.Sprintf(
				`{"errors":[{"message":"conflict","code":%d}]}`,
				responses.CodeUnassigned,
			),
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
