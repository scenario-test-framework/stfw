// Package validate は stfw validate のビジネスフローを制御する。
// ディレクトリ規約・プラグイン解決可否・config.yml 存在を静的に検証する
// (v0.2 の dig 生成 + 検証から、dig 生成を廃止して静的検証に昇格したもの)。
package validate

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Run はシナリオ構造を検証する。names が空の場合は全シナリオを対象とする。
// エラーレベルの違反があれば error を返す (exit 6)。警告のみは正常終了 (exit 0)。
func Run(log *slog.Logger, out io.Writer, projDir string, names []string) error {
	tree, err := repository.LoadScenarioTree(projDir, names)
	if err != nil {
		return err
	}

	installed, err := repository.ListProcessPlugins(projDir)
	if err != nil {
		return err
	}

	violations := tree.Validate(installed)
	for _, v := range violations {
		fmt.Fprintln(out, v.String())
	}

	// プラグインのランタイム依存 (plugin.yml requires) の存在チェック。
	// validate は実行環境と異なるマシンで走り得るため警告に留める
	// (実行前ゲートは run 側で error として扱う)。
	missing, err := repository.CheckPluginRequires(projDir, tree.ProcessTypes())
	if err != nil {
		return err
	}
	for _, m := range missing {
		v := scenario.Violation{Path: m.ProcessType, Level: scenario.ViolationWarn,
			Message: fmt.Sprintf("required command not found: %s", m.Command)}
		fmt.Fprintln(out, v.String())
	}

	// 接続情報 (host / password 等) の直書き禁止。設定は環境非依存の静的性質
	// のため error 扱い (グループ名参照の徹底)。
	forbidden, err := repository.CheckForbiddenConnConfig(projDir, tree.ScenarioViews())
	if err != nil {
		return err
	}
	for _, f := range forbidden {
		v := scenario.Violation{Path: f.ProcessPath, Level: scenario.ViolationError,
			Message: fmt.Sprintf("config で接続情報を直書きしています (%s)。inventory グループ名参照 + secret 参照を使ってください", f.Key)}
		fmt.Fprintln(out, v.String())
	}

	errCount, warns := violations.Count()
	errCount += len(forbidden)
	warns += len(missing)
	if errCount > 0 {
		return fmt.Errorf("validation failed: %d error(s), %d warning(s)", errCount, warns)
	}
	log.Info("validation passed", "scenarios", len(tree.Scenarios()), "warnings", warns)
	return nil
}
