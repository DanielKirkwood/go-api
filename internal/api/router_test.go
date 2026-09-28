package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/DanielKirkwood/go-api/internal/api"
)

func TestRouters_HealthAndSecurityHeaders(t *testing.T) {
	t.Parallel()

	deps := api.RouterDeps{Logger: slog.Default(), TracerProvider: noop.NewTracerProvider()}

	constructors := map[string]func(api.RouterDeps) http.Handler{
		"public":    func(d api.RouterDeps) http.Handler { return api.NewPublicRouter(d) },
		"protected": func(d api.RouterDeps) http.Handler { return api.NewProtectedRouter(d) },
		"hidden":    func(d api.RouterDeps) http.Handler { return api.NewHiddenRouter(d) },
	}

	for name, build := range constructors {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			router := build(deps)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/alive", nil))

			if rec.Code != http.StatusOK {
				t.Errorf("GET /health/alive status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
		})
	}
}
