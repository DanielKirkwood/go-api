package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/DanielKirkwood/go-api/internal/config"
	"github.com/spf13/cobra"
)

const redacted = "<redacted>"

const (
	tabWriterMinWidth = 0
	tabWriterTabWidth = 4
	tabWriterPadding  = 2
	tabWriterPadChar  = ' '
)

// NewInfoCmd builds the `info` command group: read-only diagnostics over the
// resolved configuration.
func NewInfoCmd() *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Print diagnostic information about the resolved configuration",
	}

	infoCmd.AddCommand(newInfoEnvCmd())
	infoCmd.AddCommand(newInfoFeaturesCmd())

	return infoCmd
}

func newInfoEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Print resolved environment variables",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := config.Load()
			if err != nil {
				return err
			}

			return printEnv(cmd, env)
		},
	}
}

func newInfoFeaturesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "features",
		Short: "Print each feature's enabled/disabled state",
		RunE: func(cmd *cobra.Command, _ []string) error {
			env, err := config.Load()
			if err != nil {
				return err
			}

			return printFeatures(cmd, env)
		},
	}
}

func printEnv(cmd *cobra.Command, env config.EnvVars) error {
	w := newTabWriter(cmd)

	fmt.Fprintln(w, "ENV VAR\tVALUE")
	for _, f := range env.Fields() {
		value := f.Value
		if f.Secret && value != "" {
			value = redacted
		}
		fmt.Fprintf(w, "%s\t%s\n", f.Env, value)
	}

	return w.Flush()
}

func printFeatures(cmd *cobra.Command, env config.EnvVars) error {
	registry := config.DefaultRegistry()
	if err := registry.ResolveFeatureEnabledState(env); err != nil {
		return err
	}
	if err := registry.Configure(env); err != nil {
		return err
	}

	w := newTabWriter(cmd)

	fmt.Fprintln(w, "FEATURE\tENABLED\tREASON")
	for _, f := range registry.All() {
		fmt.Fprintf(w, "%s\t%t\t%s\n", f.Name, f.Enabled, f.Reason)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return registry.ValidateReadiness()
}

func newTabWriter(cmd *cobra.Command) *tabwriter.Writer {
	return tabwriter.NewWriter(
		cmd.OutOrStdout(), tabWriterMinWidth, tabWriterTabWidth, tabWriterPadding, tabWriterPadChar, 0,
	)
}
