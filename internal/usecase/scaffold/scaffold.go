// Package scaffold は stfw new (scenario / bizdate / process) のビジネスフローを制御する。
package scaffold

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Scenario はシナリオ scaffold を生成する。
// 生成先は {projDir}/scenario/{name} に固定 (v0.2 はカレントがシナリオルートで
// あることを要求していたが、プロジェクトにシナリオルートは 1 つなので固定できる)。
func Scenario(log *slog.Logger, out io.Writer, projDir, nameStr string) error {
	name, err := scenario.NewScenarioName(nameStr)
	if err != nil {
		return err
	}

	root := filepath.Join(projDir, scenario.RootDirName)
	if !repository.ProjectConfigExists(projDir, project.ConfigFileName) || !repository.DirExists(root) {
		return fmt.Errorf("%s is not scenario-root-dir", root)
	}

	created, err := repository.CreateNodeScaffold(root, name.String())
	if err != nil {
		return fmt.Errorf("scenario scaffold: %w", err)
	}

	printCreated(out, projDir, created)
	log.Info("scenario initialized", "name", name.String())
	return nil
}

// Bizdate は業務日付 scaffold を生成する。cwd はシナリオディレクトリであること。
func Bizdate(log *slog.Logger, out io.Writer, projDir, cwd, seqStr, bizdateStr string) error {
	seq, err := scenario.NewSeq(seqStr)
	if err != nil {
		return err
	}
	bizdate, err := scenario.NewBizdate(bizdateStr)
	if err != nil {
		return err
	}

	if !isHierarchyDir(projDir, cwd, scenario.IsScenarioDir) {
		return fmt.Errorf("%s is not scenario-dir", cwd)
	}

	dirName := scenario.BizdateDirName(seq, bizdate)
	created, err := repository.CreateNodeScaffold(cwd, dirName)
	if err != nil {
		return fmt.Errorf("bizdate scaffold: %w", err)
	}

	printCreated(out, projDir, created)
	log.Info("bizdate initialized", "dir", dirName)
	return nil
}

// Process はプロセス scaffold を生成する。cwd は業務日付ディレクトリであること。
// プラグインの template/ を展開する (既存の同名ディレクトリは作り直し。v0.2 互換)。
func Process(log *slog.Logger, out io.Writer, projDir, cwd, seqStr, groupStr, processType string) error {
	seq, err := scenario.NewSeq(seqStr)
	if err != nil {
		return err
	}
	group, err := scenario.NewGroup(groupStr)
	if err != nil {
		return err
	}
	if err := scenario.ValidateProcessType(processType); err != nil {
		return err
	}

	if !isHierarchyDir(projDir, cwd, scenario.IsBizdateDir) {
		return fmt.Errorf("%s is not bizdate-dir", cwd)
	}

	// プラグイン解決 (プロジェクト plugins/ → 同梱の順)
	loc, err := repository.ResolveProcessPlugin(projDir, processType)
	if err != nil {
		return err
	}

	dirName := scenario.ProcessDirName(seq, group, processType)
	created, err := repository.CreateProcessScaffold(loc, cwd, dirName)
	if err != nil {
		return fmt.Errorf("process scaffold: %w", err)
	}

	printCreated(out, projDir, created)
	log.Info("process initialized", "dir", dirName, "type", processType)
	return nil
}

// isHierarchyDir は cwd がプロジェクト内の期待する階層かを判定する。
// v0.2 の is_*-dir 判定 (深さ + stfw.yml の存在) に対応する。
func isHierarchyDir(projDir, cwd string, isLevel func(rel string) bool) bool {
	if !repository.ProjectConfigExists(projDir, project.ConfigFileName) {
		return false
	}
	rel, err := filepath.Rel(projDir, cwd)
	if err != nil {
		return false
	}
	return isLevel(filepath.ToSlash(rel))
}

// printCreated は作成ファイルをプロジェクトルートからの相対パスで出力する。
func printCreated(out io.Writer, projDir string, created []string) {
	for _, p := range created {
		rel, err := filepath.Rel(projDir, p)
		if err != nil {
			rel = p
		}
		fmt.Fprintln(out, filepath.ToSlash(rel))
	}
}
