package cmd

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/DanielKirkwood/go-api/internal/app"
	"github.com/DanielKirkwood/go-api/internal/config"
	"github.com/spf13/cobra"
)

// Execute builds the root command and runs it. It is called by main.main().
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// NewRootCmd builds the CLI's root command and its full subcommand tree.
// Flags are bound to local variables captured by PersistentPreRunE's
// closure rather than package-level state.
func NewRootCmd() *cobra.Command {
	var (
		devMode  bool
		envMode  bool
		logLevel string
	)

	name := moduleName()

	rootCmd := &cobra.Command{
		Use:   name,
		Short: name + " cli",
		Long: `
  ____            _    ____ ___
 / ___| ___      / \  |  _ \_ _|
| |  _ / _ \    / _ \ | |_) | |
| |_| | (_) |  / ___ \|  __/| |
 \____|\___/  /_/   \_\_|  |___|
	` + name + ` cli`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			env, err := config.Load()
			if err != nil {
				return err
			}

			if devMode {
				env.Env = "development"
			}
			if logLevel != "" {
				env.LogLevel = logLevel
			}
			if envMode {
				return printEnv(cmd, env)
			}

			a, err := app.Bootstrap(cmd.Context(), env)
			if err != nil {
				return err
			}
			cmd.SetContext(withApp(cmd.Context(), a))

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			a, ok := appFromContext(cmd.Context())
			if !ok {
				return nil
			}

			return a.Shutdown(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&devMode, "dev", "d", false, "Run in development mode")
	rootCmd.PersistentFlags().BoolVarP(&envMode, "env", "e", false, "Print environment before execution")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log", "l", "", "Log level")

	rootCmd.AddCommand(NewInfoCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewDBCmd())
	rootCmd.AddCommand(NewTaskCmd())

	return rootCmd
}

// moduleName derives the CLI's display name from the running binary's build
// info, falling back to a fixed name if build info isn't available.
func moduleName() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "go-api"
	}

	parts := strings.Split(info.Main.Path, "/")

	return parts[len(parts)-1]
}
