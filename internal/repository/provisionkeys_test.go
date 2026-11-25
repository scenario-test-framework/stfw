package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// writeFile はテスト用のファイルを親ディレクトリごと作る。
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newProvisionPluginProject は provision_keys を宣言したプロジェクトプラグイン (tool) を持つ
// プロジェクトを作り、プロジェクトディレクトリとプラグインの場所を返す。
func newProvisionPluginProject(t *testing.T) (string, PluginLocation) {
	t.Helper()
	projDir := t.TempDir()
	pluginDir := filepath.Join(projDir, "plugins", "process", "tool")
	writeFile(t, filepath.Join(pluginDir, "plugin.yml"), "requires: []\nprovision_keys:\n  - tool_version\n  - tool_arches\n")
	writeFile(t, filepath.Join(pluginDir, "config.yml"),
		"# tool の既定設定\nstfw:\n  process:\n    tool:\n      tool_version: v1.0.0\n      tool_arches:\n        - linux_amd64\n      mode: fast # 実行方式\n")
	loc, err := ResolveProcessPlugin(projDir, "tool")
	if err != nil {
		t.Fatal(err)
	}
	return projDir, loc
}

// TestProcessConfigEnvProvisionKeys は、provision_keys のキーがプロセス設定に書かれていても
// プロジェクト共通までの値で決まり、それ以外のキーは従来どおりプロセス設定が勝つことを固定する。
func TestProcessConfigEnvProvisionKeys(t *testing.T) {
	t.Run("ProcessConfigEnv_プロセス設定に版がある場合_プロジェクト共通までの値になること", func(t *testing.T) {
		// Arrange
		projDir, loc := newProvisionPluginProject(t)
		writeFile(t, filepath.Join(projDir, "config", "plugins", "process", "tool", "config.yml"),
			"stfw:\n  process:\n    tool:\n      tool_version: v2.0.0\n")
		processDir := filepath.Join(projDir, "scenario", "s", "_10_99990101", "_10_g_tool")
		writeFile(t, filepath.Join(processDir, "config", "config.yml"),
			"stfw:\n  process:\n    tool:\n      tool_version: v0.9.0\n      tool_arches:\n        - darwin_arm64\n        - linux_arm64\n      mode: slow\n")

		// Act
		flat, err := ProcessConfigEnv(projDir, loc, "tool", processDir)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if got := flat["stfw_process_tool_tool_version"]; got != "v2.0.0" {
			t.Errorf("tool_version = %q, want v2.0.0 (プロジェクト共通)", got)
		}
		if got := flat["stfw_process_tool_tool_arches_0"]; got != "linux_amd64" {
			t.Errorf("tool_arches_0 = %q, want linux_amd64 (プラグイン既定)", got)
		}
		if _, ok := flat["stfw_process_tool_tool_arches_1"]; ok {
			t.Error("tool_arches_1 はプロセス設定由来なので残らないべき")
		}
		if got := flat["stfw_process_tool_mode"]; got != "slow" {
			t.Errorf("mode = %q, want slow (provision_keys 以外はプロセス設定が勝つ)", got)
		}
	})

	t.Run("ProcessConfigEnv_provision_keysの無いプラグインの場合_プロセス設定が勝つこと", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		pluginDir := filepath.Join(projDir, "plugins", "process", "plain")
		writeFile(t, filepath.Join(pluginDir, "config.yml"), "stfw:\n  process:\n    plain:\n      tool_version: v1.0.0\n")
		loc, err := ResolveProcessPlugin(projDir, "plain")
		if err != nil {
			t.Fatal(err)
		}
		processDir := filepath.Join(projDir, "p")
		writeFile(t, filepath.Join(processDir, "config", "config.yml"), "stfw:\n  process:\n    plain:\n      tool_version: v0.9.0\n")

		// Act
		flat, err := ProcessConfigEnv(projDir, loc, "plain", processDir)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if got := flat["stfw_process_plain_tool_version"]; got != "v0.9.0" {
			t.Errorf("tool_version = %q, want v0.9.0 (宣言が無ければ従来どおり)", got)
		}
	})
}

