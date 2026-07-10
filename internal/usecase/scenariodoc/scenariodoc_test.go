package scenariodoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildFixtureProject は RenderDoc / ExportSpec が対象にする小さなシナリオを組み立てる。
func buildFixtureProject(t *testing.T) string {
	t.Helper()
	projDir := t.TempDir()
	base := filepath.Join(projDir, "scenario", "sample")

	writeFixtureFile(t, filepath.Join(base, "metadata.yml"),
		"description: サンプルシナリオ\n\nrequirement_specifications:\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "metadata.yml"),
		"description: Day1\n\nrequirement_specifications:\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_10_pre_scripts", "metadata.yml"),
		"description: setup\n\nrequirement_specifications:\n  - SPEC-1\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_10_pre_scripts", "config", "config.yml"),
		"stfw:\n  process:\n    scripts: {}\n")

	return projDir
}

func TestRenderDoc(t *testing.T) {
	projDir := buildFixtureProject(t)

	doc, err := RenderDoc(projDir, "sample")
	if err != nil {
		t.Fatalf("RenderDoc: %v", err)
	}
	if !strings.HasPrefix(doc, "# シナリオ: sample\n") {
		t.Errorf("RenderDoc output does not start with expected heading:\n%s", doc)
	}
	if !strings.Contains(doc, "SPEC-1") {
		t.Errorf("RenderDoc output missing requirement spec:\n%s", doc)
	}
	if !strings.Contains(doc, "_10_pre_scripts") {
		t.Errorf("RenderDoc output missing process dir:\n%s", doc)
	}
}

func TestRenderDocScenarioNotFound(t *testing.T) {
	projDir := buildFixtureProject(t)
	if _, err := RenderDoc(projDir, "nosuch"); err == nil {
		t.Fatal("expected error for non-existent scenario")
	}
}

func TestExportSpec(t *testing.T) {
	projDir := buildFixtureProject(t)

	spec, err := ExportSpec(projDir, "sample")
	if err != nil {
		t.Fatalf("ExportSpec: %v", err)
	}
	if spec.Scenario != "sample" {
		t.Errorf("Scenario = %q, want %q", spec.Scenario, "sample")
	}
	if len(spec.Bizdates) != 1 || len(spec.Bizdates[0].Processes) != 1 {
		t.Fatalf("unexpected spec shape: %#v", spec)
	}
	if spec.Bizdates[0].Processes[0].Type != "scripts" {
		t.Errorf("process type = %q, want %q", spec.Bizdates[0].Processes[0].Type, "scripts")
	}
}

func TestExportSpecScenarioNotFound(t *testing.T) {
	projDir := buildFixtureProject(t)
	if _, err := ExportSpec(projDir, "nosuch"); err == nil {
		t.Fatal("expected error for non-existent scenario")
	}
}

// ディレクトリ名規約違反 (process dir が `_{seq}_{group}_{type}` として parse できない) を
// 含むシナリオは、seq/group/type が空のまま成功扱いにせず RenderDoc/ExportSpec とも
// 失敗させる (F2)。プラグイン未インストールでの失敗ではないことを区別するため、
// メッセージにディレクトリパスを含める。
func buildFixtureProjectWithBadProcessDir(t *testing.T) string {
	t.Helper()
	projDir := t.TempDir()
	base := filepath.Join(projDir, "scenario", "broken")

	writeFixtureFile(t, filepath.Join(base, "metadata.yml"), "description: broken\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "metadata.yml"), "description: Day1\n")
	// `_{seq}_{group}_{type}` はフィールド 3 つ必須。フィールド不足で parse エラーになる。
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_xx_bad_scripts", "metadata.yml"), "description: bad\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_xx_bad_scripts", "config", "config.yml"),
		"stfw:\n  process:\n    scripts: {}\n")

	return projDir
}

func TestRenderDocStructureViolation(t *testing.T) {
	projDir := buildFixtureProjectWithBadProcessDir(t)
	if _, err := RenderDoc(projDir, "broken"); err == nil {
		t.Fatal("expected error for scenario with directory naming violation")
	} else if !strings.Contains(err.Error(), "_xx_bad_scripts") {
		t.Errorf("error = %v, want message to reference the offending dir", err)
	}
}

func TestExportSpecStructureViolation(t *testing.T) {
	projDir := buildFixtureProjectWithBadProcessDir(t)
	if _, err := ExportSpec(projDir, "broken"); err == nil {
		t.Fatal("expected error for scenario with directory naming violation")
	} else if !strings.Contains(err.Error(), "_xx_bad_scripts") {
		t.Errorf("error = %v, want message to reference the offending dir", err)
	}
}

// bizdate dir 側の命名違反 (`_{seq}_{bizdate}` として parse できない) でも同様に失敗する。
func TestRenderDocStructureViolationBizdate(t *testing.T) {
	projDir := t.TempDir()
	base := filepath.Join(projDir, "scenario", "broken2")
	writeFixtureFile(t, filepath.Join(base, "metadata.yml"), "description: broken\n")
	// seq が数字でない → NewSeq が失敗し bizdate dir の parseErr になる。
	writeFixtureFile(t, filepath.Join(base, "_1x_99990101", "metadata.yml"), "description: bad\n")

	if _, err := RenderDoc(projDir, "broken2"); err == nil {
		t.Fatal("expected error for scenario with bizdate directory naming violation")
	}
}

// 未インストールのプラグイン type は doc/spec を妨げない (validate の責務であって、
// doc/spec の責務ではない)。config.yml が無くても RenderDoc/ExportSpec は成功する。
func TestRenderDocDoesNotRequirePluginInstalled(t *testing.T) {
	projDir := t.TempDir()
	base := filepath.Join(projDir, "scenario", "uninstalled")
	writeFixtureFile(t, filepath.Join(base, "metadata.yml"), "description: x\n")
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "metadata.yml"), "description: Day1\n")
	// config/config.yml すら無い、未インストール想定のプラグイン type。
	writeFixtureFile(t, filepath.Join(base, "_10_20240101", "_10_pre_customPlugin", "metadata.yml"), "description: x\n")

	if _, err := RenderDoc(projDir, "uninstalled"); err != nil {
		t.Errorf("RenderDoc should not require plugin resolvability: %v", err)
	}
	if _, err := ExportSpec(projDir, "uninstalled"); err != nil {
		t.Errorf("ExportSpec should not require plugin resolvability: %v", err)
	}
}
