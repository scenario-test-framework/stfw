package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

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
			out := cmd.OutOrStdout()
			if err := status.Show(out, a.projDir, runID, isTerminal(out)); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

// isTerminal は出力先が端末 (TTY) かを返す (パイプ・リダイレクト時は色付けしない)。
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
