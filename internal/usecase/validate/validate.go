// Package validate は stfw validate のビジネスフローを制御する。
// ディレクトリ規約・プラグイン解決可否・config.yml 存在を静的に検証する
// (v0.2 の dig 生成 + 検証から、dig 生成を廃止して静的検証に昇格したもの)。
package validate

import (
	"fmt"
	"io"
	"log/slog"

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

	errors, warns := violations.Count()
	if errors > 0 {
		return fmt.Errorf("validation failed: %d error(s), %d warning(s)", errors, warns)
	}
	log.Info("validation passed", "scenarios", len(tree.Scenarios()), "warnings", warns)
	return nil
}
