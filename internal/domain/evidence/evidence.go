// Package evidence はエビデンスディレクトリ規約 (ディレクトリ規約・プラグイン env 契約に
// 次ぐ第 3 の互換境界) のパス構築・検証ロジックを持つ。
// 収集系プラグイン (collectFile / collectLog / exportXxx) の出力ルートは
// 自プロセスディレクトリ配下の evidence/ (gitignore 対象) とする。
package evidence

import (
	"fmt"
	"path"
	"strings"
)

// DirName はプロセスディレクトリ配下のエビデンス出力ルートディレクトリ名。
const DirName = "evidence"

// compare プラグインのプロセスディレクトリ配下の規約ディレクトリ名。
//   - ExpectDirName: 期待値 (git 管理)。直下は「同一 bizdate 内の収集系
//     process ディレクトリ名」で、その配下は当該 process の evidence/ と同型。
//   - ActualDirName: 実測値 (gitignore・自動生成)。expect と同じ構造で、実体は
//     各収集系 process の evidence/ への symlink。
//   - ResultDirName: compare-files の比較結果出力 (gitignore)。
const (
	ExpectDirName = "expect"
	ActualDirName = "actual"
	ResultDirName = "result"
)

// HostFilePath は evidence/{host}/{収集元の絶対パスをそのまま再現} の
// 相対パス (スラッシュ区切り) を構築する。collectFile / collectLog の出力先規約。
// srcPath は収集元ホスト上の絶対パスであること。
func HostFilePath(host, srcPath string) (string, error) {
	if err := validateSegment("host", host); err != nil {
		return "", err
	}
	if !strings.HasPrefix(srcPath, "/") {
		return "", fmt.Errorf("source path must be absolute: %s", srcPath)
	}
	// 絶対パスの Clean は `..` がルートを越えないため、再現パスは
	// 必ず evidence/{host}/ 配下に収まる
	cleaned := path.Clean(srcPath)
	if cleaned == "/" {
		return "", fmt.Errorf("source path must not be root: %s", srcPath)
	}
	return path.Join(DirName, host, cleaned), nil
}

// DatabaseTablePath は evidence/{database}/{table}.csv の相対パス
// (スラッシュ区切り) を構築する。exportMysql / exportPostgres の出力先規約。
func DatabaseTablePath(database, table string) (string, error) {
	if err := validateSegment("database", database); err != nil {
		return "", err
	}
	if err := validateSegment("table", table); err != nil {
		return "", err
	}
	return path.Join(DirName, database, table+".csv"), nil
}

// validateSegment はパスセグメントとして安全な値であることを検証する
// (空・パス区切り・`.` `..` の禁止)。
func validateSegment(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	if strings.ContainsAny(value, `/\`) || value == "." || value == ".." {
		return fmt.Errorf("%s contains invalid characters: %s", field, value)
	}
	return nil
}
