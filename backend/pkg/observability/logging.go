// Package observability — общие для сервисов F5 метрики Prometheus и структурированные логи.
package observability

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// SetupLogging включает JSON-логи через log/slog. Каждая строка получает поле service.
//
// Логи пишутся в stdout и, если задан LOG_DIR, дублируются в файл <LOG_DIR>/<service>.log —
// оттуда их забирает Alloy (см. observability/). slog.SetDefault перенаправляет и
// стандартный log.Printf, поэтому существующие вызовы log.* тоже становятся JSON-строками.
//
// Возвращённую функцию нужно вызвать при завершении сервиса, чтобы закрыть файл.
func SetupLogging(service string) (func(), error) {
	var out io.Writer = os.Stdout
	closeFn := func() {}

	if dir := os.Getenv("LOG_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(filepath.Join(dir, service+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, err
		}
		out = io.MultiWriter(os.Stdout, f)
		closeFn = func() { f.Close() }
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(out, nil)).With("service", service))
	return closeFn, nil
}
