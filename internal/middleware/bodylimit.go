package middleware

import (
	"net/http"
	"strconv"
)

// LimitBody caps how many bytes a client may send in one request body.
//
// Requests that declare an oversized Content-Length are rejected before the
// handler runs. Chunked requests (no Content-Length) are capped by
// http.MaxBytesReader, so reading fails inside the handler instead of the
// server buffering an unbounded payload. A non-positive maxBytes disables
// the limit.
func LimitBody(next http.Handler, maxBytes int64) http.Handler {
	if maxBytes <= 0 {
		return next
	}

	message := "Request body too large (max " + strconv.FormatInt(maxBytes, 10) + " bytes)"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxBytes {
			w.Header().Set("Connection", "close")
			writeJsonError(w, http.StatusRequestEntityTooLarge, message)
			return
		}

		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}

		next.ServeHTTP(w, r)
	})
}
