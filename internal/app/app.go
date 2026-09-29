package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/DanielKirkwood/go-api/internal/clients/ketoclient"
	"github.com/DanielKirkwood/go-api/internal/clients/kratosclient"
	"github.com/DanielKirkwood/go-api/internal/clients/logger"
	"github.com/DanielKirkwood/go-api/internal/clients/otelclient"
	"github.com/DanielKirkwood/go-api/internal/config"
	"github.com/DanielKirkwood/go-api/internal/db"
)

// App holds every dependency constructed at startup: resolved config, the
// feature registry, the logger, and the otel providers. Bootstrap builds
// one; nothing here is a package-level singleton.
//
// Servers is nil until a command that serves HTTP (start and its
// subcommands) calls BuildServers and assigns the result — commands like
// `info` never need routers built.
type App struct {
	Env      config.EnvVars
	Registry *config.Registry
	Logger   *slog.Logger
	Otel     *otelclient.Providers
	Store    *db.Store
	Kratos   *kratosclient.Client
	Keto     *ketoclient.Client
	Servers  *Servers
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

	dbFeature, _ := registry.Feature("database")
	dbCfg, _ := dbFeature.Config.(config.DatabaseConfig)

	store, err := db.New(dbCfg, dbFeature.Enabled, providers.TracerProvider, providers.MeterProvider)
	if err != nil {
		return nil, fmt.Errorf("app: bootstrap database: %w", err)
	}

	kratosFeature, _ := registry.Feature("kratos")
	kratosCfg, _ := kratosFeature.Config.(config.KratosConfig)

	kratos, err := kratosclient.New(kratosCfg, kratosFeature.Enabled)
	if err != nil {
		return nil, fmt.Errorf("app: bootstrap kratos: %w", err)
	}

	ketoFeature, _ := registry.Feature("keto")
	ketoCfg, _ := ketoFeature.Config.(config.KetoConfig)

	keto, err := ketoclient.New(ketoCfg, ketoFeature.Enabled)
	if err != nil {
		return nil, fmt.Errorf("app: bootstrap keto: %w", err)
	}

	log.DebugContext(ctx, "bootstrap complete",
		"env", env.Env, "otel_enabled", otelFeature.Enabled, "database_enabled", dbFeature.Enabled,
		"kratos_enabled", kratosFeature.Enabled, "keto_enabled", ketoFeature.Enabled)

	return &App{
		Env: env, Registry: registry, Logger: log, Otel: providers,
		Store: store, Kratos: kratos, Keto: keto,
	}, nil
}

// Shutdown gracefully closes every running server (if any), then the Store,
// then flushes the otel providers. Servers close first so no in-flight
// request's spans/metrics/queries are lost to a premature Store/otel
// shutdown; the Store closes before otel so its instrumentation can still
// flush.
func (a *App) Shutdown(ctx context.Context) error {
	if a.Servers != nil {
		if err := a.Servers.Shutdown(ctx); err != nil {
			return errors.Join(err, a.Store.Close(), a.Otel.Shutdown(ctx))
		}
	}

	return errors.Join(a.Store.Close(), a.Otel.Shutdown(ctx))
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
