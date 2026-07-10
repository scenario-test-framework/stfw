package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
	"github.com/scenario-test-framework/stfw/internal/usecase/scaffold"
	"github.com/scenario-test-framework/stfw/internal/usecase/scenariodoc"
)

// newScenarioCmd は `stfw scenario` コマンドグループ (doc / spec / scaffold) を定義する。
// `stfw new scenario` (対話・単一ノード生成) とは別物で、こちらは
// tree ⇄ doc・tree ⇄ spec の投影・往復を担う (tree が真実の源、spec が往復の媒体、
// doc は読み取り専用の投影)。
func newScenarioCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scenario",
		Short: "project scenario as doc/spec, or scaffold from spec (tree <-> doc/spec)",
	}
	cmd.AddCommand(
		newScenarioDocCmd(a),
		newScenarioSpecCmd(a),
		newScenarioScaffoldCmd(a),
	)
	return cmd
}

func newScenarioDocCmd(a *app) *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "doc <name>",
		Short: "render scenario as markdown doc (tree -> doc)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := scenariodoc.RenderDoc(a.projDir, args[0])
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			if err := writeScenarioOutput(cmd, outFile, doc); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output file (default: stdout)")
	return cmd
}

func newScenarioSpecCmd(a *app) *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "spec <name>",
		Short: "export scenario as spec yaml (tree -> spec, roundtrip exit point)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := scenariodoc.ExportSpec(a.projDir, args[0])
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			raw, err := repository.MarshalSpec(spec)
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			if err := writeScenarioOutput(cmd, outFile, string(raw)); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&outFile, "out", "o", "", "output file (default: stdout)")
	return cmd
}

func newScenarioScaffoldCmd(a *app) *cobra.Command {
	var force bool
	var prune bool
	cmd := &cobra.Command{
		Use:   "scaffold <spec.yml>",
		Short: "generate scenario scaffold from spec yaml (spec -> tree, roundtrip entry point)",
		Long: "spec (structured yaml) からシナリオのディレクトリ骨格 (metadata.yml + config/config.yml) を生成する。\n" +
			"data/scripts/expect 等の葉は生成しない。既存シナリオは既定でエラー (--force で再生成)。\n" +
			"--prune は spec との差分同期: spec に無い bizdate/process ディレクトリを実装済みの葉ごと削除する (破壊的・--force を含意)。",
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
			if err := scaffold.ScaffoldFromSpec(a.log, cmd.OutOrStdout(), a.projDir, spec, force, prune); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "regenerate even if the scenario directory already exists")
	cmd.Flags().BoolVar(&prune, "prune", false, "sync with spec: delete bizdate/process directories not in the spec (destructive; removes implemented leaves; implies --force)")
	return cmd
}

// writeScenarioOutput は content を outFile へ書き出す。outFile が空なら stdout へ出力する。
func writeScenarioOutput(cmd *cobra.Command, outFile, content string) error {
	if outFile == "" {
		fmt.Fprint(cmd.OutOrStdout(), content)
		return nil
	}
	if err := os.WriteFile(outFile, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), outFile)
	return nil
}
