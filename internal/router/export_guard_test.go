package router

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestExportGuardRejectsWhenFull(t *testing.T) {
	guard := newExportGuard(1)
	release := make(chan struct{})
	started := make(chan struct{})
	var once sync.Once

	blocking := func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		once.Do(func() { close(started) })
		<-release
		w.WriteHeader(http.StatusOK)
	}
	handler := guard.wrap(blocking)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/pdf", nil), nil)
	}()

	<-started

	rejected := httptest.NewRecorder()
	handler(rejected, httptest.NewRequest(http.MethodGet, "/pdf", nil), nil)
	if rejected.Code != http.StatusServiceUnavailable {
		t.Errorf("expected %d while the only slot is busy, got %d", http.StatusServiceUnavailable, rejected.Code)
	}

	close(release)
	wg.Wait()

	// The slot must be free again once the first export finished.
	after := httptest.NewRecorder()
	handler(after, httptest.NewRequest(http.MethodGet, "/pdf", nil), nil)
	if after.Code != http.StatusOK {
		t.Errorf("expected %d after the slot was released, got %d", http.StatusOK, after.Code)
	}
}

func TestExportGuardReleasesSlotOnPanic(t *testing.T) {
	guard := newExportGuard(1)
	handler := guard.wrap(func(http.ResponseWriter, *http.Request, httprouter.Params) {
		panic("boom")
	})

	func() {
		defer func() { _ = recover() }()
		handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/pdf", nil), nil)
	}()

	if busy := len(guard.sem); busy != 0 {
		t.Errorf("expected the slot to be released after a panic, %d still busy", busy)
	}
}

func TestExportGuardDisabled(t *testing.T) {
	guard := newExportGuard(0)
	called := 0
	handler := guard.wrap(func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		called++
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, "/pdf", nil), nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("call %d: expected %d, got %d", i+1, http.StatusOK, rec.Code)
		}
	}
	if called != 3 {
		t.Errorf("expected the handler to run 3 times, got %d", called)
	}
}
