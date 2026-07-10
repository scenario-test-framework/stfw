package scenario

// Phase はプロセスタイプから推定した実行フェーズ (doc 投影の表示用)。
// v0.2 の 4 フェーズ (Arrange/Act/Collect/Assert) にプロセスタイプを固定マッピングする。
// scripts やプロジェクト独自のプラグインタイプはこのマッピングに現れないため PhaseUnknown になる。
type Phase string

const (
	PhaseArrange Phase = "Arrange"
	PhaseAct     Phase = "Act"
	PhaseCollect Phase = "Collect"
	PhaseAssert  Phase = "Assert"
	// PhaseUnknown は組込みフェーズへ推定できないプロセスタイプ (scripts・ユーザー定義) を表す。
	PhaseUnknown Phase = "-"
)

// phaseByType は組込みプロセスプラグインの type → フェーズ 固定表。
// AS-BUILT.md §4.2 (プロセス 5 フェーズ) の Arrange/Act/Collect/Assert と対応する
// (Pre/Post は共通フックであり特定 type を持たないためここには現れない)。
var phaseByType = map[string]Phase{
	"importMysql":    PhaseArrange,
	"importPostgres": PhaseArrange,
	"importRedis":    PhaseArrange,
	"clearMysql":     PhaseArrange,
	"clearPostgres":  PhaseArrange,
	"clearRedis":     PhaseArrange,
	"scpPut":         PhaseArrange,
	"invokeRest":     PhaseAct,
	"invokeWeb":      PhaseAct,
	"sshExec":        PhaseAct,
	"collectLog":     PhaseCollect,
	"collectFile":    PhaseCollect,
	"exportMysql":    PhaseCollect,
	"exportPostgres": PhaseCollect,
	"exportRedis":    PhaseCollect,
	"compare":        PhaseAssert,
}

// PhaseOf はプロセスタイプから実行フェーズを推定する。
// 未知の type (scripts・ユーザー定義プラグイン) は PhaseUnknown を返す。
func PhaseOf(processType string) Phase {
	if p, ok := phaseByType[processType]; ok {
		return p
	}
	return PhaseUnknown
}

// String はフェーズの表示用文字列を返す。
func (p Phase) String() string { return string(p) }
