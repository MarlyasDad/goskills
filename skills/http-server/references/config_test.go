package http

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validConfig() Config {
	return Config{
		Port:           8080,
		MaxHeaderBytes: 1 << 20,
	}
}

func TestConfig_Addr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "ipv4", cfg: Config{Host: "127.0.0.1", Port: 8080}, want: "127.0.0.1:8080"},
		{name: "all interfaces", cfg: Config{Port: 8080}, want: ":8080"},
		{name: "ipv6", cfg: Config{Host: "::1", Port: 8080}, want: "[::1]:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.cfg.Addr())
		})
	}
}

func TestConfig_WithDefaults(t *testing.T) {
	t.Parallel()

	t.Run("zero values", func(t *testing.T) {
		t.Parallel()

		cfg := Config{Port: 8080}.withDefaults()

		require.Equal(t, defaultReadHeaderTimeout, cfg.ReadHeaderTimeout)
		require.Equal(t, defaultReadTimeout, cfg.ReadTimeout)
		require.Equal(t, defaultWriteTimeout, cfg.WriteTimeout)
		require.Equal(t, defaultIdleTimeout, cfg.IdleTimeout)
	})

	t.Run("explicit values", func(t *testing.T) {
		t.Parallel()

		cfg := Config{
			Port:              8080,
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       2 * time.Second,
			WriteTimeout:      3 * time.Second,
			IdleTimeout:       4 * time.Second,
		}.withDefaults()

		require.Equal(t, time.Second, cfg.ReadHeaderTimeout)
		require.Equal(t, 2*time.Second, cfg.ReadTimeout)
		require.Equal(t, 3*time.Second, cfg.WriteTimeout)
		require.Equal(t, 4*time.Second, cfg.IdleTimeout)
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{name: "valid without timeouts"},
		{name: "zero port for tests", mutate: func(c *Config) { c.Port = 0 }},
		{name: "negative port", mutate: func(c *Config) { c.Port = -1 }, wantErr: "port"},
		{name: "negative read header timeout", mutate: func(c *Config) { c.ReadHeaderTimeout = -time.Second }, wantErr: "read header timeout"},
		{name: "negative read timeout", mutate: func(c *Config) { c.ReadTimeout = -time.Second }, wantErr: "read timeout"},
		{name: "negative write timeout", mutate: func(c *Config) { c.WriteTimeout = -time.Second }, wantErr: "write timeout"},
		{name: "negative idle timeout", mutate: func(c *Config) { c.IdleTimeout = -time.Second }, wantErr: "idle timeout"},
		{name: "negative max header bytes", mutate: func(c *Config) { c.MaxHeaderBytes = -1 }, wantErr: "max header bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validConfig()
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}

			err := cfg.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestNewServer_AppliesDefaults(t *testing.T) {
	t.Parallel()

	server, err := NewServer(
		validConfig(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	)
	require.NoError(t, err)

	require.Equal(t, defaultReadHeaderTimeout, server.ReadHeaderTimeout)
	require.Equal(t, defaultReadTimeout, server.ReadTimeout)
	require.Equal(t, defaultWriteTimeout, server.WriteTimeout)
	require.Equal(t, defaultIdleTimeout, server.IdleTimeout)
}
