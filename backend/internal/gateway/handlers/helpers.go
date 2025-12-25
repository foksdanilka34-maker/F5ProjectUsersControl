package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// parseID парсит int64 из строки (path value или query param)
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// parseOptionalID парсит optional int64, возвращает nil при пустой строке
func parseOptionalID(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getClientIP(r *http.Request) string {
	// Сначала проверяем X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Затем X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// В противном случае - RemoteAddr
	addr := r.RemoteAddr
	// Убираем порт если есть
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return addr
}
