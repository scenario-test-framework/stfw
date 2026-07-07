package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/runscenario"
)

func newRunCmd(a *app) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "run <scenario...>",
		Short: "run scenarios with the built-in runner",
		Long: "シナリオを内蔵ランナーで実行する。実行イベントは .stfw/runs/{run_id}/journal.jsonl に記録する。\n" +
			"--dry-run は execute / post_execute をスキップする (setup / teardown と計画列挙は行う)。",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runscenario.Run(a.log, cmd.OutOrStdout(), cmd.ErrOrStderr(),
				a.projDir, a.config, Version, args, dryRun, time.Now)
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "skip execute / post_execute")
	return cmd
}
