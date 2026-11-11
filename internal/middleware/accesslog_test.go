package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// captureLog redirects the standard logger for one test. Tests in this package
// must not run in parallel, otherwise they would share the log output.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buffer bytes.Buffer
	log.SetOutput(&buffer)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buffer
}

func TestAccessLogRecordsStatusAndBytes(t *testing.T) {
	logs := captureLog(t)

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/products?page=2", nil))

	line := logs.String()
	for _, want := range []string{
		"method=POST",
		"uri=/products?page=2",
		"status=201",
		"bytes=5",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("log line missing %q: %s", want, line)
		}
	}
}

func TestAccessLogDefaultsToOK(t *testing.T) {
	logs := captureLog(t)

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(logs.String(), "status=200") {
		t.Errorf("expected implicit 200, got %s", logs.String())
	}
}

func TestAccessLogEchoesClientRequestID(t *testing.T) {
	logs := captureLog(t)

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.Header.Set(RequestIDHeader, "abc-123")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(RequestIDHeader); got != "abc-123" {
		t.Errorf("expected request id echo, got %q", got)
	}
	if !strings.Contains(logs.String(), "id=abc-123") {
		t.Errorf("log line missing the client request id: %s", logs.String())
	}
}

func TestAccessLogGeneratesRequestID(t *testing.T) {
	logs := captureLog(t)

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/products", nil))

	id := recorder.Header().Get(RequestIDHeader)
	if id == "" {
		t.Fatal("expected a generated request id")
	}
	if !strings.Contains(logs.String(), "id="+id) {
		t.Errorf("log line missing the generated id %q: %s", id, logs.String())
	}
}

// Probes run on a timer; logging them would bury real requests.
func TestAccessLogSkipsLivenessProbe(t *testing.T) {
	logs := captureLog(t)

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	if logs.Len() != 0 {
		t.Errorf("expected no log line for /health, got %s", logs.String())
	}
}

// A recovered panic must be logged as the 500 Recovery wrote, not as the status
// the handler never set.
func TestAccessLogSeesRecoveredPanic(t *testing.T) {
	logs := captureLog(t)

	handler := AccessLog(Recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/products", nil))

	if !strings.Contains(logs.String(), "status=500") {
		t.Errorf("expected status=500 after a recovered panic, got %s", logs.String())
	}
}

func TestAccessLogAttachesRequestIDToContext(t *testing.T) {
	captureLog(t)

	var seen string
	handler := AccessLog(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		seen = RequestIDFromContext(req.Context())
	}))

	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.Header.Set(RequestIDHeader, "ctx-id")
	handler.ServeHTTP(httptest.NewRecorder(), request)

	if seen != "ctx-id" {
		t.Errorf("expected request id in context, got %q", seen)
	}
}

func TestStatusRecorderKeepsFirstStatus(t *testing.T) {
	recorder := &statusRecorder{ResponseWriter: httptest.NewRecorder()}

	recorder.WriteHeader(http.StatusBadRequest)
	recorder.WriteHeader(http.StatusTeapot)

	if recorder.status != http.StatusBadRequest {
		t.Errorf("expected the first status to win, got %d", recorder.status)
	}
}
