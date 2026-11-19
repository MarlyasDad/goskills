package http

import (
	"errors"
	"log/slog"
	"net/http"
)

// NewServer собирает *http.Server. Порт не биндится, горутины не
// запускаются: запуск и останов — в composition root.
func NewServer(cfg Config, log *slog.Logger, handler http.Handler) (*http.Server, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, errors.New("http: handler is nil")
	}
	if log == nil {
		return nil, errors.New("http: logger is nil")
	}

	return &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}, nil
}
