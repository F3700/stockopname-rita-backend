package dto

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SetOperatorRequest carries the free-text name typed after login. It is not
// verified: see the honesty note in the plan, the name is a habit trail, not
// proof of identity.
type SetOperatorRequest struct {
	Name string `json:"name"`
}

// AuthState is what every /auth endpoint reports back. Operator is empty until
// the user fills the operator screen.
type AuthState struct {
	Authenticated bool   `json:"authenticated"`
	Operator      string `json:"operator"`
}
