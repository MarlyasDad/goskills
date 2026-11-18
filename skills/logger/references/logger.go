// Package logger создаёт slog.Logger с настройками приложения.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config — настройки логгера. Заполняется в internal/config и передаётся
// аргументом: env и файлы сам пакет не читает.
type Config struct {
	Level       string // debug | info | warn | warning | error
	Format      string // json | text
	Output      string // stdout | stderr
	ServiceName string
	Version     string
	Environment string
}

// New создаёт логгер с общими атрибутами сервиса.
//
// Неверные значения не мешают запуску: level → info, format → json,
// output → stdout. Логирование не должно быть причиной падения сервиса.
func New(cfg Config) *slog.Logger {
	return newLogger(cfg, parseOutput(cfg.Output))
}

// newLogger отделяет выбор writer, чтобы тесты писали в bytes.Buffer,
// а не перехватывали os.Stdout.
func newLogger(cfg Config, output io.Writer) *slog.Logger {
	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{
		Level: level,
		// source (файл:строка) дорог и нужен только при отладке.
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "text") {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.Version),
		slog.String("env", cfg.Environment),
	})

	return slog.New(handler)
}

// parseLevel переводит строку в уровень slog. Неизвестное значение — info.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// parseOutput выбирает поток. По умолчанию — stdout: контейнер сам собирает
// его, файлы и ротация — задача сборщика логов.
func parseOutput(output string) io.Writer {
	switch strings.ToLower(output) {
	case "stderr":
		return os.Stderr
	default:
		return os.Stdout
	}
}
