package repository

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/scenario-test-framework/stfw/internal/domain/evidence"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
)

// workspaceDirName は run ディレクトリ配下の実行ワークスペースディレクトリ名。
const workspaceDirName = "workspace"

// WorkspaceDir は run の実行ワークスペースルート
// (.stfw/runs/{run_id}/workspace) を返す (AS-BUILT §5.7)。
func WorkspaceDir(projDir, runID string) string {
	return filepath.Join(runDir(projDir, runID), workspaceDirName)
}

// CopyScenarioToWorkspace は scenario/{name} を実行ワークスペースへ複製し、
// 複製先のシナリオディレクトリを返す。実行時生成物を run_id で名前空間化し、
// 同一シナリオを含む複数 run の並走を可能にする (AS-BUILT §5.7)。
// プロセスディレクトリ直下の予約出力ディレクトリ (evidence/actual/result) は
// 旧バージョンの実行残骸のため複製しない。
func CopyScenarioToWorkspace(projDir, runID, name string) (string, error) {
	src := filepath.Join(projDir, scenario.RootDirName, name)
	dest := filepath.Join(WorkspaceDir(projDir, runID), name)
	// 自己複製ガード用に複製先とその祖先の FileInfo を確定させる
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", fmt.Errorf("workspace copy: %w", err)
	}
	destAncestors, err := statAncestors(dest)
	if err != nil {
		return "", fmt.Errorf("workspace copy: %w", err)
	}
	if err := copyDirTree(src, dest, destAncestors, 0, "", nil, copyOpts{skipReserved: true}); err != nil {
		return "", fmt.Errorf("workspace copy: %w", err)
	}
	return dest, nil
}

// CheckRunWorkspaceScenario は resume の引き継ぎ元ワークスペースに対象シナリオが
// あるかを検証する (新 run_id 採番前の fail-fast 用。AS-BUILT §5.8)。
func CheckRunWorkspaceScenario(projDir, fromRunID, name string) error {
	src := filepath.Join(WorkspaceDir(projDir, fromRunID), name)
	if info, err := os.Stat(src); err != nil || !info.IsDir() {
		return fmt.Errorf("resume: scenario %s is not in run %s workspace", name, fromRunID)
	}
	return nil
}

// MergeRunWorkspace は fromRunID のワークスペースのシナリオツリーを、runID の
// ワークスペースへ**既存ファイルを上書きせずに**マージ取り込みする (resume。
// AS-BUILT §5.8)。CopyScenarioToWorkspace (正本の複製) の後に呼ぶ前提で、
// 正本が優先され、前 run にのみ存在する生成物 (evidence/ 等) が引き継がれる。
func MergeRunWorkspace(projDir, fromRunID, runID, name string) error {
	if err := CheckRunWorkspaceScenario(projDir, fromRunID, name); err != nil {
		return err
	}
	src := filepath.Join(WorkspaceDir(projDir, fromRunID), name)
	dest := filepath.Join(WorkspaceDir(projDir, runID), name)
	destAncestors, err := statAncestors(dest)
	if err != nil {
		return fmt.Errorf("resume merge: %w", err)
	}
	if err := copyDirTree(src, dest, destAncestors, 0, "", nil, copyOpts{skipExisting: true}); err != nil {
		return fmt.Errorf("resume merge: %w", err)
	}
	return nil
}

// copyOpts は copyDirTree の複製モード。
type copyOpts struct {
	// skipReserved はプロセス位置直下の予約出力ディレクトリを複製しない
	// (正本 → ワークスペースの複製。旧実行残骸の除外)。
	skipReserved bool
	// skipExisting は複製先に既存のファイル・ディレクトリ属性を変更しない
	// (前 run → ワークスペースのマージ。正本優先)。
	skipExisting bool
}

// statAncestors は path 自身とその全祖先ディレクトリの FileInfo を返す。
// パス文字列の比較は case-insensitive なファイルシステムで同一判定を誤るため、
// os.SameFile による物理同一性 (デバイス + inode) の比較に使う。
func statAncestors(path string) ([]os.FileInfo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for p := abs; ; {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
		parent := filepath.Dir(p)
		if parent == p {
			return infos, nil
		}
		p = parent
	}
}

