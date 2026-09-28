package cmd

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"

	"github.com/DanielKirkwood/go-api/internal/api"
)

func newInfoPublicRoutesCmd() *cobra.Command {
	return newInfoRoutesCmd("public-api-routes", "Print the public router's registered routes", api.NewPublicRouter)
}

func newInfoProtectedRoutesCmd() *cobra.Command {
	return newInfoRoutesCmd(
		"protected-api-routes", "Print the protected router's registered routes", api.NewProtectedRouter,
	)
}

func newInfoHiddenRoutesCmd() *cobra.Command {
	return newInfoRoutesCmd("hidden-api-routes", "Print the hidden router's registered routes", api.NewHiddenRouter)
}

func newInfoRoutesCmd(use, short string, build func(api.RouterDeps) *chi.Mux) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, ok := appFromContext(cmd.Context())
			if !ok {
				return errAppNotBootstrapped
			}

			router := build(api.RouterDeps{Logger: a.Logger, TracerProvider: a.Otel.TracerProvider})

			return printRoutes(cmd, router)
		},
	}
}

func printRoutes(cmd *cobra.Command, router chi.Router) error {
	w := newTabWriter(cmd)

	fmt.Fprintln(w, "METHOD\tROUTE")

	walkErr := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		fmt.Fprintf(w, "%s\t%s\n", method, route)
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	return w.Flush()
}
