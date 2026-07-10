package gateway

import (
	"fmt"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// RenderScenarioDoc は scenario.DocData を Markdown へレンダリングする
// (`stfw scenario reverse` の tree → doc 投影)。データの組み立て (metadata.yml /
// config.yml の読取・グループ抽出・要求トレーサビリティ集約) は repository 層で
// 完了済みで、ここでは文字列の組み立てのみを担う (htmlwriter.go のテンプレート
// レンダラに相当する役割)。golden テストで完全一致比較できるよう、テーブル・
// コードブロックの改行制御を text/template の空白トリムに頼らず直接組み立てる。
func RenderScenarioDoc(doc scenario.DocData) (string, error) {
	var b strings.Builder

	fmt.Fprintf(&b, "# シナリオ: %s\n\n", doc.Name)
	if doc.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", doc.Description)
	}

	if len(doc.Traceability) > 0 {
		b.WriteString("## 要求トレーサビリティ\n\n")
		b.WriteString("| 要求仕様 | 検証する process |\n|---|---|\n")
		for _, row := range doc.Traceability {
			fmt.Fprintf(&b, "| %s | %s |\n", row.RequirementSpecification, row.ProcessPaths)
		}
		b.WriteString("\n")
	}

	for _, bz := range doc.Bizdates {
		fmt.Fprintf(&b, "## %s\n\n", bz.Title)

		b.WriteString("| # | process | グループ | プラグイン | 説明 |\n|---|---|---|---|---|\n")
		for _, p := range bz.Processes {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", p.SeqLabel, p.DirName, p.Group, p.Type, p.Description)
		}
		b.WriteString("\n")

		for _, p := range bz.Processes {
			writeProcessSection(&b, p)
		}
	}

	return strings.TrimRight(b.String(), "\n") + "\n", nil
}

func writeProcessSection(b *strings.Builder, p scenario.DocProcess) {
	fmt.Fprintf(b, "### %s\n\n", p.DirName)
	fmt.Fprintf(b, "- グループ: %s\n", p.Group)

	reqs := "-"
	if len(p.RequirementSpecifications) > 0 {
		reqs = strings.Join(p.RequirementSpecifications, ", ")
	}
	fmt.Fprintf(b, "- 要求仕様: %s\n", reqs)

	if p.ConfigYAML != "" {
		b.WriteString("- 設定:\n\n")
		b.WriteString("    ```yaml\n")
		for _, line := range strings.Split(p.ConfigYAML, "\n") {
			b.WriteString("    " + line + "\n")
		}
		b.WriteString("    ```\n")
	}
	b.WriteString("\n")
}
