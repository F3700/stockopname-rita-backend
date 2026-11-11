package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

type fakePinger struct {
	err   error
	delay time.Duration
}

func (f fakePinger) Ping(ctx context.Context) error {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return f.err
}

func TestHealthLive(t *testing.T) {
	handler := NewHealthHandler(fakePinger{err: errors.New("database down")}, time.Second)

	recorder := httptest.NewRecorder()
	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/health", nil), httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Errorf("liveness must ignore the database, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"ok"`) {
		t.Errorf("unexpected body %s", recorder.Body.String())
	}
}

func TestHealthReadyWhenDatabaseAnswers(t *testing.T) {
	handler := NewHealthHandler(fakePinger{}, time.Second)

	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil), httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"db":"ok"`) {
		t.Errorf("unexpected body %s", recorder.Body.String())
	}
}

func TestHealthReadyWhenDatabaseFails(t *testing.T) {
	handler := NewHealthHandler(fakePinger{err: errors.New("connection refused")}, time.Second)

	recorder := httptest.NewRecorder()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil), httprouter.Params{})

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

// A hanging database must not hold the probe open forever.
func TestHealthReadyTimesOut(t *testing.T) {
	handler := NewHealthHandler(fakePinger{delay: 200 * time.Millisecond}, 10*time.Millisecond)

	recorder := httptest.NewRecorder()
	start := time.Now()
	handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil), httprouter.Params{})

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on timeout, got %d", recorder.Code)
	}
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Errorf("probe did not honour its timeout, took %s", elapsed)
	}
}
