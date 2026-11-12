package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stockopname-rita-backend/internal/auth"
)

func newTestManager(t *testing.T) *auth.Manager {
	t.Helper()

	manager, err := auth.NewManager([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	if err != nil {
		t.Fatalf("building the manager: %v", err)
	}

	return manager
}

// serve runs one request through RequireAuth and reports the status plus the
// actor the inner handler saw.
func serve(manager *auth.Manager, method, path, token string) (int, string) {
	var seenActor string

	handler := RequireAuth(manager, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		seenActor = auth.ActorFromContext(req.Context())
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(method, path, nil)
	if token != "" {
		request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return recorder.Code, seenActor
}

func TestRequireAuthRejectsMissingCookie(t *testing.T) {
	code, _ := serve(newTestManager(t), http.MethodGet, "/stockopname/sessions", "")

	if code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestRequireAuthRejectsBrokenCookie(t *testing.T) {
	code, _ := serve(newTestManager(t), http.MethodGet, "/stockopname/sessions", "not-a-token")

	if code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestRequireAuthRejectsUnknownOperator(t *testing.T) {
	manager := newTestManager(t)

	code, _ := serve(manager, http.MethodGet, "/stockopname/sessions", manager.Issue(""))

	if code != http.StatusUnauthorized {
		t.Errorf("expected 401 before a name is typed, got %d", code)
	}
}

func TestRequireAuthPassesOperatorToContext(t *testing.T) {
	manager := newTestManager(t)

	code, actor := serve(manager, http.MethodPost, "/stockopname/sessions", manager.Issue("Niko"))

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if actor != "Niko" {
		t.Errorf("expected the operator in the context, got %q", actor)
	}
}

func TestRequireAuthAllowsOperatorScreenBeforeName(t *testing.T) {
	manager := newTestManager(t)

	for _, path := range []string{"/auth/operator", "/auth/logout"} {
		code, _ := serve(manager, http.MethodPost, path, manager.Issue(""))
		if code != http.StatusOK {
			t.Errorf("%s must work before a name is set, got %d", path, code)
		}
	}
}

func TestRequireAuthStillNeedsSessionForOperatorScreen(t *testing.T) {
	code, _ := serve(newTestManager(t), http.MethodPost, "/auth/operator", "")

	if code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestRequireAuthLetsPublicRoutesThrough(t *testing.T) {
	manager := newTestManager(t)

	tests := []struct{ method, path string }{
		{http.MethodGet, "/health"},
		{http.MethodGet, "/health/ready"},
		{http.MethodPost, "/auth/login"},
		{http.MethodGet, "/auth/me"},
		{http.MethodPost, "/stockopname/coordinators/join-by-qr"},
		{http.MethodGet, "/stockopname/coordinators/1"},
		{http.MethodGet, "/stockopname/sessions/1"},
		{http.MethodGet, "/products"},
		{http.MethodGet, "/stockopname/racks"},
		{http.MethodPost, "/stockopname/uploads"},
		{http.MethodGet, "/docs/apispec.json"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			if code, _ := serve(manager, tt.method, tt.path, ""); code != http.StatusOK {
				t.Errorf("expected the public route to pass, got %d", code)
			}
		})
	}
}

// The reply carries the same envelope as the rest of the API so the SPA can
// surface the message without a special case.
func TestRequireAuthWritesJsonEnvelope(t *testing.T) {
	handler := RequireAuth(newTestManager(t), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/stockopname/sessions", nil))

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected a JSON body, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), `"message"`) {
		t.Errorf("unexpected body %s", recorder.Body.String())
	}
}
