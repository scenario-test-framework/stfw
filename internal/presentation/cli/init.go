package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/initialize"
)

func newInitCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialize stfw project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initialize.Run(a.log, cmd.OutOrStdout(), a.projDir); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}
