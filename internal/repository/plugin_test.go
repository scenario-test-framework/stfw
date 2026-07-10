package repository

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveProcessPluginEmptyType は空のプロセスタイプが plugins/process
// ディレクトリ自体に解決されないこと (parse error プロセスへの防御) を固定する。
func TestResolveProcessPluginEmptyType(t *testing.T) {
	t.Run("ResolveProcessPlugin_空のプロセスタイプの場合_エラーであること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		// plugins/process ディレクトリを実在させる (空タイプの filepath.Join が
		// ここに解決してしまう退行を検出するため)。
		if err := os.MkdirAll(filepath.Join(projDir, "plugins", "process"), 0o755); err != nil {
			t.Fatal(err)
		}

		// Act
		_, err := ResolveProcessPlugin(projDir, "")

		// Assert
		if err == nil {
			t.Fatal("空のプロセスタイプは解決エラーになるべき (plugins/process への誤解決を防ぐ)")
		}
	})
}

// TestCreateProcessScaffoldTemplatePlugin は template/ を持つ組込みプラグイン
// (scripts) が template 内容を展開することを固定する。
func TestCreateProcessScaffoldTemplatePlugin(t *testing.T) {
	t.Run("CreateProcessScaffold_template付きプラグインの場合_template内容が展開されること", func(t *testing.T) {
		// Arrange
		loc, err := ResolveProcessPlugin(t.TempDir(), "scripts")
		if err != nil {
			t.Fatal(err)
		}
		if !PluginHasTemplate(loc) {
			t.Fatal("scripts プラグインは template/ を持つべき")
		}
		parent := t.TempDir()

		// Act
		_, err = CreateProcessScaffold(loc, parent, "_10_pre_scripts")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(parent, "_10_pre_scripts")
		for _, rel := range []string{"config/config.yml", "scripts/100_1st_step", "metadata.yml"} {
			if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
				t.Errorf("template 展開で %s が作られるべき: %v", rel, err)
			}
		}
	})
}

// TestCreateProcessScaffoldConfigPlugin は template/ を持たない config 駆動の
// 組込みプラグイン (clearPostgres) が、デフォルト config.yml を
// config/config.yml として配置し、scripts/ を作らないことを固定する。
// (組込みプラグインで template/ 不在により new process が失敗していた退行の防止)
func TestCreateProcessScaffoldConfigPlugin(t *testing.T) {
	t.Run("CreateProcessScaffold_config駆動プラグインの場合_config.ymlを配置しscriptsを作らないこと", func(t *testing.T) {
		// Arrange
		loc, err := ResolveProcessPlugin(t.TempDir(), "clearPostgres")
		if err != nil {
			t.Fatal(err)
		}
		if PluginHasTemplate(loc) {
			t.Fatal("clearPostgres プラグインは template/ を持たないはず")
		}
		parent := t.TempDir()

		// Act
		created, err := CreateProcessScaffold(loc, parent, "_20_db_clearPostgres")

		// Assert
		if err != nil {
			t.Fatalf("template/ 不在でも scaffold は成功するべき: %v", err)
		}
		dir := filepath.Join(parent, "_20_db_clearPostgres")
		if _, err := os.Stat(filepath.Join(dir, "config", "config.yml")); err != nil {
			t.Errorf("デフォルト config.yml が config/config.yml として配置されるべき: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "metadata.yml")); err != nil {
			t.Errorf("metadata.yml が作られるべき: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "scripts")); !os.IsNotExist(err) {
			t.Errorf("config 駆動プラグインでは scripts/ は作られないべき")
		}
		if len(created) == 0 {
			t.Error("作成ファイル一覧が空であるべきでない")
		}
	})
}
