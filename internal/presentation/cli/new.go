package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/scaffold"
)

func newNewCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "generate scaffold (scenario / bizdate / process)",
	}
	cmd.AddCommand(
		newNewScenarioCmd(a),
		newNewBizdateCmd(a),
		newNewProcessCmd(a),
	)
	return cmd
}

func newNewScenarioCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "scenario <name>",
		Short: "generate scenario scaffold to scenario/",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := scaffold.Scenario(a.log, cmd.OutOrStdout(), a.projDir, args[0]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newNewBizdateCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "bizdate <seq> <bizdate>",
		Short: "generate bizdate scaffold to current scenario directory (bizdate format: YYYYMMDD)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := scaffold.Bizdate(a.log, cmd.OutOrStdout(), a.projDir, cwd, args[0], args[1]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newNewProcessCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "process <seq> <group> <type>",
		Short: "generate process scaffold to current bizdate directory",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := scaffold.Process(a.log, cmd.OutOrStdout(), a.projDir, cwd, args[0], args[1], args[2]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}
