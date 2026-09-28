package cmd

import (
	"context"

	"github.com/DanielKirkwood/go-api/internal/app"
)

// appContextKey is an unexported type so no other package can collide with
// this key, mirroring the identity-in-context convention planned for the
// auth middleware.
type appContextKey struct{}

// withApp returns a copy of ctx carrying a, retrievable via appFromContext.
func withApp(ctx context.Context, a *app.App) context.Context {
	return context.WithValue(ctx, appContextKey{}, a)
}

// appFromContext returns the *app.App stored on ctx by withApp, if any.
func appFromContext(ctx context.Context) (*app.App, bool) {
	a, ok := ctx.Value(appContextKey{}).(*app.App)
	return a, ok
}
