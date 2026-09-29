package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DanielKirkwood/go-api/internal/db"
)

var errDatabaseNotConfigured = errors.New("cmd: database feature is not enabled")

// NewDBCmd builds the `db` command group: migration and seed management
// against the bootstrapped app's Store.
func NewDBCmd() *cobra.Command {
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Manage the database: migrations and seed data",
	}

	dbCmd.AddCommand(newDBCreateCmd())
	dbCmd.AddCommand(newDBMigrateCmd())
	dbCmd.AddCommand(newDBRollbackCmd())
	dbCmd.AddCommand(newDBStatusCmd())
	dbCmd.AddCommand(newDBSeedCmd())

	return dbCmd
}

func newDBCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Scaffold a new empty migration file",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return db.Create(args[0])
		},
	}
}

func newDBMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Apply every pending migration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := storeFromContext(cmd)
			if err != nil {
				return err
			}

			return db.Migrate(cmd.Context(), store.DB)
		},
	}
}

func newDBRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Revert the most recently applied migration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := storeFromContext(cmd)
			if err != nil {
				return err
			}

			return db.Rollback(cmd.Context(), store.DB)
		},
	}
}

func newDBStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show which migrations are applied vs pending",
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := storeFromContext(cmd)
			if err != nil {
				return err
			}

			return db.Status(cmd.Context(), store.DB)
		},
	}
}

func newDBSeedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with development data",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if _, err := storeFromContext(cmd); err != nil {
				return err
			}

			// Seed data is introduced alongside the first real domain
			// table in Phase 7 — nothing to seed yet.
			fmt.Fprintln(cmd.OutOrStdout(), "db seed: nothing to seed yet")

			return nil
		},
	}
}

// storeFromContext returns the bootstrapped app's *db.Store, erroring if the
// app isn't bootstrapped or the database feature is disabled.
func storeFromContext(cmd *cobra.Command) (*db.Store, error) {
	a, ok := appFromContext(cmd.Context())
	if !ok {
		return nil, errAppNotBootstrapped
	}
	if a.Store == nil {
		return nil, errDatabaseNotConfigured
	}

	return a.Store, nil
}
