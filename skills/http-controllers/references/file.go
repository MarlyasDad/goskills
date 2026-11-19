package responses

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	// ContentTypeDOCX — MIME-тип docx.
	ContentTypeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	// ContentTypeBinary — fallback, если тип файла не задан.
	ContentTypeBinary = "application/octet-stream"
)

// RespondWithFile отправляет файл потоком. Body обязателен.
//
// Размер берётся из bytes.Buffer/bytes.Reader (Len) или io.Seeker. Если
// размер неизвестен, Content-Length не ставится и ответ идёт chunked.
//
// Ошибка после WriteHeader означает, что заголовки и часть тела уже ушли:
// изменить статус нельзя, вызывающий должен только залогировать ошибку.
func RespondWithFile(
	w http.ResponseWriter,
	status int,
	name string,
	contentType string,
	body io.Reader,
) error {
	if body == nil {
		return fmt.Errorf("responses: file body is nil")
	}
	if contentType == "" {
		contentType = ContentTypeBinary
	}

	size, known, err := readerSize(body)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", contentDisposition(name))
	if known {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}

	w.WriteHeader(status)
	_, err = io.Copy(w, body)
	return err
}

func contentDisposition(name string) string {
	name = sanitizeFilename(name)
	if name == "" {
		name = "download"
	}

	value := mime.FormatMediaType("attachment", map[string]string{"filename": name})
	if value == "" {
		return "attachment"
	}
	return value
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)

	if name == "." || name == "/" {
		return ""
	}
	return name
}

func readerSize(r io.Reader) (int64, bool, error) {
	if l, ok := r.(interface{ Len() int }); ok {
		return int64(l.Len()), true, nil
	}

	seeker, ok := r.(io.Seeker)
	if !ok {
		return 0, false, nil
	}

	current, err := seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, false, nil
	}
	end, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, false, nil
	}
	if _, err := seeker.Seek(current, io.SeekStart); err != nil {
		return 0, false, fmt.Errorf("responses: restore file position: %w", err)
	}

	return end - current, true, nil
}
