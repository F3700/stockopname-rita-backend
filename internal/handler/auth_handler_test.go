package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"

	"stockopname-rita-backend/internal/auth"
)

const (
	testUsername = "admin"
	testPassword = "rahasia-gudang"
)

// newTestAuthHandler hashes the password once per test: PBKDF2 is deliberately
// slow, so sharing the hash keeps the suite quick.
func newTestAuthHandler(t *testing.T, opts ...AuthOption) *AuthHandler {
	t.Helper()

	hash, err := auth.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("hashing the password: %v", err)
	}

	manager, err := auth.NewManager([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	if err != nil {
		t.Fatalf("building the manager: %v", err)
	}

	options := append([]AuthOption{WithCookieAttributes(false, "")}, opts...)

	return NewAuthHandler(manager, testUsername, hash, options...)
}

// sessionCookie returns the cookie the handler set, failing the test when there
// is none.
func sessionCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == auth.CookieName {
			return cookie
		}
	}

	t.Fatal("expected a session cookie")
	return nil
}

func TestLoginIssuesSessionWithoutOperator(t *testing.T) {
	handler := newTestAuthHandler(t)

	recorder := httptest.NewRecorder()
	handler.Login(recorder,
		newJSONRequest(t, http.MethodPost, "/auth/login", `{"username":"admin","password":"rahasia-gudang"}`),
		httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	cookie := sessionCookie(t, recorder)
	if !cookie.HttpOnly {
		t.Error("the session cookie must be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("expected SameSite Lax, got %v", cookie.SameSite)
	}
	if cookie.MaxAge != int(time.Hour.Seconds()) {
		t.Errorf("expected a one hour cookie, got %d seconds", cookie.MaxAge)
	}
	if cookie.Secure {
		t.Error("the test handler is built for plain http")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"authenticated":true`) || !strings.Contains(body, `"operator":""`) {
		t.Errorf("unexpected body %s", body)
	}
}

func TestLoginRejectsWrongCredentials(t *testing.T) {
	tests := map[string]string{
		"wrong password": `{"username":"admin","password":"salah"}`,
		"wrong username": `{"username":"guest","password":"rahasia-gudang"}`,
		"empty body":     `{"username":"","password":""}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			handler := newTestAuthHandler(t)

			recorder := httptest.NewRecorder()
			handler.Login(recorder, newJSONRequest(t, http.MethodPost, "/auth/login", body), httprouter.Params{})

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", recorder.Code)
			}
			if len(recorder.Result().Cookies()) != 0 {
				t.Error("a failed login must not set a cookie")
			}
		})
	}
}