// TestCopyPluginDefaultConfigProvisionKeys は、雛形でプロセスの設定へ既定 config.yml を複製するとき、
// provision_keys のキーだけを除くことを固定する。
func TestCopyPluginDefaultConfigProvisionKeys(t *testing.T) {
	t.Run("CopyPluginDefaultConfig_provision_keysがある場合_そのキーを除いて複製すること", func(t *testing.T) {
		// Arrange
		_, loc := newProvisionPluginProject(t)
		dest := t.TempDir()

		// Act
		_, err := CopyPluginDefaultConfig(loc, dest)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(dest, "config", "config.yml"))
		if err != nil {
			t.Fatal(err)
		}
		got := string(raw)
		for _, k := range []string{"tool_version", "tool_arches", "linux_amd64"} {
			if strings.Contains(got, k) {
				t.Errorf("雛形に %s が残っている:\n%s", k, got)
			}
		}
		if !strings.Contains(got, "mode: fast") {
			t.Errorf("provision_keys 以外のキーは残すべき:\n%s", got)
		}
	})

	t.Run("CopyPluginDefaultConfig_組込みcollectLogの場合_版のキーを複製しないこと", func(t *testing.T) {
		// Arrange
		loc, err := ResolveProcessPlugin(t.TempDir(), "collectLog")
		if err != nil {
			t.Fatal(err)
		}
		dest := t.TempDir()

		// Act
		_, err = CopyPluginDefaultConfig(loc, dest)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(dest, "config", "config.yml"))
		if err != nil {
			t.Fatal(err)
		}
		got := string(raw)
		if strings.Contains(got, "logfilter_version:") || strings.Contains(got, "logfilter_arches:") {
			t.Errorf("組込み collectLog の雛形に版のキーが残っている:\n%s", got)
		}
		if !strings.Contains(got, "targets:") {
			t.Errorf("targets は残すべき:\n%s", got)
		}
	})
}

// TestStripProvisionKeys は既定 config.yml から provision_keys を取り除く処理の境界を固定する。
func TestStripProvisionKeys(t *testing.T) {
	t.Run("stripProvisionKeys_取り除くキーが無い場合_元のバイト列のままであること", func(t *testing.T) {
		// Arrange
		raw := []byte("# コメント\nstfw:\n  process:\n    tool:\n      mode: fast\n")

		// Act
		got, err := stripProvisionKeys(raw, "tool", []string{"tool_version"})

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(raw) {
			t.Errorf("該当キーが無ければ書式を保つべき: got %q", got)
		}
	})

	t.Run("stripProvisionKeys_stfw.process.typeが無い場合_元のバイト列のままであること", func(t *testing.T) {
		// Arrange
		raw := []byte("stfw:\n  other: 1\n")

		// Act
		got, err := stripProvisionKeys(raw, "tool", []string{"tool_version"})

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(raw) {
			t.Errorf("対象のマッピングが無ければそのまま返すべき: got %q", got)
		}
	})
}

// TestCheckProvisionKeysInProcessConfig は、プロセスの設定に書かれた provision key を検出し、
// 書かれていなければ何も返さないことを固定する。
func TestCheckProvisionKeysInProcessConfig(t *testing.T) {
	t.Run("CheckProvisionKeysInProcessConfig_プロセス設定に版がある場合_検出すること", func(t *testing.T) {
		// Arrange
		projDir, _ := newProvisionPluginProject(t)
		writeFile(t, filepath.Join(projDir, "scenario", "s", "_10_99990101", "_10_g_tool", "config", "config.yml"),
			"stfw:\n  process:\n    tool:\n      tool_arches:\n        - linux_arm64\n      mode: slow\n")
		writeFile(t, filepath.Join(projDir, "scenario", "s", "_10_99990101", "_20_g_tool", "config", "config.yml"),
			"stfw:\n  process:\n    tool:\n      mode: slow\n")
		views := []scenario.ScenarioView{{Name: "s", Bizdates: []scenario.BizdateView{{DirName: "_10_99990101",
			Processes: []scenario.ProcessView{{DirName: "_10_g_tool", ProcessType: "tool"}, {DirName: "_20_g_tool", ProcessType: "tool"}}}}}}

		// Act
		found, err := CheckProvisionKeysInProcessConfig(projDir, views)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if len(found) != 1 {
			t.Fatalf("found = %+v, want 1 件 (_10_g_tool の tool_arches)", found)
		}
		if found[0].ProcessPath != "scenario/s/_10_99990101/_10_g_tool" || found[0].Key != "tool_arches" || found[0].ProcessType != "tool" {
			t.Errorf("found[0] = %+v", found[0])
		}
	})
}

