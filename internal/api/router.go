package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel/trace"
)

const requestTimeout = 30 * time.Second

// RouterDeps holds everything a router constructor needs. It's built once
// in app and passed to each of NewPublicRouter/NewProtectedRouter/
// NewHiddenRouter — no router reaches for a package-level dependency.
type RouterDeps struct {
	Logger         *slog.Logger
	TracerProvider trace.TracerProvider
	Ready          ReadyCheck
}

// NewPublicRouter builds the public router: the externally-facing surface.
func NewPublicRouter(deps RouterDeps) *chi.Mux {
	return newRouter("public", deps)
}

// NewProtectedRouter builds the protected router: authenticated but
// non-admin endpoints (Phase 5 adds the auth middleware here).
func NewProtectedRouter(deps RouterDeps) *chi.Mux {
	return newRouter("protected", deps)
}

// NewHiddenRouter builds the hidden router: admin-only endpoints, meant to
// be bound to an interface not exposed to the public internet.
func NewHiddenRouter(deps RouterDeps) *chi.Mux {
	return newRouter("hidden", deps)
}

func newRouter(name string, deps RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	r.NotFound(problemHandler(Problem{Status: http.StatusNotFound, Title: "Not Found"}))
	r.MethodNotAllowed(problemHandler(Problem{Status: http.StatusMethodNotAllowed, Title: "Method Not Allowed"}))

	r.Use(middleware.RequestID)
	r.Use(RequestLogger(deps.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))
	r.Use(EnforceJSON)
	r.Use(SecurityHeaders)
	r.Use(otelchi.Middleware(name, otelchi.WithChiRoutes(r), otelchi.WithTracerProvider(deps.TracerProvider)))

	MountHealth(r, deps.Ready)

	return r
}

// problemHandler returns an [http.HandlerFunc] that always writes p, for
// use as a router's NotFound/MethodNotAllowed handler.
func problemHandler(p Problem) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_ = WriteProblem(w, p)
	}
}
