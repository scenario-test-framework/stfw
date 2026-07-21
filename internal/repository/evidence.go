package repository

import (
	"path/filepath"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/evidence"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// EvidenceDir は収集系プラグインの出力ルート (processDir/evidence) の
// ディスクパスを返す。エビデンスディレクトリ規約の基点。
func EvidenceDir(processDir string) string {
	return filepath.Join(processDir, evidence.DirName)
}

// CompareDirs は compare プラグインのプロセスディレクトリ配下の
// expect / actual / result のディスクパスを返す。
func CompareDirs(processDir string) (expect, actual, result string) {
	return filepath.Join(processDir, evidence.ExpectDirName),
		filepath.Join(processDir, evidence.ActualDirName),
		filepath.Join(processDir, evidence.ResultDirName)
}

// forbiddenConnKeySegments は config.yml への直書きを禁止する接続情報キーの
// 末尾セグメント (条件「プラグイン接続情報のグループ名参照」)。ホスト・パスワードは
// inventory グループ名参照 + secret の {host}-{user} 自動参照で解決させる。
// host_group は末尾が group のため該当しない。
var forbiddenConnKeySegments = map[string]bool{
	"host":     true,
	"hosts":    true,
	"password": true,
	"passwd":   true,
}

// ForbiddenConnConfig は接続情報を config.yml に直書きした違反 1 件。
type ForbiddenConnConfig struct {
	ProcessPath string // 表示用のプロセスディレクトリパス (scenario/.../{process})
	Key         string // 違反したフラット化済みキー
}

// CheckForbiddenConnConfig はシナリオ配下の各プロセス (parallel の子を含む) の
// 実効設定を検査し、接続情報 (host / hosts / password / passwd) を直書きした違反を返す。
// 設定は環境非依存の静的性質のため、validate・run 実行前ゲートの双方で
// エラーとして扱う。プラグインを解決できないプロセスはスキップする。
func CheckForbiddenConnConfig(projDir string, views []scenario.ScenarioView) ([]ForbiddenConnConfig, error) {
	var found []ForbiddenConnConfig
	for _, sv := range views {
		for _, bv := range sv.Bizdates {
			for _, pv := range bv.Processes {
				f, err := checkProcessConnConfig(projDir, []string{scenario.RootDirName, sv.Name, bv.DirName}, pv)
				if err != nil {
					return nil, err
				}
				found = append(found, f...)
			}
		}
	}
	return found, nil
}

// checkProcessConnConfig はプロセス 1 件 (parallel の場合は子を含む) の実効設定を検査する。
// parents はプロジェクトルート相対の親パス要素列。
func checkProcessConnConfig(projDir string, parents []string, pv scenario.ProcessView) ([]ForbiddenConnConfig, error) {
	var found []ForbiddenConnConfig

	// ディレクトリ名 parse error のプロセスは ProcessType="" になる。
	// 構造検証側で別途 error になるため、ここでは検査対象外にする
	// (空タイプで設定を読むと無関係な誤検出を生むため)。
	if pv.ProcessType == "" {
		return nil, nil
	}
	segments := append(append([]string(nil), parents...), pv.DirName)
	if loc, err := ResolveProcessPlugin(projDir, pv.ProcessType); err == nil {
		processDir := filepath.Join(append([]string{projDir}, segments...)...)
		flat, err := ProcessConfigEnv(projDir, loc, pv.ProcessType, processDir)
		if err != nil {
			return nil, err
		}
		display := strings.Join(segments, "/")
		for key := range flat {
			seg := key
			if i := strings.LastIndex(key, "_"); i >= 0 {
				seg = key[i+1:]
			}
			if forbiddenConnKeySegments[seg] {
				found = append(found, ForbiddenConnConfig{ProcessPath: display, Key: key})
			}
		}
	}

	for _, cv := range pv.Children {
		f, err := checkProcessConnConfig(projDir, segments, cv)
		if err != nil {
			return nil, err
		}
		found = append(found, f...)
	}
	return found, nil
}
