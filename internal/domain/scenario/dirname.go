package scenario

import (
	"fmt"
	"strings"
)

// 階層ディレクトリ命名規約 (v0.2 互換のディレクトリ規約):
//
//	scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/
//
// bizdate 階層は `_{seq}_{bizdate}`、process 階層は `_{seq}_{group}_{type}`。
// v0.2 の bizdate_spec.dirname / process_spec.dirname と 1:1 対応する。

// BizdateDirName は bizdate 階層のディレクトリ名 `_{seq}_{bizdate}` を組み立てる。
func BizdateDirName(seq Seq, bizdate Bizdate) string {
	return "_" + seq.String() + "_" + bizdate.String()
}

// ParseBizdateDirName はディレクトリ名を `_{seq}_{bizdate}` としてパースする。
// v0.2 の dig_repository は cut -d '_' -f 2,3 で切り出すのみで検証しなかったが、
// v1.0 では seq / bizdate の値オブジェクト検証まで行う。
func ParseBizdateDirName(name string) (Seq, Bizdate, error) {
	parts := strings.Split(name, "_")
	if len(parts) != 3 || parts[0] != "" {
		return Seq{}, Bizdate{}, fmt.Errorf("%s is not _{seq}_{bizdate} format", name)
	}
	seq, err := NewSeq(parts[1])
	if err != nil {
		return Seq{}, Bizdate{}, err
	}
	bizdate, err := NewBizdate(parts[2])
	if err != nil {
		return Seq{}, Bizdate{}, err
	}
	return seq, bizdate, nil
}

// ProcessDirName はプロセス階層のディレクトリ名 `_{seq}_{group}_{type}` を組み立てる。
func ProcessDirName(seq Seq, group Group, processType string) string {
	return "_" + seq.String() + "_" + group.String() + "_" + processType
}

// ParseProcessDirName はディレクトリ名を `_{seq}_{group}_{type}` としてパースする。
// v0.2 の process_spec.type は cut -d '_' -f 4 で切り出していた
// (group / type に `_` を含められないのはこのパースの保護)。
func ParseProcessDirName(name string) (Seq, Group, string, error) {
	parts := strings.Split(name, "_")
	if len(parts) != 4 || parts[0] != "" {
		return Seq{}, Group{}, "", fmt.Errorf("%s is not _{seq}_{group}_{type} format", name)
	}
	seq, err := NewSeq(parts[1])
	if err != nil {
		return Seq{}, Group{}, "", err
	}
	group, err := NewGroup(parts[2])
	if err != nil {
		return Seq{}, Group{}, "", err
	}
	if err := ValidateProcessType(parts[3]); err != nil {
		return Seq{}, Group{}, "", err
	}
	return seq, group, parts[3], nil
}

// ValidateProcessType はプロセスタイプ名の形式を検証する。
// ディレクトリ名の最終フィールドになるため `_` とパス区切り文字を含められない。
func ValidateProcessType(processType string) error {
	if processType == "" {
		return fmt.Errorf("process_type must not null")
	}
	if strings.Contains(processType, "_") {
		return fmt.Errorf("%q can not contains %q", processType, "_")
	}
	if strings.ContainsAny(processType, `/\`) {
		return fmt.Errorf("%q can not contains path separator", processType)
	}
	return nil
}
