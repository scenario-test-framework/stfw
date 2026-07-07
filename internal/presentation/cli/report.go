package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/report"
)

func newReportCmd(a *app) *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "report [run_id]",
		Short: "generate html report from the journal",
		Long: "ジャーナルから HTML レポート (index.html + runs/{run_id}.html) を再生成する。\n" +
			"run_id 省略時は最新の run を対象とする。--out 省略時は .stfw/reports へ出力する。",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := ""
			if len(args) == 1 {
				runID = args[0]
			}
			if err := report.Generate(cmd.OutOrStdout(), a.projDir, runID, outDir); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outDir, "out", "o", "", "output directory (default: .stfw/reports)")
	return cmd
}
