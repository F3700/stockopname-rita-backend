package middleware

import (
	"net/http"

	"stockopname-rita-backend/internal/auth"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
)

const (
	invalidSessionMessage  = "Sesi tidak valid atau sudah berakhir. Silakan login."
	missingOperatorMessage = "Nama operator belum diisi."
)

// RequireAuth closes every route that is not whitelisted by auth.IsPublic. It
// must sit inside CORS so a preflight still gets an answer, and outside
// LimitBody so an anonymous request cannot make the server read a large body.
func RequireAuth(manager *auth.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if auth.IsPublic(req.Method, req.URL.Path) {
			next.ServeHTTP(writer, req)
			return
		}

		cookie, err := req.Cookie(auth.CookieName)
		if err != nil {
			writeAuthError(writer, invalidSessionMessage)
			return
		}

		claims, err := manager.Parse(cookie.Value)
		if err != nil {
			writeAuthError(writer, invalidSessionMessage)
			return
		}

		// The operator screen and logout run before a name exists, so they only
		// need the session. Everything else also needs the name, which is what
		// the audit trail records.
		if claims.Operator == "" && !auth.IsSessionOnly(req.URL.Path) {
			writeAuthError(writer, missingOperatorMessage)
			return
		}

		next.ServeHTTP(writer, req.WithContext(auth.WithActor(req.Context(), claims.Operator)))
	})
}

// writeAuthError answers with the same envelope as the rest of the API but
// without logging: a rejected session is an expected outcome, not a fault.
func writeAuthError(writer http.ResponseWriter, message string) {
	_ = helper.ResponseJson(writer, http.StatusUnauthorized, dto.Response{Message: message})
}
