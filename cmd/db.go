package cmd

import (
	"errors"
	"fmt"

	keto "github.com/ory/keto-client-go/v26"
	"github.com/spf13/cobra"

	"github.com/DanielKirkwood/go-api/internal/db"
)

var (
	errDatabaseNotConfigured           = errors.New("cmd: database feature is not enabled")
	errKetoSeedAdminIdentityIDRequired = errors.New(
		"cmd: KETO_SEED_ADMIN_IDENTITY_ID must be set to seed Keto relation tuples",
	)
)

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

			// SQLite domain data is introduced alongside the first real
			// domain table in Phase 7 — nothing to seed there yet. Keto
			// relation-tuple seeding lands in Phase 6, so it's the only
			// seed data today.
			if err := seedKeto(cmd); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "db seed: seeded Keto relation tuples")

			return nil
		},
	}
}

// seedKeto grants the Role:admin relation tuple to
// EnvVars.KetoSeedAdminIdentityID and makes Identities:admin's manage
// permission traverse that role, matching deploy/keto/identities.ts's OPL.
// It's a no-op when the keto feature is disabled.
func seedKeto(cmd *cobra.Command) error {
	a, ok := appFromContext(cmd.Context())
	if !ok {
		return errAppNotBootstrapped
	}
	if a.Keto == nil {
		return nil
	}
	if a.Env.KetoSeedAdminIdentityID == "" {
		return errKetoSeedAdminIdentityIDRequired
	}

	ctx := cmd.Context()
	adminIdentityID := a.Env.KetoSeedAdminIdentityID

	if err := a.Keto.CreateRelationTuple(ctx, "Role", "admin", "members", &adminIdentityID, nil); err != nil {
		return fmt.Errorf("cmd: seed keto admin role member: %w", err)
	}

	managerRole := &keto.SubjectSet{Namespace: "Role", Object: "admin", Relation: "members"}
	if err := a.Keto.CreateRelationTuple(ctx, "Identities", "admin", "managers", nil, managerRole); err != nil {
		return fmt.Errorf("cmd: seed keto identities manager: %w", err)
	}

	return nil
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