// copyDirTree は src ディレクトリ配下を dest へ複製する。
// パーミッションは保持する。シンボリックリンクはリンク先の実体を複製する
// (リンクのまま複製すると、実行が正本や外部ツリーへ書き込む・相対リンクが
// 複製後に切れる、を防ぐ)。参照切れ・リンク循環・複製先の祖先を指すリンク先
// (自己複製) は複製時エラー。同一性判定はパス文字列ではなく os.SameFile で行う
// (case-insensitive ファイルシステム対策)。destAncestors は複製先ルートと
// その全祖先の FileInfo (自己複製の検出用)。depth はシナリオルートからの深さ
// (ルート = 0)。procType は src がディレクトリ規約上のプロセス位置 (bizdate 直下、
// または parallel の子) のときのプロセスタイプで、その直下に限り予約出力
// ディレクトリをスキップする。stack は走査中ディレクトリの FileInfo (循環検出用)。
func copyDirTree(src, dest string, destAncestors []os.FileInfo, depth int, procType string, stack []os.FileInfo, opts copyOpts) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	for _, anc := range stack {
		if os.SameFile(anc, srcInfo) {
			return fmt.Errorf("symlink cycle detected: %s", src)
		}
	}
	// リンク先が複製先ルートの祖先 (プロジェクトルート等) だと、走査が生成中の
	// ワークスペース自身へ到達して自己複製が無限に続くため、複製前に検出する
	for _, a := range destAncestors {
		if os.SameFile(srcInfo, a) {
			return fmt.Errorf("symlink resolves to an ancestor of the workspace destination: %s", src)
		}
	}
	stack = append(stack, srcInfo)

	destExisted := false
	if _, err := os.Stat(dest); err == nil {
		destExisted = true
	}
	// 読み取り専用ディレクトリでも配下を複製できるよう、書き込み可能な既定で
	// 作成し、配下の複製完了後に元のパーミッションへ合わせる。
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		p := filepath.Join(src, e.Name())
		// symlink はリンク先の実体で種別判定する
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if opts.skipReserved && procType != "" && isReservedOutputDir(e.Name()) {
				continue
			}
			// マージ時、複製先 (正本) 側に同名の非ディレクトリがある場合は
			// 正本優先で前 run のサブツリーを取り込まない (型競合)
			if opts.skipExisting {
				if dinfo, err := os.Lstat(filepath.Join(dest, e.Name())); err == nil && !dinfo.IsDir() {
					continue
				}
			}
			if err := copyDirTree(p, filepath.Join(dest, e.Name()), destAncestors, depth+1, childProcType(depth, filepath.Base(src), procType, e.Name()), stack, opts); err != nil {
				return err
			}
			continue
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type (not a regular file): %s", p)
		}
		if err := copyFile(p, filepath.Join(dest, e.Name()), info.Mode().Perm(), opts.skipExisting); err != nil {
			return err
		}
	}
	// マージ時は既存ディレクトリの属性を変更しない (正本優先)
	if destExisted && opts.skipExisting {
		return nil
	}
	return os.Chmod(dest, srcInfo.Mode().Perm())
}

// childProcType は子ディレクトリがディレクトリ規約上のプロセス位置に当たる場合、
// そのプロセスタイプを返す (それ以外は空)。プロセス位置は次の 2 つのみ:
//   - 業務日付ディレクトリ (深さ 1・bizdate 形式) 直下のプロセスディレクトリ
//   - parallel プロセス直下の子プロセスディレクトリ (AS-BUILT §4.14)
//
// 入力ディレクトリ (data/ 等) の配下にプロセス形式の名前があっても
// プロセス位置ではないため、その直下の evidence 等はスキップしない。
func childProcType(parentDepth int, parentName, parentProcType, name string) string {
	underBizdate := false
	if parentDepth == 1 {
		_, _, err := scenario.ParseBizdateDirName(parentName)
		underBizdate = err == nil
	}
	if !underBizdate && parentProcType != scenario.ParallelProcessType {
		return ""
	}
	_, _, typ, err := scenario.ParseProcessDirName(name)
	if err != nil {
		return ""
	}
	return typ
}

// isReservedOutputDir は実行時にのみ生成される予約出力ディレクトリ名かを返す
// (エビデンス規約。AS-BUILT §4.7)。
func isReservedOutputDir(name string) bool {
	return name == evidence.DirName || name == evidence.ActualDirName || name == evidence.ResultDirName
}

// copyFile は通常ファイルを複製する。パーミッションは perm を保持する。
// skipExisting の場合、複製先に既存のファイルは上書きせずスキップする。
func copyFile(src, dest string, perm os.FileMode, skipExisting bool) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		if skipExisting && os.IsExist(err) {
			return nil
		}
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	// O_CREATE のモードは umask の影響を受けるため、明示的に合わせる
	if err := out.Chmod(perm); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
