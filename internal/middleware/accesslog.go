package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"stockopname-rita-backend/internal/helper"
)

// RequestIDHeader carries the correlation id of one request. It is echoed back
// so a user report can be matched to a log line.
const RequestIDHeader = "X-Request-Id"

// skipLogPaths holds probes that run on a timer: logging them would bury every
// other line without adding information.
var skipLogPaths = map[string]struct{}{
	"/health": {},
}

type requestIDKey struct{}

// RequestIDFromContext returns the id AccessLog attached, or "" when the
// request did not pass through it.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// statusRecorder captures what the handler actually wrote: the status code and
// the number of body bytes.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	written, err := r.ResponseWriter.Write(body)
	r.bytes += written
	return written, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// AccessLog writes one line per request. It must sit outside Recovery: a
// recovered panic then shows up as the 500 Recovery wrote, instead of the status
// the handler never got to set.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		requestID := strings.TrimSpace(req.Header.Get(RequestIDHeader))
		if requestID == "" {
			requestID = helper.NewUUID()
		}
		writer.Header().Set(RequestIDHeader, requestID)
		req = req.WithContext(context.WithValue(req.Context(), requestIDKey{}, requestID))

		recorder := &statusRecorder{ResponseWriter: writer}
		start := time.Now()

		next.ServeHTTP(recorder, req)

		if _, skip := skipLogPaths[req.URL.Path]; skip {
			return
		}

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		// Milliseconds with a fixed precision keeps the line ASCII-only: the
		// default Duration.String() uses the micro sign, which mangles in some
		// consoles and log shippers.
		log.Printf(
			"request id=%s method=%s uri=%s status=%d dur=%.2fms bytes=%d remote=%s",
			requestID,
			req.Method,
			req.URL.RequestURI(),
			status,
			float64(time.Since(start).Microseconds())/1000,
			recorder.bytes,
			req.RemoteAddr,
		)
	})
}
