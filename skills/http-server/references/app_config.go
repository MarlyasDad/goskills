// Скопировать в internal/config/config.go. Заменить <module> и <appname>.
// Требует helpers String/Duration/Int/Bool/CSV из internal/config/env.go.
package config

import (
	"fmt"
	"time"

	"<module>/internal/platform/logger"
	apihttp "<module>/internal/transport/<appname>/http"
)

// Config — зеркало конфиг-файла, общее для всех приложений модуля.
type Config struct {
	ShutdownTimeout time.Duration
	Logger          logger.Config
	HTTP            apihttp.Config
}

// Load читает env и разбирает все секции. Обязательные поля HTTP проверяет
// http.NewServer, а не Load: приложение без HTTP не обязано их заполнять.
func Load() (*Config, error) {
	shutdownTimeout, err := Duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
	}
	if shutdownTimeout <= 0 {
		return nil, fmt.Errorf("shutdown timeout must be positive")
	}

	port, err := Int("HTTP_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("HTTP_PORT: %w", err)
	}
	readHeaderTimeout, err := Duration("HTTP_READ_HEADER_TIMEOUT", 0)
	if err != nil {
		return nil, fmt.Errorf("HTTP_READ_HEADER_TIMEOUT: %w", err)
	}
	readTimeout, err := Duration("HTTP_READ_TIMEOUT", 0)
	if err != nil {
		return nil, fmt.Errorf("HTTP_READ_TIMEOUT: %w", err)
	}
	writeTimeout, err := Duration("HTTP_WRITE_TIMEOUT", 0)
	if err != nil {
		return nil, fmt.Errorf("HTTP_WRITE_TIMEOUT: %w", err)
	}
	idleTimeout, err := Duration("HTTP_IDLE_TIMEOUT", 0)
	if err != nil {
		return nil, fmt.Errorf("HTTP_IDLE_TIMEOUT: %w", err)
	}
	maxHeaderBytes, err := Int("HTTP_MAX_HEADER_BYTES", 1<<20)
	if err != nil {
		return nil, fmt.Errorf("HTTP_MAX_HEADER_BYTES: %w", err)
	}
	metrics, err := Bool("HTTP_METRICS", false)
	if err != nil {
		return nil, fmt.Errorf("HTTP_METRICS: %w", err)
	}
	swagger, err := Bool("HTTP_SWAGGER", false)
	if err != nil {
		return nil, fmt.Errorf("HTTP_SWAGGER: %w", err)
	}
	swaggerHost := String("HTTP_SWAGGER_HOST", "")
	swaggerBasePath := String("HTTP_SWAGGER_BASE_PATH", "/")

	return &Config{
		ShutdownTimeout: shutdownTimeout,
		Logger: logger.Config{
			Level:       String("LOG_LEVEL", "info"),
			Format:      String("LOG_FORMAT", "json"),
			Output:      String("LOG_OUTPUT", "stdout"),
			ServiceName: String("SERVICE_NAME", "<appname>"),
			Version:     String("SERVICE_VERSION", "dev"),
			Environment: String("SERVICE_ENVIRONMENT", ""),
		},
		HTTP: apihttp.Config{
			Host:              String("HTTP_HOST", ""),
			Port:              port,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			MaxHeaderBytes:    maxHeaderBytes,
			AllowedOrigins:    CSV("HTTP_CORS_ALLOWED_ORIGINS"),
			Metrics:           metrics,
			Swagger:           swagger,
			SwaggerHost:       swaggerHost,
			SwaggerBasePath:   swaggerBasePath,
		},
	}, nil
}
