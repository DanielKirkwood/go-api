package config_test

import (
	"testing"

	"github.com/DanielKirkwood/go-api/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want config.EnvVars
	}{
		{
			name: "defaults when unset",
			env:  map[string]string{},
			want: config.EnvVars{Env: "development", LogLevel: "info", Port: "8080"},
		},
		{
			name: "env vars override defaults",
			env: map[string]string{
				"ENV":       "production",
				"LOG_LEVEL": "warn",
				"PORT":      "9090",
			},
			want: config.EnvVars{Env: "production", LogLevel: "warn", Port: "9090"},
		},
		{
			name: "optional fields have no default",
			env: map[string]string{
				"DATABASE_PATH": "/tmp/app.db",
			},
			want: config.EnvVars{
				Env: "development", LogLevel: "info", Port: "8080",
				DatabasePath: "/tmp/app.db",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := config.Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