// TestProvisionKeysBoundary は provision_keys の選び方の境界 (完全一致・空のリスト・アンカー) を固定する。
func TestProvisionKeysBoundary(t *testing.T) {
	t.Run("ProcessConfigEnv_名前の先頭が同じ別のキーの場合_プロセス設定が勝つこと", func(t *testing.T) {
		// Arrange
		projDir, loc := newProvisionPluginProject(t)
		processDir := filepath.Join(projDir, "p")
		writeFile(t, filepath.Join(processDir, "config", "config.yml"),
			"stfw:\n  process:\n    tool:\n      tool_version: v0.9.0\n      tool_version_label: local\n")

		// Act
		flat, err := ProcessConfigEnv(projDir, loc, "tool", processDir)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if got := flat["stfw_process_tool_tool_version"]; got != "v1.0.0" {
			t.Errorf("tool_version = %q, want v1.0.0 (プラグイン既定)", got)
		}
		if got := flat["stfw_process_tool_tool_version_label"]; got != "local" {
			t.Errorf("tool_version_label = %q, want local (宣言していないキーはプロセス設定が勝つ)", got)
		}
	})

	t.Run("CheckProvisionKeysInProcessConfig_空のリストで書いた場合_検出すること", func(t *testing.T) {
		// Arrange
		projDir, _ := newProvisionPluginProject(t)
		writeFile(t, filepath.Join(projDir, "scenario", "s", "_10_99990101", "_10_g_tool", "config", "config.yml"),
			"stfw:\n  process:\n    tool:\n      tool_arches: []\n")
		views := []scenario.ScenarioView{{Name: "s", Bizdates: []scenario.BizdateView{{DirName: "_10_99990101",
			Processes: []scenario.ProcessView{{DirName: "_10_g_tool", ProcessType: "tool"}}}}}}

		// Act
		found, err := CheckProvisionKeysInProcessConfig(projDir, views)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if len(found) != 1 || found[0].Key != "tool_arches" {
			t.Errorf("found = %+v, want tool_arches 1 件 (空のリストでも書かれている)", found)
		}
	})

	t.Run("stripProvisionKeys_取り除くキーのアンカーを参照している場合_読み戻せるYAMLであること", func(t *testing.T) {
		// Arrange
		raw := []byte("stfw:\n  process:\n    tool:\n      tool_version: &ver v1.0.0\n      expected_version: *ver\n")

		// Act
		got, err := stripProvisionKeys(raw, "tool", []string{"tool_version"})

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		flat := map[string]string{}
		if err := flattenYAML(got, flat); err != nil {
			t.Fatalf("読み戻せない YAML を生成した: %v\n%s", err, got)
		}
		if flat["stfw_process_tool_expected_version"] != "v1.0.0" {
			t.Errorf("参照していた値が失われた: %v\n%s", flat, got)
		}
		if _, ok := flat["stfw_process_tool_tool_version"]; ok {
			t.Errorf("tool_version は取り除くべき:\n%s", got)
		}
	})
}

// TestProvisionKeysRound3 は、別の書き方で同じ env 名に展開される値と、別名の循環の扱いを固定する。
func TestProvisionKeysRound3(t *testing.T) {
	t.Run("ProcessConfigEnv_別のYAMLパスが同じenv名になる場合_プロセス設定で上書きできないこと", func(t *testing.T) {
		// Arrange
		projDir, loc := newProvisionPluginProject(t)
		processDir := filepath.Join(projDir, "p")
		writeFile(t, filepath.Join(processDir, "config", "config.yml"),
			"stfw:\n  process:\n    tool:\n      tool:\n        version: v0.9.0\n      tool_version_label: local\n")

		// Act
		flat, err := ProcessConfigEnv(projDir, loc, "tool", processDir)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if got := flat["stfw_process_tool_tool_version"]; got != "v1.0.0" {
			t.Errorf("tool_version = %q, want v1.0.0 (tool: {version: ..} でも上書きさせない)", got)
		}
		if got := flat["stfw_process_tool_tool_version_label"]; got != "local" {
			t.Errorf("tool_version_label = %q, want local (別のキーは残す)", got)
		}
	})

	t.Run("stripProvisionKeys_別名が循環している場合_無限に展開せずエラーであること", func(t *testing.T) {
		// Arrange
		raw := []byte("stfw:\n  process:\n    tool:\n      tool_version: v1.0.0\n      loop: &l\n        - *l\n")

		// Act
		_, err := stripProvisionKeys(raw, "tool", []string{"tool_version"})

		// Assert
		if err == nil {
			t.Fatal("循環する別名はエラーにすべき (無限に展開しない)")
		}
	})
}
