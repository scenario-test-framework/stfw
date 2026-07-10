package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
	"github.com/scenario-test-framework/stfw/internal/usecase/scaffold"
	"github.com/scenario-test-framework/stfw/internal/usecase/scenariodoc"
)

// defaultReverseDir はリバース生成 (spec + doc) の既定出力ディレクトリ名。
const defaultReverseDir = "docs"

// newScenarioCmd は `stfw scenario` コマンドグループ (reverse / scaffold) を定義する。
// `stfw new scenario` (対話・単一ノード生成) とは別物で、こちらは
// tree ⇄ spec の往復 (reverse で tree → spec + doc、scaffold で spec → tree) を担う
// (tree が真実の源、spec が往復の媒体、doc は読み取り専用の投影)。
func newScenarioCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scenario",
		Short: "reverse a scenario tree to spec+doc, or scaffold a tree from spec (tree <-> spec)",
	}
	cmd.AddCommand(
		newScenarioReverseCmd(a),
		newScenarioScaffoldCmd(a),
	)
	return cmd
}

func newScenarioReverseCmd(a *app) *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "reverse <name>",
		Short: "reverse-generate spec yaml + markdown doc from a scenario tree (tree -> spec + doc)",
		Long: "既存シナリオ (tree) から spec (<name>.yml) と doc (<name>.md) をまとめて生成する。\n" +
			"spec は往復の媒体 (scaffold の入力)、doc は要求トレーサビリティ表つきのレビュー資料。\n" +
			"出力先は -o で指定するディレクトリ (既定: docs/)。ファイル名はシナリオ名に固定。",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			specYAML, docMD, err := scenariodoc.Reverse(a.projDir, name)
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}

			dir := outDir
			if dir == "" {
				dir = filepath.Join(a.projDir, defaultReverseDir)
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}

			specPath := filepath.Join(dir, name+".yml")
			docPath := filepath.Join(dir, name+".md")
			for path, content := range map[string]string{specPath: specYAML, docPath: docMD} {
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					a.log.Error(err.Error())
					return &exitError{code: run.ExitError, err: err}
				}
			}
			// 出力順を安定させる (map の走査順に依存しない)
			printScenarioPath(cmd, a.projDir, specPath)
			printScenarioPath(cmd, a.projDir, docPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "output directory (default: docs/)")
	return cmd
}

// printScenarioPath は書き出したファイルをプロジェクトルート相対 (可能なら) で出力する。
func printScenarioPath(cmd *cobra.Command, projDir, path string) {
	rel, err := filepath.Rel(projDir, path)
	if err != nil {
		rel = path
	}
	fmt.Fprintln(cmd.OutOrStdout(), filepath.ToSlash(rel))
}

func newScenarioScaffoldCmd(a *app) *cobra.Command {
	var sync bool
	cmd := &cobra.Command{
		Use:   "scaffold <spec.yml>",
		Short: "generate scenario scaffold from spec yaml (spec -> tree, roundtrip entry point)",
		Long: "spec (structured yaml) からシナリオのディレクトリ骨格 (metadata.yml + config/config.yml) を生成する。\n" +
			"data/scripts/expect 等の葉は生成しない。既存シナリオは既定でエラー。\n" +
			"--sync は spec との差分同期: spec に無い bizdate/process ディレクトリを実装済みの葉ごと削除する (破壊的)。",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(args[0])
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			spec, err := repository.UnmarshalSpec(raw)
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			if err := scaffold.ScaffoldFromSpec(a.log, cmd.OutOrStdout(), a.projDir, spec, sync); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&sync, "sync", false, "sync an existing scenario with the spec: add missing, overwrite skeleton, and delete bizdate/process directories not in the spec (destructive; removes implemented leaves)")
	return cmd
}
