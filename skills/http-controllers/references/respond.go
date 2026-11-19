package respond

import (
	"encoding/json"
	"net/http"
)

// ErrorBody — единый формат ошибки API.
type ErrorBody struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// JSON сериализует ответ до WriteHeader: при ошибке маршалинга ещё можно
// отдать 500, а не пустой 200.
func JSON(w http.ResponseWriter, status int, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		writeInternalError(w)
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status == http.StatusNoContent {
		return nil
	}

	_, err = w.Write(body)
	return err
}

// Error отдаёт ошибку в общем формате.
func Error(w http.ResponseWriter, status int, code, message string) error {
	return JSON(w, status, ErrorBody{Error: message, Code: code})
}

func writeInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(ErrorBody{
		Error: "internal server error",
		Code:  "internal_error",
	})
}
