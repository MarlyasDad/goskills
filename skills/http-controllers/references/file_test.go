package responses

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRespondWithFile(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	body := bytes.NewBufferString("docx")

	err := RespondWithFile(
		rec,
		http.StatusOK,
		"act.docx",
		ContentTypeDOCX,
		body,
	)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, ContentTypeDOCX, rec.Header().Get("Content-Type"))
	require.Equal(t, "attachment; filename=act.docx", rec.Header().Get("Content-Disposition"))
	require.Equal(t, "4", rec.Header().Get("Content-Length"))
	require.Equal(t, "docx", rec.Body.String())
}

type readerOnly struct {
	io.Reader
}

func TestRespondWithFile_UnknownSize(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	err := RespondWithFile(
		rec,
		http.StatusOK,
		"act.docx",
		ContentTypeDOCX,
		readerOnly{Reader: strings.NewReader("docx")},
	)
	require.NoError(t, err)

	require.Empty(t, rec.Header().Get("Content-Length"))
	require.Equal(t, "docx", rec.Body.String())
}

func TestRespondWithFile_FallbackContentType(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	err := RespondWithFile(rec, http.StatusOK, "data.bin", "", bytes.NewBufferString("x"))
	require.NoError(t, err)

	require.Equal(t, ContentTypeBinary, rec.Header().Get("Content-Type"))
}

func TestRespondWithFile_SanitizesFilename(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	err := RespondWithFile(rec, http.StatusOK, "../../etc/passwd", ContentTypeDOCX, bytes.NewBufferString("x"))
	require.NoError(t, err)

	require.Equal(t, "attachment; filename=passwd", rec.Header().Get("Content-Disposition"))
}

type failingWriter struct {
	header http.Header
	status int
}

func (w *failingWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingWriter) WriteHeader(status int) {
	w.status = status
}

func (w *failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("broken pipe")
}

func TestRespondWithFile_WriteError(t *testing.T) {
	t.Parallel()

	w := &failingWriter{}

	err := RespondWithFile(
		w,
		http.StatusOK,
		"act.docx",
		ContentTypeDOCX,
		bytes.NewBufferString("docx"),
	)
	require.Error(t, err)

	require.Equal(t, http.StatusOK, w.status)
	require.Equal(t, ContentTypeDOCX, w.Header().Get("Content-Type"))
}
