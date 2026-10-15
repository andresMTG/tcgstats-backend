package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var allowedCORSMethods = map[string]struct{}{
	http.MethodGet:  {},
	http.MethodPost: {},
}

var allowedCORSHeaders = map[string]struct{}{
	"authorization": {},
	"content-type":  {},
}

func NewCORSMiddleware(origins []string) (func(http.Handler) http.Handler, error) {
	allowedOrigins := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if err := validateCORSOrigin(origin); err != nil {
			return nil, err
		}
		allowedOrigins[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowedOrigins[origin]; !ok {
				http.Error(w, "origin not allowed", http.StatusForbidden)
				return
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if !isAllowedPreflight(r) {
					http.Error(w, "CORS preflight not allowed", http.StatusForbidden)
					return
				}
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			next.ServeHTTP(w, r)
		})
	}, nil
}

func validateCORSOrigin(origin string) error {
	if origin == "" || origin == "*" {
		return fmt.Errorf("CORS origin %q is invalid; configure explicit origins", origin)
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		(parsed.Path != "" && parsed.Path != "/") {
		return fmt.Errorf("CORS origin %q must be an explicit http or https origin without a path", origin)
	}

	canonicalOrigin := parsed.Scheme + "://" + parsed.Host
	if origin != canonicalOrigin {
		return fmt.Errorf("CORS origin %q must not contain a path or trailing slash", origin)
	}
	return nil
}

func isAllowedPreflight(r *http.Request) bool {
	method := strings.ToUpper(strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")))
	if _, ok := allowedCORSMethods[method]; !ok {
		return false
	}

	for _, header := range strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",") {
		header = strings.ToLower(strings.TrimSpace(header))
		if header == "" {
			continue
		}
		if _, ok := allowedCORSHeaders[header]; !ok {
			return false
		}
	}
	return true
}
