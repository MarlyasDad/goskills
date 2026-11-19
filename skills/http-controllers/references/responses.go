package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CodeUnassigned — код ещё не присвоен эксплуатацией. Это не код из
// корпоративного справочника, а заглушка: значения кодов не выдумываются.
const CodeUnassigned = 0

// Error — одна ошибка API.
type Error struct {
	Message string `json:"message"`
	// Code — код из корпоративного справочника; 0 — ещё не присвоен.
	Code int `json:"code"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Errors — коллекция ошибок.
type Errors []Error

func (es Errors) Error() string {
	msgs := make([]string, 0, len(es))
	for _, err := range es {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// ErrorResponse — тело ответа с ошибками.
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// RespondWithJSON отправляет JSON. Тело сериализуется до WriteHeader:
// при ошибке маршалинга ещё можно отдать 500, а не пустой 200.
func RespondWithJSON(w http.ResponseWriter, status int, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		if writeErr := RespondWithText(w, http.StatusInternalServerError, "internal server error"); writeErr != nil {
			return fmt.Errorf("marshal response: %w; write fallback: %v", err, writeErr)
		}
		return fmt.Errorf("marshal response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return nil
	}

	_, err = w.Write(body)
	return err
}

// RespondWithText отправляет текстовый ответ.
func RespondWithText(w http.ResponseWriter, status int, message string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write([]byte(message))
	return err
}

// RespondWithError отправляет одну ошибку в JSON.
func RespondWithError(w http.ResponseWriter, status, code int, message string) error {
	return RespondWithJSON(w, status, ErrorResponse{
		Errors: []Error{{Message: message, Code: code}},
	})
}

// ValidationError — ошибка валидации поля.
type ValidationError struct {
	// Field не входит в контракт 400: имя поля уже внутри Message.
	// Если проект всё же отдаёт field, заполни его — omitempty оставит
	// поведение совместимым.
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	// Code — код из корпоративного справочника; 0 — ещё не присвоен.
	Code int `json:"code"`
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors — коллекция ошибок валидации.
type ValidationErrors []ValidationError

func (es ValidationErrors) Error() string {
	msgs := make([]string, 0, len(es))
	for _, err := range es {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// ValidationErrorResponse — тело ответа с ошибками валидации.
type ValidationErrorResponse struct {
	Errors []ValidationError `json:"errors"`
}

// RespondWithValidationErrors отправляет 400 с ошибками валидации.
func RespondWithValidationErrors(w http.ResponseWriter, errs []ValidationError) error {
	return RespondWithJSON(w, http.StatusBadRequest, ValidationErrorResponse{Errors: errs})
}
