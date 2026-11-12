package auth

import "testing"

func TestIsPublic(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{"liveness", "GET", "/health", true},
		{"readiness", "GET", "/health/ready", true},
		{"login", "POST", "/auth/login", true},
		{"me", "GET", "/auth/me", true},

		// Method matters: the same path with a mutating verb stays closed.
		{"health as post", "POST", "/health", false},
		{"login as get", "GET", "/auth/login", false},

		// Mobile surface.
		{"join by qr", "POST", "/stockopname/coordinators/join-by-qr", true},
		{"coordinator detail", "GET", "/stockopname/coordinators/12", true},
		{"session detail", "GET", "/stockopname/sessions/7", true},
		{"products", "GET", "/products", true},
		{"racks", "GET", "/stockopname/racks", true},
		{"upload", "POST", "/stockopname/uploads", true},

		// Admin-only siblings of the public routes.
		{"product detail", "GET", "/products/1", false},
		{"product create", "POST", "/products", false},
		{"coordinator list", "GET", "/stockopname/coordinators", false},
		{"session list", "GET", "/stockopname/sessions", false},
		{"rack progress", "GET", "/stockopname/racks/progress", false},
		{"upload list", "GET", "/stockopname/uploads", false},
		{"results", "GET", "/stockopname/results", false},
		{"audits", "GET", "/stockopname/audits", false},

		// Auth endpoints that must not be reachable without a cookie.
		{"operator", "POST", "/auth/operator", false},
		{"logout", "POST", "/auth/logout", false},

		// Docs.
		{"docs redirect", "GET", "/docs", true},
		{"docs ui", "GET", "/docs/", true},
		{"docs spec", "GET", "/docs/apispec.json", true},
		{"docs lookalike", "GET", "/docsomething", false},

		{"unknown", "GET", "/nope", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPublic(tt.method, tt.path); got != tt.want {
				t.Errorf("IsPublic(%q, %q) = %v, want %v", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

func TestIsSessionOnly(t *testing.T) {
	tests := map[string]bool{
		"/auth/operator": true,
		"/auth/logout":   true,
		"/auth/login":    false,
		"/auth/me":       false,
		"/products":      false,
	}

	for path, want := range tests {
		if got := IsSessionOnly(path); got != want {
			t.Errorf("IsSessionOnly(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"/products", "/products", true},
		{"/products", "/products/", true},
		{"/products", "/products/1", false},
		{"/stockopname/sessions/:id", "/stockopname/sessions/12", true},
		{"/stockopname/sessions/:id", "/stockopname/sessions", false},
		{"/stockopname/sessions/:id", "/stockopname/sessions/12/pdf", false},
	}

	for _, tt := range tests {
		if got := matchPath(tt.pattern, tt.path); got != tt.want {
			t.Errorf("matchPath(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}
