package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/status"
)

func newStatusCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "status [run_id]",
		Short: "show run status from the journal",
		Long:  "ジャーナルをリプレイして実行ツリーとステータスを表示する。run_id 省略時は最新の run を対象とする。",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := ""
			if len(args) == 1 {
				runID = args[0]
			}
			if err := status.Show(cmd.OutOrStdout(), a.projDir, runID); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}
