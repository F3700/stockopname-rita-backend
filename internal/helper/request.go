package helper

import (
	"encoding/json"
	"net/http"
	"strings"
)

func ReadFromRequestBody(req *http.Request, result interface{}) error {
	defer req.Body.Close()
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(result)
}

// ActorFromRequest returns the self-declared admin name for the audit log.
func ActorFromRequest(req *http.Request) string {
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
