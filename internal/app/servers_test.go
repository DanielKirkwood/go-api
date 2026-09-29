package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielKirkwood/go-api/internal/app"
	"github.com/DanielKirkwood/go-api/internal/config"
)

func TestBuildServers_KratosDisabled_IdentitiesNotMounted(t *testing.T) {
	t.Parallel()

	env := config.EnvVars{Env: "development", LogLevel: "debug"}

	a, err := app.Bootstrap(t.Context(), env)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v, want nil", err)
	}

	servers, err := app.BuildServers(a)
	if err != nil {
		t.Fatalf("BuildServers() error = %v, want nil", err)
	}

	rec := httptest.NewRecorder()
	servers.Hidden.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/identities", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /admin/identities status = %d, want 404 (kratos disabled, route never mounted)", rec.Code)
	}
}

func TestBuildServers_KratosEnabled_HiddenRouterRequiresAuth(t *testing.T) {
	t.Parallel()

	env := config.EnvVars{
		Env:             "production",
		LogLevel:        "info",
		KratosPublicURL: "http://127.0.0.1:4433",
		KratosAdminURL:  "http://127.0.0.1:4434",
	}

	a, err := app.Bootstrap(t.Context(), env)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v, want nil", err)
	}

	servers, err := app.BuildServers(a)
	if err != nil {
		t.Fatalf("BuildServers() error = %v, want nil", err)
	}

	// No Cookie header: AuthenticationMiddleware rejects before ever
	// reaching Kratos, so this needs no live Kratos instance to assert 401.
	rec := httptest.NewRecorder()
	servers.Hidden.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/identities", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /admin/identities status = %d, want 401 (no session cookie)", rec.Code)
	}

	// /health/* stays exempt even with kratos enabled.
	healthRec := httptest.NewRecorder()
	servers.Hidden.Handler.ServeHTTP(healthRec, httptest.NewRequest(http.MethodGet, "/health/alive", nil))

	if healthRec.Code != http.StatusOK {
		t.Errorf("GET /health/alive status = %d, want 200", healthRec.Code)
	}
}