func TestLoginRejectsMalformedBody(t *testing.T) {
	handler := newTestAuthHandler(t)

	recorder := httptest.NewRecorder()
	handler.Login(recorder, newJSONRequest(t, http.MethodPost, "/auth/login", `{"username":123}`), httprouter.Params{})

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestLoginRateLimited(t *testing.T) {
	handler := newTestAuthHandler(t, WithLoginLimiter(auth.NewLimiter(1, time.Minute)))

	wrong := newJSONRequest(t, http.MethodPost, "/auth/login", `{"username":"admin","password":"salah"}`)
	first := httptest.NewRecorder()
	handler.Login(first, wrong, httprouter.Params{})

	if first.Code != http.StatusUnauthorized {
		t.Fatalf("expected the first attempt to be a plain 401, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.Login(second, newJSONRequest(t, http.MethodPost, "/auth/login", `{"username":"admin","password":"salah"}`), httprouter.Params{})

	if second.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 on the second attempt, got %d", second.Code)
	}
}

func TestSetOperatorStoresNormalizedName(t *testing.T) {
	handler := newTestAuthHandler(t)

	recorder := httptest.NewRecorder()
	handler.SetOperator(recorder,
		newJSONRequest(t, http.MethodPost, "/auth/operator", `{"name":"  Niko   Santoso "}`),
		httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	cookie := sessionCookie(t, recorder)
	claims, err := handler.Manager.Parse(cookie.Value)
	if err != nil {
		t.Fatalf("the new cookie must be valid: %v", err)
	}
	if claims.Operator != "Niko Santoso" {
		t.Errorf("expected the name to be tidied to %q, got %q", "Niko Santoso", claims.Operator)
	}
}

func TestSetOperatorRejectsInvalidNames(t *testing.T) {
	tests := map[string]string{
		"empty":      `{"name":""}`,
		"only space": `{"name":"   "}`,
		"too long":   `{"name":"` + strings.Repeat("a", 51) + `"}`,
		"control":    `{"name":"Niko\u0000"}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			handler := newTestAuthHandler(t)

			recorder := httptest.NewRecorder()
			handler.SetOperator(recorder, newJSONRequest(t, http.MethodPost, "/auth/operator", body), httprouter.Params{})

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if len(recorder.Result().Cookies()) != 0 {
				t.Error("a rejected name must not refresh the cookie")
			}
		})
	}
}

func TestMeWithoutCookie(t *testing.T) {
	handler := newTestAuthHandler(t)

	recorder := httptest.NewRecorder()
	handler.Me(recorder, newJSONRequest(t, http.MethodGet, "/auth/me", ""), httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"authenticated":false`) {
		t.Errorf("unexpected body %s", body)
	}
}

func TestMeWithBrokenCookie(t *testing.T) {
	handler := newTestAuthHandler(t)

	request := newJSONRequest(t, http.MethodGet, "/auth/me", "")
	request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: "garbage"})

	recorder := httptest.NewRecorder()
	handler.Me(recorder, request, httprouter.Params{})

	if body := recorder.Body.String(); !strings.Contains(body, `"authenticated":false`) {
		t.Errorf("unexpected body %s", body)
	}
}

func TestMeReportsOperator(t *testing.T) {
	handler := newTestAuthHandler(t)

	request := newJSONRequest(t, http.MethodGet, "/auth/me", "")
	request.AddCookie(&http.Cookie{Name: auth.CookieName, Value: handler.Manager.Issue("Niko")})

	recorder := httptest.NewRecorder()
	handler.Me(recorder, request, httprouter.Params{})

	body := recorder.Body.String()
	if !strings.Contains(body, `"authenticated":true`) || !strings.Contains(body, `"operator":"Niko"`) {
		t.Errorf("unexpected body %s", body)
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	handler := newTestAuthHandler(t)

	recorder := httptest.NewRecorder()
	handler.Logout(recorder, newJSONRequest(t, http.MethodPost, "/auth/logout", ""), httprouter.Params{})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	cookie := sessionCookie(t, recorder)
	if cookie.MaxAge >= 0 {
		t.Errorf("expected the cookie to be expired, got MaxAge %d", cookie.MaxAge)
	}
	if cookie.Value != "" {
		t.Errorf("expected an empty value, got %q", cookie.Value)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		realIP     string
		forwarded  string
		remoteAddr string
		want       string
	}{
		{name: "real ip wins", realIP: "203.0.113.9", forwarded: "198.51.100.1", remoteAddr: "127.0.0.1:5555", want: "203.0.113.9"},
		{name: "forwarded first hop", forwarded: "198.51.100.1, 10.0.0.2", remoteAddr: "127.0.0.1:5555", want: "198.51.100.1"},
		{name: "socket address", remoteAddr: "192.0.2.7:5555", want: "192.0.2.7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
			request.RemoteAddr = tt.remoteAddr
			if tt.realIP != "" {
				request.Header.Set("X-Real-IP", tt.realIP)
			}
			if tt.forwarded != "" {
				request.Header.Set("X-Forwarded-For", tt.forwarded)
			}

			if got := clientIP(request); got != tt.want {
				t.Errorf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
