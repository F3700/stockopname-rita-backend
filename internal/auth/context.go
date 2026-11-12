package auth

import "context"

type actorKey struct{}

// WithActor attaches the operator name proven by the session cookie.
func WithActor(ctx context.Context, operator string) context.Context {
	return context.WithValue(ctx, actorKey{}, operator)
}

// ActorFromContext returns the name RequireAuth attached, or "" when the
// request arrived without a session (the mobile endpoints).
func ActorFromContext(ctx context.Context) string {
	operator, _ := ctx.Value(actorKey{}).(string)
	return operator
}
