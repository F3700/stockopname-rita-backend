package handler

import (
	"crypto/subtle"
	"errors"
	"net"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"

	"stockopname-rita-backend/internal/auth"
	"stockopname-rita-backend/internal/dto"
	"stockopname-rita-backend/internal/helper"
)

const (
	maxOperatorNameLength = 50

	msgInvalidCredentials = "Username atau password salah."
	msgTooManyAttempts    = "Terlalu banyak percobaan. Coba lagi beberapa menit."
	msgOperatorRequired   = "Nama operator wajib diisi (1-50 karakter)."
	msgOperatorControl    = "Nama operator mengandung karakter yang tidak diizinkan."
	msgInvalidBody        = "Invalid request body"
)

var (
	errOperatorNameLength  = errors.New("auth: operator name must be 1-50 characters")
	errOperatorNameControl = errors.New("auth: operator name contains control characters")
)

// AuthHandler owns the four session endpoints. It needs no service and no
// database: the credentials come from the environment and the session lives
// entirely in a signed cookie.
type AuthHandler struct {
	Manager      *auth.Manager
	Username     string
	PasswordHash string
	Limiter      *auth.Limiter
	CookieSecure bool
	CookieDomain string
}

// AuthOption tunes an AuthHandler. Tests use them to shrink the login limit or
// turn the secure cookie off over plain HTTP.
type AuthOption func(*AuthHandler)

func WithCookieAttributes(secure bool, domain string) AuthOption {
	return func(h *AuthHandler) {
		h.CookieSecure = secure
		h.CookieDomain = domain
	}
}

func WithLoginLimiter(limiter *auth.Limiter) AuthOption {
	return func(h *AuthHandler) {
		h.Limiter = limiter
	}
}

func NewAuthHandler(manager *auth.Manager, username, passwordHash string, opts ...AuthOption) *AuthHandler {
	handler := &AuthHandler{
		Manager:      manager,
		Username:     username,
		PasswordHash: passwordHash,
		Limiter:      auth.NewLimiter(auth.LoginAttemptLimit, auth.LoginAttemptWindow),
	}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

// Login exchanges the shared credential for a session cookie. The cookie starts
// without an operator name: the app cannot be used until one is typed.
func (h *AuthHandler) Login(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	client := clientIP(req)

	if !h.Limiter.Allowed(client) {
		h.writeMessage(writer, http.StatusTooManyRequests, msgTooManyAttempts)
		return
	}

	var body dto.LoginRequest
	if err := helper.ReadFromRequestBody(req, &body); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, msgInvalidBody, err)
		return
	}

	// Both checks always run: skipping the password step for an unknown
	// username would answer faster and tell an attacker which half was wrong.
	usernameOK := subtle.ConstantTimeCompare([]byte(body.Username), []byte(h.Username)) == 1
	passwordOK := auth.VerifyPassword(h.PasswordHash, body.Password)

	if !usernameOK || !passwordOK {
		h.Limiter.Fail(client)
		h.writeMessage(writer, http.StatusUnauthorized, msgInvalidCredentials)
		return
	}

	h.Limiter.Reset(client)
	h.issueSession(writer, "", "Login berhasil")
}

// SetOperator stores the free-text name in a fresh cookie.
func (h *AuthHandler) SetOperator(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body dto.SetOperatorRequest
	if err := helper.ReadFromRequestBody(req, &body); err != nil {
		helper.WriteError(writer, http.StatusBadRequest, msgInvalidBody, err)
		return
	}

	name, err := normalizeOperatorName(body.Name)
	if err != nil {
		message := msgOperatorControl
		if errors.Is(err, errOperatorNameLength) {
			message = msgOperatorRequired
		}
		h.writeMessage(writer, http.StatusBadRequest, message)
		return
	}

	h.issueSession(writer, name, "Nama operator disimpan")
}

// Logout drops the cookie. It needs a session but not a name, so it still works
// from the operator screen.
func (h *AuthHandler) Logout(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	http.SetCookie(writer, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	h.writeState(writer, http.StatusOK, "Sesi berakhir", dto.AuthState{})
}

// Me reports whether the request carries a usable session. It is public on
// purpose: the SPA asks before it knows whether it can show anything.
func (h *AuthHandler) Me(writer http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	state := dto.AuthState{}

	if cookie, err := req.Cookie(auth.CookieName); err == nil {
		if claims, err := h.Manager.Parse(cookie.Value); err == nil {
			state.Authenticated = true
			state.Operator = claims.Operator
		}
	}

	h.writeState(writer, http.StatusOK, "Status sesi", state)
}

func (h *AuthHandler) issueSession(writer http.ResponseWriter, operator, message string) {
	http.SetCookie(writer, &http.Cookie{
		Name:     auth.CookieName,
		Value:    h.Manager.Issue(operator),
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   int(h.Manager.TTL().Seconds()),
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	h.writeState(writer, http.StatusOK, message, dto.AuthState{Authenticated: true, Operator: operator})
}

func (h *AuthHandler) writeState(writer http.ResponseWriter, status int, message string, state dto.AuthState) {
	if err := helper.ResponseJson(writer, status, dto.Response{Message: message, Data: state}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
	}
}

func (h *AuthHandler) writeMessage(writer http.ResponseWriter, status int, message string) {
	if err := helper.ResponseJson(writer, status, dto.Response{Message: message}); err != nil {
		helper.WriteError(writer, http.StatusInternalServerError, "Failed to send response", err)
	}
}

// normalizeOperatorName trims, collapses runs of whitespace and bounds the
// length. The name is free text by design; only characters that would break a
// log line or a CSV export are rejected.
func normalizeOperatorName(raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")

	length := utf8.RuneCountInString(name)
	if length < 1 || length > maxOperatorNameLength {
		return "", errOperatorNameLength
	}

	for _, char := range name {
		if unicode.IsControl(char) {
			return "", errOperatorNameControl
		}
	}

	return name, nil
}

// clientIP reads the address nginx reports, falling back to the socket. The
// headers are trustworthy only because the backend is reached through the
// proxy; a direct connection could claim any address.
func clientIP(req *http.Request) string {
	if realIP := strings.TrimSpace(req.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		first, _, _ := strings.Cut(forwarded, ",")
		if ip := strings.TrimSpace(first); ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}

	return host
}
