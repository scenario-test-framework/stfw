package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/runscenario"
	"github.com/scenario-test-framework/stfw/internal/usecase/secret"
)

func newRunCmd(a *app) *cobra.Command {
	var dryRun bool
	var from, only, resume string
	cmd := &cobra.Command{
		Use:   "run <scenario...>",
		Short: "run scenarios with the built-in runner",
		Long: "シナリオを内蔵ランナーで実行する。実行イベントは .stfw/runs/{run_id}/journal.jsonl に記録する。\n" +
			"--dry-run は execute / post_execute をスキップする (setup / teardown と計画列挙は行う)。\n" +
			"--from / --only は部分実行 (排他・シナリオ 1 つ指定時のみ)。パスはシナリオ相対の\n" +
			"{bizdate_dir}[/{process_dir}]。--from は指定ノードから最後まで、--only は指定サブツリーのみ実行する。\n" +
			"--resume は前 run のワークスペース (エビデンス等の生成物) を引き継いで実行する。\n" +
			"値省略時は最新 run、指定時は --resume=<run_id> 形式 (--from / --only との併用が主用途)。",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// `--from=` のような空値の明示指定は「フィルタ未指定 = 全体実行」に
			// 化けさせず fail-fast する (空文字は契約上有効なパスではない。§3.4)
			for _, f := range []struct{ name, value string }{{"from", from}, {"only", only}} {
				if cmd.Flags().Changed(f.name) && f.value == "" {
					err := fmt.Errorf("--%s requires a non-empty {bizdate_dir}[/{process_dir}] path", f.name)
					a.log.Error(err.Error())
					return &exitError{code: run.ExitError, err: err}
				}
			}
			// `--resume=` (空値) も「最新 run」に化けさせず fail-fast する (§5.8)
			if cmd.Flags().Changed("resume") && resume == "" {
				err := fmt.Errorf("--resume requires a run_id (or omit the value for the latest run)")
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
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
			opts := runscenario.Options{DryRun: dryRun, From: from, Only: only, Resume: resume}
			err := runscenario.Run(a.log, out, errOut,
				a.projDir, a.config, Version, args, opts, time.Now)
			if err != nil {
				// Warn 完走 (Error なし) は exit 3 で「差分あり」を CI へ伝える (SPEC-023-03)
				var st *runscenario.StatusError
				if errors.As(err, &st) && st.Status == run.NodeWarn {
					a.log.Warn(err.Error())
					return &exitError{code: run.ExitWarn, err: err}
				}
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "skip execute / post_execute")
	cmd.Flags().StringVar(&from, "from", "", "resume from {bizdate_dir}[/{process_dir}] (single scenario only)")
	cmd.Flags().StringVar(&only, "only", "", "run only {bizdate_dir}[/{process_dir}] subtree (single scenario only)")
	cmd.Flags().StringVar(&resume, "resume", "", "carry over a previous run's workspace (no value = latest run; use --resume=<run_id> to specify)")
	// `--resume` (値なし) を「最新 run からの引き継ぎ」にする (§5.8)
	cmd.Flags().Lookup("resume").NoOptDefVal = "latest"
	cmd.MarkFlagsMutuallyExclusive("from", "only")
	return cmd
}
