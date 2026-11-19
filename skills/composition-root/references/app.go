package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"<module>/internal/config"
	"<module>/internal/infra/postgres"
	"<module>/internal/services/api/document"
	apihttp "<module>/internal/transport/api/http"
	"<module>/internal/transport/api/http/documents"
)

// App — composition root приложения: собранные компоненты и их lifecycle.
type App struct {
	log             *slog.Logger
	server          *http.Server
	shutdownTimeout time.Duration
	cleanup         func() error
}

// New создаёт адаптеры, репозитории, сервисы, контроллеры, router и сервер.
// Ничего не запускает и не биндит порт.
func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	pg, err := postgres.New(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	repo := postgres.NewDocumentRepository(pg)
	svc := document.New(repo)
	handler := documents.New(log, svc)

	router := apihttp.NewRouter(apihttp.Deps{
		Config:   cfg.HTTP,
		Log:      log,
		Handlers: apihttp.HandlerSet{Documents: handler},
	})

	server, err := apihttp.NewServer(cfg.HTTP, log, router)
	if err != nil {
		pg.Close()
		return nil, fmt.Errorf("http server: %w", err)
	}

	return &App{
		log:             log,
		server:          server,
		shutdownTimeout: cfg.ShutdownTimeout,
		cleanup: func() error {
			pg.Close()
			return nil
		},
	}, nil
}

// Run запускает сервер, ждёт отмены контекста или ошибки и останавливает
// компоненты. Cleanup вызывается всегда, в том числе при ошибке запуска.
func (a *App) Run(ctx context.Context) (err error) {
	defer func() {
		if a.cleanup != nil {
			err = errors.Join(err, a.cleanup())
		}
	}()

	ln, err := net.Listen("tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", a.server.Addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("http server started", "addr", ln.Addr().String())
		serveErr := a.server.Serve(ln)
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		_ = a.server.Close()
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if err := <-errCh; err != nil {
		return fmt.Errorf("serve http: %w", err)
	}
	return nil
}
