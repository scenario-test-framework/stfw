// Package scenariodoc は stfw scenario reverse (tree → spec + doc) のビジネスフローを
// 制御する (tree が真実の源で spec/doc はその投影という方式に基づく。往復の入口
// spec → tree は usecase/scaffold が担う)。
package scenariodoc

import (
	"fmt"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// RenderDoc はシナリオ 1 つを Markdown ドキュメントへ投影する (tree → doc)。
func RenderDoc(projDir, name string) (string, error) {
	view, err := loadScenarioView(projDir, name)
	if err != nil {
		return "", err
	}
	doc, err := repository.BuildDocFromTree(projDir, view)
	if err != nil {
		return "", err
	}
	return gateway.RenderScenarioDoc(doc)
}

// ExportSpec はシナリオ 1 つを spec DTO へ投影する (tree → spec、往復の出口)。
func ExportSpec(projDir, name string) (repository.ScenarioSpec, error) {
	view, err := loadScenarioView(projDir, name)
	if err != nil {
		return repository.ScenarioSpec{}, err
	}
	return repository.BuildSpecFromTree(projDir, view)
}

// Reverse はシナリオ 1 つを spec YAML と doc Markdown の両方へ投影する
// (tree → spec + doc のリバース生成)。`stfw scenario reverse` の中身で、
// spec (往復の媒体) と doc (レビュー資料) を常にセットで返す。
func Reverse(projDir, name string) (specYAML string, docMD string, err error) {
	docMD, err = RenderDoc(projDir, name)
	if err != nil {
		return "", "", err
	}
	spec, err := ExportSpec(projDir, name)
	if err != nil {
		return "", "", err
	}
	raw, err := repository.MarshalSpec(spec)
	if err != nil {
		return "", "", err
	}
	return string(raw), docMD, nil
}

// loadScenarioView は projDir 配下のシナリオ name を走査して ScenarioView を返す。
// ディレクトリ名規約違反 (seq/bizdate/group/type が parse できない) がある場合は失敗する。
// doc/spec はプラグイン未インストールでも投影できるべきなので、プラグイン解決可否・
// config.yml 存在は見ない (それらは `stfw validate` の責務)。
func loadScenarioView(projDir, name string) (scenario.ScenarioView, error) {
	tree, err := repository.LoadScenarioTree(projDir, []string{name})
	if err != nil {
		return scenario.ScenarioView{}, err
	}
	if vs := tree.StructureViolations(); len(vs) > 0 {
		msgs := make([]string, 0, len(vs))
		for _, v := range vs {
			msgs = append(msgs, v.String())
		}
		return scenario.ScenarioView{}, fmt.Errorf("scenario: %s has directory naming violations: %s", name, strings.Join(msgs, "; "))
	}
	view, ok := tree.ScenarioView(name)
	if !ok {
		return scenario.ScenarioView{}, fmt.Errorf("scenario: %s is not exist", name)
	}
	return view, nil
}
