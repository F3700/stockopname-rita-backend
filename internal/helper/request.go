package helper

import (
	"encoding/json"
	"net/http"
	"strings"

	"stockopname-rita-backend/internal/auth"
)

func ReadFromRequestBody(req *http.Request, result interface{}) error {
	defer req.Body.Close()
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(result)
}

// ActorFromRequest returns the name recorded in the audit log. The admin web
// sends it inside the session cookie, which RequireAuth puts on the context; the
// mobile app still self-declares it in a header, which is why the fallback stays
// until the join token lands.
func ActorFromRequest(req *http.Request) string {
	if actor := auth.ActorFromContext(req.Context()); actor != "" {
		return actor
	}

	actor := strings.TrimSpace(req.Header.Get("X-Actor"))
	if actor == "" {
		return "admin-web"
	}
	runes := []rune(actor)
	if len(runes) > 100 {
		return string(runes[:100])
	}
	return actor
}
