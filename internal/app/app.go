package app

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/DanielKirkwood/go-api/internal/clients/logger"
	"github.com/DanielKirkwood/go-api/internal/clients/otelclient"
	"github.com/DanielKirkwood/go-api/internal/config"
)

// App holds every dependency constructed at startup: resolved config, the
// feature registry, the logger, and the otel providers. Bootstrap builds
// one; nothing here is a package-level singleton.
type App struct {
	Env      config.EnvVars
	Registry *config.Registry
	Logger   *slog.Logger
	Otel     *otelclient.Providers
}

// Bootstrap resolves env's feature-enabled state and configures every
// enabled feature, then constructs the logger and otel providers from the
// result. It deliberately does not call Registry.ValidateReadiness: that
// stays an opt-in step for commands (such as `info features`) that need
// full readiness, so read-only commands keep working even if a feature is
// misconfigured.
func Bootstrap(ctx context.Context, env config.EnvVars) (*App, error) {
	registry := config.DefaultRegistry()
	if err := registry.ResolveFeatureEnabledState(env); err != nil {
		return nil, fmt.Errorf("app: resolve features: %w", err)
	}
	if err := registry.Configure(env); err != nil {
		return nil, fmt.Errorf("app: configure features: %w", err)
	}

	log, err := logger.New(logger.Config{
		Dev:   env.Env == "development",
		Level: env.LogLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("app: build logger: %w", err)
	}

	otelFeature, _ := registry.Feature("otel")
	otelCfg, _ := otelFeature.Config.(config.OtelConfig)

	providers, err := otelclient.New(ctx, otelCfg, otelFeature.Enabled, buildInfo())
	if err != nil {
		return nil, fmt.Errorf("app: bootstrap otel: %w", err)
	}

	log.DebugContext(ctx, "bootstrap complete", "env", env.Env, "otel_enabled", otelFeature.Enabled)

	return &App{Env: env, Registry: registry, Logger: log, Otel: providers}, nil
}

// Shutdown flushes and closes every dependency that needs it. Currently
// that's just the otel providers; Phase 3 wires this into the real
// signal-driven shutdown path.
func (a *App) Shutdown(ctx context.Context) error {
	return a.Otel.Shutdown(ctx)
}

// buildInfo derives the otel resource's service name/version from the
// running binary's build info, the same source cmd's moduleName() uses.
func buildInfo() otelclient.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return otelclient.BuildInfo{Name: "go-api"}
	}

	parts := strings.Split(info.Main.Path, "/")

	return otelclient.BuildInfo{Name: parts[len(parts)-1], Version: info.Main.Version}
}
