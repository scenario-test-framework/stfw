package scenario

// DocData は `stfw scenario reverse` (tree → doc の投影) のレンダリング用データ。
// ScenarioView 同様、tree 走査 + metadata.yml / config.yml 読取 (repository の責務) の
// 結果を保持するだけの値オブジェクトであり、組み立てロジックは持たない。
type DocData struct {
	Name         string
	Description  string
	Traceability []TraceRow
	Bizdates     []DocBizdate
}

// TraceRow は要求トレーサビリティ表の 1 行 (要求仕様 1 件 → 検証する process 一覧)。
type TraceRow struct {
	RequirementSpecification string
	ProcessPaths             string // カンマ区切り ("{bizdate_dir}/{process_dir}" の列挙)
}

// DocBizdate は doc 上の業務日付 1 節分のデータ。
type DocBizdate struct {
	DirName   string
	Title     string // "{DirName} — {説明1行目}" (説明が空なら DirName のみ)
	Processes []DocProcess
}

// DocProcess は doc 上のプロセス 1 行 / 1 節分のデータ。
type DocProcess struct {
	SeqLabel                  string // "_" + seq (プロセス一覧表の # 列)
	DirName                   string // 子プロセスは "{親 dir}/{子 dir}" の連結表示 (repository が組み立てる)
	Group                     string // ディレクトリ名 `_{seq}_{group}_{type}` の group 部 (作者が宣言した確定値)
	Type                      string
	Description               string // 1 行 (説明の先頭行。テーブルセルが改行で壊れないよう畳む)
	RequirementSpecifications []string
	ConfigYAML                string       // config/config.yml の stfw.process.{type} サブツリー。空なら「設定」節を省略
	Children                  []DocProcess // parallel タイプのみ持てる子プロセス (AS-BUILT §4.14)
}
