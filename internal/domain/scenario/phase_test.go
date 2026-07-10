package scenario

import "testing"

// GUIDE.md §2 (フェーズと組込みプラグイン) の表と同じマッピングであることを固定する。
func TestPhaseOf(t *testing.T) {
	tests := map[string]Phase{
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
		// scripts (汎用) とユーザー定義 type は PhaseUnknown ("-")
		"scripts":      PhaseUnknown,
		"myCustomType": PhaseUnknown,
		"":             PhaseUnknown,
	}
	for processType, want := range tests {
		if got := PhaseOf(processType); got != want {
			t.Errorf("PhaseOf(%q) = %q, want %q", processType, got, want)
		}
	}
}

func TestPhaseString(t *testing.T) {
	if PhaseArrange.String() != "Arrange" {
		t.Errorf("PhaseArrange.String() = %q, want %q", PhaseArrange.String(), "Arrange")
	}
	if PhaseUnknown.String() != "-" {
		t.Errorf("PhaseUnknown.String() = %q, want %q", PhaseUnknown.String(), "-")
	}
}
