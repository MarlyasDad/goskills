package http

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Таймауты по умолчанию. Ноль в Config означает «не задан»: NewServer
// подставляет значение отсюда, а не config.Load.
const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 15 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultIdleTimeout       = 60 * time.Second
)

// Config — настройки HTTP-сервера. Значения приходят из internal/config.
// Таймауты необязательны: 0 — значение по умолчанию.
type Config struct {
	Host string
	// Port 0 — эфемерный порт для тестов; в проде — 8080.
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	AllowedOrigins    []string
	Metrics           bool
	Swagger           bool
	// SwaggerHost пустой — Swagger UI берёт текущий origin.
	SwaggerHost string
	// SwaggerBasePath по умолчанию "/".
	SwaggerBasePath string
}

// Addr собирает адрес через net.JoinHostPort: это корректно и для IPv6.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// withDefaults подставляет таймауты по умолчанию. Отрицательные значения
// не исправляет: это ошибка конфигурации, её ловит Validate.
func (c Config) withDefaults() Config {
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultWriteTimeout
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = defaultIdleTimeout
	}
	return c
}

// Validate проверяет обязательные настройки сервера. Вызывается из NewServer
// после withDefaults, а не из config.Load.
func (c Config) Validate() error {
	// 0 — эфемерный порт для тестов; в проде задаётся 8080.
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("http: port %d out of range", c.Port)
	}
	if c.ReadHeaderTimeout < 0 {
		return fmt.Errorf("http: read header timeout must not be negative")
	}
	if c.ReadTimeout < 0 {
		return fmt.Errorf("http: read timeout must not be negative")
	}
	if c.WriteTimeout < 0 {
		return fmt.Errorf("http: write timeout must not be negative")
	}
	if c.IdleTimeout < 0 {
		return fmt.Errorf("http: idle timeout must not be negative")
	}
	if c.MaxHeaderBytes < 0 {
		return fmt.Errorf("http: max header bytes must not be negative")
	}
	return nil
}
