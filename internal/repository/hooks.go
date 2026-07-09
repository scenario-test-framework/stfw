package repository

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// ListHookScripts はプロジェクトの階層フック
// plugins/{level}/_common/{phase}/ 直下のスクリプトを昇順の絶対パスで返す。
// ディレクトリが無い場合は空を返す (フック未定義は正常)。
// v0.2 のプロジェクトプラグイン (project_plugin_service + stfw.bulk_exec_scripts)
// の走査規則 (直下ファイルのみ・名前昇順) に対応する。
func ListHookScripts(projDir string, level run.NodeType, phase string) ([]string, error) {
	return listFilesAsc(filepath.Join(projDir, "plugins", string(level), "_common", phase))
}

// ListScriptSteps は scripts プロセスのステップスクリプト名を昇順で返す。
// v0.2 の plugin.process.scripts.list_files (scripts/ 直下・ファイルのみ・名前昇順) と同じ規則。
func ListScriptSteps(processDir string) ([]string, error) {
	paths, err := listFilesAsc(filepath.Join(processDir, "scripts"))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(paths))
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	return names, nil
}

// EnsureStepScriptsExecutable は scripts/ 直下のステップへ実行権限を付与する
// (v0.2 の scripts プラグイン pre_execute の chmod +x 相当)。
func EnsureStepScriptsExecutable(processDir string) error {
	paths, err := listFilesAsc(filepath.Join(processDir, "scripts"))
	if err != nil {
		return err
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if info.Mode()&0o111 != 0 {
			continue
		}
		if err := os.Chmod(p, info.Mode()|0o755); err != nil {
			return err
		}
	}
	return nil
}

// listFilesAsc は dir 直下のファイル (symlink 先がファイルのものを含む) を
// 名前昇順の絶対パスで返す。dir が無い場合は空を返す。
func listFilesAsc(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		info, err := os.Stat(path) // symlink follow (v0.2 の find -follow 相当)
		if err != nil || info.IsDir() {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}
