package cmd

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var runtimeInfo, _ = debug.ReadBuildInfo()
var runtimeModuleInfo = strings.Split(runtimeInfo.Main.Path, "/")
var runtimeModuleName = runtimeModuleInfo[len(runtimeModuleInfo)-1]

// Bound by the root command's persistent flags. Consumed starting in Phase 1
// (config resolution) and Phase 2 (logger dev-mode override) — Phase 0 only
// wires the CLI surface.
var (
	devMode  bool
	envMode  bool
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   runtimeModuleName,
	Short: runtimeModuleName + " cli",
	Long: `
  ____            _    ____ ___
 / ___| ___      / \  |  _ \_ _|
| |  _ / _ \    / _ \ | |_) | |
| |_| | (_) |  / ___ \|  __/| |
 \____|\___/  /_/   \_\_|  |___|
	` + runtimeModuleName + ` cli`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&devMode, "dev", "d", false, "Run in development mode")
	rootCmd.PersistentFlags().BoolVarP(&envMode, "env", "e", false, "Print environment before execution")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log", "l", "", "Log level")
}
