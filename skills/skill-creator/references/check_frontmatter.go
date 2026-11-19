// Command check_frontmatter проверяет frontmatter навыков goskills.
//
// Запуск из корня репозитория:
//
//	go run skills/skill-creator/references/check_frontmatter.go skills/*/SKILL.md
//
// Проверяется ровно то, на чём держится навык: блок между ---, наличие
// name и description, совпадение name с именем каталога и двойные кавычки
// у description. Полный YAML-парсер сюда не тянется: схема frontmatter
// фиксированная и простая, а внешние зависимости в репозитории не нужны.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var skillIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run check_frontmatter.go <SKILL.md>...")
		os.Exit(2)
	}

	failed := false
	for _, path := range os.Args[1:] {
		if err := check(path); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
	fmt.Printf("Проверено файлов: %d\n", len(os.Args)-1)
}

func check(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Errorf("frontmatter должен начинаться с ---")
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return fmt.Errorf("frontmatter не закрыт строкой ---")
	}

	fields := make(map[string]string)
	quoted := make(map[string]bool)
	for i := 1; i < end; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		key, raw, ok := strings.Cut(line, ":")
		if !ok {
			return fmt.Errorf("строка %d: ожидается «ключ: значение»", i+1)
		}
		key = strings.TrimSpace(key)
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue // вложенный mapping, например metadata
		}
		if _, dup := fields[key]; dup {
			return fmt.Errorf("строка %d: ключ %q повторяется", i+1, key)
		}

		value, quote, err := parseScalar(key, raw)
		if err != nil {
			return fmt.Errorf("строка %d: %w", i+1, err)
		}
		fields[key] = value
		quoted[key] = quote == '"'
	}

	if fields["name"] == "" {
		return fmt.Errorf("нет обязательного поля name")
	}
	if !skillIDPattern.MatchString(fields["name"]) {
		return fmt.Errorf("name %q: нужна латиница в нижнем регистре, цифры и дефис", fields["name"])
	}
	dir := filepath.Base(filepath.Dir(path))
	if fields["name"] != dir {
		return fmt.Errorf("name %q не совпадает с каталогом %q", fields["name"], dir)
	}
	if fields["description"] == "" {
		return fmt.Errorf("нет обязательного поля description")
	}
	if !quoted["description"] {
		return fmt.Errorf("description должен быть в двойных кавычках")
	}
	return nil
}

// parseScalar разбирает значение верхнего уровня. Возвращает значение и
// кавычку, в которой оно было: нет, одинарная или двойная.
func parseScalar(key, raw string) (string, byte, error) {
	switch raw[0] {
	case '"':
		if len(raw) < 2 || !strings.HasSuffix(raw, `"`) {
			return "", 0, fmt.Errorf("поле %s: незакрытая двойная кавычка", key)
		}
		value, err := strconv.Unquote(raw)
		if err != nil {
			return "", 0, fmt.Errorf("поле %s: %w", key, err)
		}
		return value, '"', nil

	case '\'':
		if len(raw) < 3 || !strings.HasSuffix(raw, "'") {
			return "", 0, fmt.Errorf("поле %s: незакрытая одинарная кавычка", key)
		}
		return strings.ReplaceAll(raw[1:len(raw)-1], "''", "'"), '\'', nil
	}

	if strings.Contains(raw, ": ") {
		return "", 0, fmt.Errorf("поле %s: двоеточие с пробелом без кавычек ломает YAML", key)
	}
	if strings.Contains(raw, " #") {
		return "", 0, fmt.Errorf("поле %s: « #» без кавычек начинает комментарий", key)
	}
	if strings.ContainsRune(`*&!-{[#>`, rune(raw[0])) {
		return "", 0, fmt.Errorf("поле %s: значение начинается с %q, нужны кавычки", key, raw[0])
	}
	return raw, 0, nil
}
