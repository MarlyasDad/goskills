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
	apihttp "<module>/internal/transport/api/http"
)

// App — composition root приложения: собранные компоненты и их lifecycle.
type App struct {
	log             *slog.Logger
	server          *http.Server
	shutdownTimeout time.Duration
	cleanup         func() error
}

// New собирает приложение. В каркасе есть только HTTP-сервер с системными
// ручками; адаптеры, репозитории, сервисы и контроллеры добавляются сюда по
// мере появления.
func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	router := apihttp.NewRouter(apihttp.Deps{
		Config: cfg.HTTP,
		Log:    log,
		// Ready == nil — каркас всегда ready. Когда появится зависимость
		// (БД, внешний сервис), передай func(ctx) error с проверкой Ping.
		Ready:    nil,
		Handlers: apihttp.HandlerSet{},
	})

	server, err := apihttp.NewServer(cfg.HTTP, log, router)
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	return &App{
		log:             log,
		server:          server,
		shutdownTimeout: cfg.ShutdownTimeout,
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
