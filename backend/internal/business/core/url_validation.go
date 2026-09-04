package core

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidBaseURL = errors.New("base_url must be an absolute http(s) URL")

func validateBaseURL(raw string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return "", fmt.Errorf("%w: empty", ErrInvalidBaseURL)
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBaseURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%w: scheme must be http or https", ErrInvalidBaseURL)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%w: missing host", ErrInvalidBaseURL)
	}

	return trimmed, nil
}
