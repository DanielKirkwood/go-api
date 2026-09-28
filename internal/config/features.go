package config

// Feature config stubs. These only hold data in Phase 1 — the subsystems
// that consume them (routing/otel in Phase 2-3, database in Phase 4, Kratos
// in Phase 5, Keto in Phase 6) are built out in later phases.

// ServiceConfig is the always-on core app config.
type ServiceConfig struct {
	PublicPort    string
	ProtectedPort string
	HiddenPort    string
}

// DatabaseConfig configures the SQLite connection (Phase 4).
type DatabaseConfig struct {
	Path string
}

// OtelConfig configures the OpenTelemetry SDK bootstrap (Phase 2).
type OtelConfig struct {
	ExporterURL string
}

// KratosConfig configures the Ory Kratos client (Phase 5).
type KratosConfig struct {
	PublicURL string
	AdminURL  string
}

// KetoConfig configures the Ory Keto client (Phase 6).
type KetoConfig struct {
	ReadURL  string
	WriteURL string
}

// DefaultRegistry returns a Registry with every known feature registered,
// unresolved (call ResolveFeatureEnabledState, then Configure, then
// ValidateReadiness on the result).
func DefaultRegistry() *Registry {
	r := NewRegistry()

	r.Register(Feature{
		Name: "service",
		Build: func(e EnvVars) any {
			return ServiceConfig{
				PublicPort:    e.PublicPort,
				ProtectedPort: e.ProtectedPort,
				HiddenPort:    e.HiddenPort,
			}
		},
	})

	r.Register(Feature{
		Name:        "database",
		RequiredEnv: []string{"DatabasePath"},
		Build: func(e EnvVars) any {
			return DatabaseConfig{Path: e.DatabasePath}
		},
	})

	r.Register(Feature{
		Name:        "otel",
		RequiredEnv: []string{"OtelExporterURL"},
		Build: func(e EnvVars) any {
			return OtelConfig{ExporterURL: e.OtelExporterURL}
		},
	})

	r.Register(Feature{
		Name:        "kratos",
		RequiredEnv: []string{"KratosPublicURL", "KratosAdminURL"},
		Build: func(e EnvVars) any {
			return KratosConfig{PublicURL: e.KratosPublicURL, AdminURL: e.KratosAdminURL}
		},
	})

	r.Register(Feature{
		Name:        "keto",
		RequiredEnv: []string{"KetoReadURL", "KetoWriteURL"},
		Requires:    []string{"kratos"},
		Build: func(e EnvVars) any {
			return KetoConfig{ReadURL: e.KetoReadURL, WriteURL: e.KetoWriteURL}
		},
	})

	return r
}
