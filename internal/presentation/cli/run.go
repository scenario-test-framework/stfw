package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/runscenario"
	"github.com/scenario-test-framework/stfw/internal/usecase/secret"
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
			// プラグインが `stfw secret show` で取得したパスワードを万一出力へ
			// 漏らしてもマスクされるよう、実行前に全シークレットを Masker へ登録する。
			if err := secret.RegisterAll(a.log, a.projDir, a.masker.Register); err != nil {
				a.log.Warn("failed to register secrets for masking", "err", err.Error())
			}
			// プラグイン stdout/stderr を Masker 経由にして、登録済みシークレットを
			// 出力から除去する (ロガーと同一のシークレットレジストリを共有)。
			// 行バッファ方式のため、実行後に Flush して未改行の残りを出力する。
			out := a.masker.Wrap(cmd.OutOrStdout())
			errOut := a.masker.Wrap(cmd.ErrOrStderr())
			defer func() { _ = out.Flush() }()
			defer func() { _ = errOut.Flush() }()
			err := runscenario.Run(a.log, out, errOut,
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
