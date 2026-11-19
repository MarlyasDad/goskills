package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadFile загружает env-файл в окружение. Пустой путь — no-op:
// в k8s файла нет, остаются только переменные из ConfigMap/Secret.
func LoadFile(path string) error {
	if path == "" {
		return nil
	}

	values, err := parseFile(path)
	if err != nil {
		return err
	}

	for key, value := range values {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}
	return nil
}

// parseFile читает простой env-файл: KEY=VALUE, комментарии # и пустые
// строки. Подстановок, export и inline-комментариев нет — формат один и
// тот же локально и в k8s.
func parseFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	line := 0

	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		key, value, ok := strings.Cut(text, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: expected KEY=VALUE", path, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("%s:%d: empty key", path, line)
		}

		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}
