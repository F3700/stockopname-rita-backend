package middleware

import (
	"net/http"
	"strings"
)

// CORS answers only the origins in allowedOrigins. The wildcard is gone on
// purpose: a response carrying "*" may not be used with credentials, so the
// session cookie would never reach the browser from a cross-origin call.
func CORS(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if _, ok := allowed[origin]; ok && origin != "" {
			// Echo the origin instead of "*": only then may the browser attach
			// the session cookie.
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// A cache must not reuse one origin's answer for another.
		writer.Header().Add("Vary", "Origin")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Actor")
		writer.Header().Set("Access-Control-Max-Age", "86400")

		if req.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, req)
	})
}
