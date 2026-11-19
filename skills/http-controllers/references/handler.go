package documents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"<module>/internal/domain"
	"<module>/internal/transport/<appname>/http/respond"
)

const maxBodyBytes = 1 << 20 // 1 MiB

var errBodyTooLarge = errors.New("request body too large")

// Handler — тонкий HTTP-адаптер над операцией сервиса.
type Handler struct {
	log     *slog.Logger
	creator Creator
}

func New(log *slog.Logger, creator Creator) *Handler {
	return &Handler{log: log, creator: creator}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		if errors.Is(err, errBodyTooLarge) {
			h.writeError(w, r, err)
			return
		}
		h.writeError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
		return
	}

	input, err := req.toInput()
	if err != nil {
		h.writeError(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
		return
	}

	out, err := h.creator.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := respond.JSON(w, http.StatusCreated, fromCreateOutput(out)); err != nil {
		h.log.ErrorContext(r.Context(), "write response", "error", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return errBodyTooLarge
		}
		return err
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single JSON object")
	}
	return nil
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code := http.StatusInternalServerError, "internal_error"

	switch {
	case errors.Is(err, errBodyTooLarge):
		status, code = http.StatusRequestEntityTooLarge, "payload_too_large"
	case errors.Is(err, domain.ErrInvalidInput):
		status, code = http.StatusBadRequest, "invalid_request"
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, domain.ErrConflict):
		status, code = http.StatusConflict, "conflict"
	case errors.Is(err, context.Canceled):
		h.log.WarnContext(r.Context(), "request canceled", "error", err)
		return
	case errors.Is(err, context.DeadlineExceeded):
		status, code = http.StatusGatewayTimeout, "timeout"
	}

	level := slog.LevelWarn
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	h.log.LogAttrs(r.Context(), level, "request failed",
		slog.Any("error", err),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
		slog.String("code", code),
	)

	if writeErr := respond.Error(w, status, code, publicMessage(status)); writeErr != nil {
		h.log.ErrorContext(r.Context(), "write error response", "error", writeErr)
	}
}

func publicMessage(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "invalid request"
	case http.StatusRequestEntityTooLarge:
		return "request body too large"
	case http.StatusNotFound:
		return "not found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusGatewayTimeout:
		return "timeout"
	default:
		return "internal server error"
	}
}
