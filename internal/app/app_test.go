package app_test

import (
	"testing"

	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/DanielKirkwood/go-api/internal/app"
	"github.com/DanielKirkwood/go-api/internal/config"
)

func TestBootstrapOtelDisabled(t *testing.T) {
	t.Parallel()

	env := config.EnvVars{Env: "development", LogLevel: "debug"}

	a, err := app.Bootstrap(t.Context(), env)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v, want nil", err)
	}

	if a.Logger == nil {
		t.Error("Logger = nil, want non-nil")
	}

	otelFeature, ok := a.Registry.Feature("otel")
	if !ok {
		t.Fatal("otel feature not registered")
	}
	if otelFeature.Enabled {
		t.Error("otel feature Enabled = true, want false (no OTEL_EXPORTER_OTLP_ENDPOINT set)")
	}

	if _, isNoopTracer := a.Otel.TracerProvider.(tracenoop.TracerProvider); !isNoopTracer {
		t.Errorf("TracerProvider = %T, want tracenoop.TracerProvider", a.Otel.TracerProvider)
	}
	if _, isNoopMeter := a.Otel.MeterProvider.(noop.MeterProvider); !isNoopMeter {
		t.Errorf("MeterProvider = %T, want noop.MeterProvider", a.Otel.MeterProvider)
	}

	if shutdownErr := a.Shutdown(t.Context()); shutdownErr != nil {
		t.Errorf("Shutdown() error = %v, want nil", shutdownErr)
	}
}

func TestBootstrapOtelEnabled(t *testing.T) {
	t.Parallel()

	env := config.EnvVars{
		Env:             "production",
		LogLevel:        "info",
		OtelExporterURL: "http://127.0.0.1:4318",
	}

	a, err := app.Bootstrap(t.Context(), env)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v, want nil", err)
	}

	otelFeature, ok := a.Registry.Feature("otel")
	if !ok {
		t.Fatal("otel feature not registered")
	}
	if !otelFeature.Enabled {
		t.Error("otel feature Enabled = false, want true (OTEL_EXPORTER_OTLP_ENDPOINT set)")
	}

	if a.Otel.TracerProvider == nil {
		t.Error("TracerProvider = nil, want non-nil")
	}
	if a.Otel.MeterProvider == nil {
		t.Error("MeterProvider = nil, want non-nil")
	}

	// Not asserting Shutdown() succeeds: it flushes over the network to
	// OtelExporterURL, which points at nothing reachable in this unit test.
	_ = a.Shutdown(t.Context())
}

func TestBootstrapUnknownLogLevel(t *testing.T) {
	t.Parallel()

	env := config.EnvVars{Env: "production", LogLevel: "trace"}

	if _, err := app.Bootstrap(t.Context(), env); err == nil {
		t.Fatal("Bootstrap() error = nil, want error for unknown log level")
	}
}
