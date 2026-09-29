package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DanielKirkwood/go-api/internal/tasks"
)

var errUnknownTask = errors.New("cmd: unknown task")

// NewTaskCmd builds the `task` command group: listing and running
// self-registered ad-hoc jobs from internal/tasks.
func NewTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "List and run ad-hoc operational tasks",
	}

	taskCmd.AddCommand(newTaskListCmd())
	taskCmd.AddCommand(newTaskExecCmd())

	return taskCmd
}

func newTaskListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every registered task",
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := newTabWriter(cmd)

			fmt.Fprintln(w, "NAME\tREQUIRED ARGS")
			for _, t := range tasks.All() {
				fmt.Fprintf(w, "%s\t%s\n", t.Name, t.RequiredArgs)
			}

			return w.Flush()
		},
	}
}

func newTaskExecCmd() *cobra.Command {
	var args map[string]string

	cmd := &cobra.Command{
		Use:   "exec <name>",
		Short: "Run a registered task by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			name := cmdArgs[0]

			t, ok := tasks.Get(name)
			if !ok {
				return fmt.Errorf("%w: %s", errUnknownTask, name)
			}
			if err := t.CheckRequiredArgs(args); err != nil {
				return err
			}

			a, ok := appFromContext(cmd.Context())
			if !ok {
				return errAppNotBootstrapped
			}

			// a.Keto is a typed *ketoclient.Client, nil when the keto
			// feature is disabled. Only assign it into the Deps.Keto
			// interface field when it's actually non-nil — wrapping a nil
			// pointer in a non-nil interface would make deps.Keto == nil
			// false inside the task, defeating every task's disabled-feature
			// guard.
			var deps tasks.Deps
			if a.Keto != nil {
				deps.Keto = a.Keto
			}

			return t.Run(cmd.Context(), deps, args)
		},
	}

	cmd.Flags().StringToStringVar(&args, "arg", nil, "task argument as key=value (repeatable)")

	return cmd
}
