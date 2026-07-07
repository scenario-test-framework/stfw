package scenario

import "strings"

// 階層ディレクトリ判定 (v0.2 互換の条件):
// scenario / bizdate / process ディレクトリは「scenario ルートからの深さ」で判定する。
// v0.2 の scenario_spec / bizdate_spec / process_spec の is_*-dir 判定に対応する。
// 「プロジェクト直下の stfw.yml の存在」の確認は I/O を伴うため呼び出し側 (usecase) が行う。

// RootDirName はプロジェクト直下のシナリオルートディレクトリ名。
const RootDirName = "scenario"

// IsScenarioRootDir は rel (プロジェクトルートからの相対パス、`/` 区切り) が
// シナリオルートディレクトリかを判定する。
func IsScenarioRootDir(rel string) bool {
	return rel == RootDirName
}

// IsScenarioDir は rel がシナリオディレクトリ (scenario/{name}) かを判定する。
func IsScenarioDir(rel string) bool {
	parts := splitRel(rel)
	return len(parts) == 2 && parts[0] == RootDirName
}

// IsBizdateDir は rel が業務日付ディレクトリ (scenario/{name}/_{seq}_{bizdate}) かを判定する。
func IsBizdateDir(rel string) bool {
	parts := splitRel(rel)
	return len(parts) == 3 && parts[0] == RootDirName
}

// IsProcessDir は rel がプロセスディレクトリ
// (scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}) かを判定する。
func IsProcessDir(rel string) bool {
	parts := splitRel(rel)
	return len(parts) == 4 && parts[0] == RootDirName
}

// splitRel は相対パスを要素に分解する。空文字・"." は要素なしとみなす。
func splitRel(rel string) []string {
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." {
		return nil
	}
	return strings.Split(rel, "/")
}
