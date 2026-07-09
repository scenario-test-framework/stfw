package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/validate"
)

func newValidateCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [scenario...]",
		Short: "validate scenario directory structure",
		Long: "ディレクトリ規約・プラグイン解決可否・config.yml の存在を静的に検証する。\n" +
			"エラーは exit 6、警告のみは exit 0 で終了する。",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.Run(a.log, cmd.OutOrStdout(), a.projDir, args); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}
