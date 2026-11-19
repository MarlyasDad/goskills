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
	"<module>/internal/transport/<appname>/http/responses"
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

	if errs := req.validate(); len(errs) > 0 {
		if err := responses.RespondWithValidationErrors(w, errs); err != nil {
			h.log.ErrorContext(r.Context(), "write validation response", "error", err)
		}
		return
	}

	out, err := h.creator.Create(r.Context(), req.toInput())
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	if err := responses.RespondWithJSON(w, http.StatusCreated, fromCreateOutput(out)); err != nil {
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
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, errBodyTooLarge):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, context.Canceled):
		h.log.WarnContext(r.Context(), "request canceled", "error", err)
		return
	case errors.Is(err, context.DeadlineExceeded):
		status = http.StatusGatewayTimeout
	}

	// TODO: код из корпоративного справочника. До присвоения — заглушка,
	// чтобы не выдумывать значения.
	code := responses.CodeUnassigned

	level := slog.LevelWarn
	if status >= http.StatusInternalServerError {
		level = slog.LevelError
	}
	h.log.LogAttrs(r.Context(), level, "request failed",
		slog.Any("error", err),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int("status", status),
		slog.Int("code", code),
	)

	if err := responses.RespondWithError(w, status, code, publicMessage(status)); err != nil {
		h.log.ErrorContext(r.Context(), "write error response", "error", err)
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
