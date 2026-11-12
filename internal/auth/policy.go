package auth

import (
	"strings"
)

// publicRoute is one method and path pair reachable without a session.
type publicRoute struct {
	method string
	path   string
}

// publicRoutes is the mobile surface plus the auth endpoints themselves. The
// scanner app has no login yet, so each of these must keep answering without a
// cookie; see internal/auth/policy_test.go for the pairs that stay closed.
var publicRoutes = []publicRoute{
	// Probes are polled by the orchestrator, which has no cookie.
	{method: "GET", path: "/health"},
	{method: "GET", path: "/health/ready"},

	// Login has to be reachable, and /auth/me is how the SPA asks whether the
	// cookie it already has is still good.
	{method: "POST", path: "/auth/login"},
	{method: "GET", path: "/auth/me"},

	// Mobile, unchanged by this phase.
	{method: "POST", path: "/stockopname/coordinators/join-by-qr"},
	{method: "GET", path: "/stockopname/coordinators/:id"},
	{method: "GET", path: "/stockopname/sessions/:id"},
	{method: "GET", path: "/products"},
	{method: "GET", path: "/stockopname/racks"},
	{method: "POST", path: "/stockopname/uploads"},
}

// docsPrefix is the Swagger UI: static HTML plus the spec it loads.
const docsPrefix = "/docs"

// IsPublic reports whether method and path may be served without a session.
func IsPublic(method, path string) bool {
	if path == docsPrefix || strings.HasPrefix(path, docsPrefix+"/") {
		return true
	}

	for _, route := range publicRoutes {
		if route.method == method && matchPath(route.path, path) {
			return true
		}
	}

	return false
}

// IsSessionOnly reports the routes that need a valid session but not yet a
// name: the operator screen and logout both happen before the app is usable.
func IsSessionOnly(path string) bool {
	return path == "/auth/operator" || path == "/auth/logout"
}

// matchPath compares a registered pattern with a request path segment by
// segment. A ":name" segment matches any non-empty segment; everything else has
// to be equal. Query strings never reach this function.
func matchPath(pattern, path string) bool {
	patternSegments := splitPath(pattern)
	pathSegments := splitPath(path)

	if len(patternSegments) != len(pathSegments) {
		return false
	}

	for i, segment := range patternSegments {
		if strings.HasPrefix(segment, ":") {
			if pathSegments[i] == "" {
				return false
			}
			continue
		}

		if segment != pathSegments[i] {
			return false
		}
	}

	return true
}

func splitPath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}
