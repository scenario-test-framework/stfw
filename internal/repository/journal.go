package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// runsDirName は .stfw 配下の実行ジャーナル配置ディレクトリ名
// (v0.2 の run_spec.digdag_workspace_dir と同じ .stfw/runs)。
const runsDirName = "runs"

// journalFileName は実行ジャーナルのファイル名。
const journalFileName = "journal.jsonl"

// runDir は run_id の実行データディレクトリを返す。
func runDir(projDir, runID string) string {
	return filepath.Join(projDir, project.DataDirName, runsDirName, runID)
}

// RunDir は run_id の実行データディレクトリ (.stfw/runs/{run_id}) を返す。
// 実行ワークスペース・run 単位のプラグイン展開先の親ディレクトリ (AS-BUILT §5.7)。
func RunDir(projDir, runID string) string {
	return runDir(projDir, runID)
}

// JournalPath は journal.jsonl のパスを返す。
func JournalPath(projDir, runID string) string {
	return filepath.Join(runDir(projDir, runID), journalFileName)
}

// Journal は journal.jsonl への追記専用ライター。
// 1 イベント = 1 行で追記し、行ごとに flush する (実行中でもリプレイ可能)。
type Journal struct {
	f *os.File
}

// CreateJournal は .stfw/runs/{run_id}/journal.jsonl を新規作成する。
// 同一 run_id のディレクトリが既に存在する場合は fs.ErrExist を返す。
func CreateJournal(projDir string, runID run.RunID) (*Journal, error) {
	dir := runDir(projDir, runID.String())
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return nil, err
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, journalFileName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Journal{f: f}, nil
}

// Append はイベントを 1 行追記して flush する。
func (j *Journal) Append(ev run.Event) error {
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if _, err := j.f.Write(append(raw, '\n')); err != nil {
		return err
	}
	return j.f.Sync()
}

// Close はジャーナルを閉じる。
func (j *Journal) Close() error {
	return j.f.Close()
}

// ReadJournal は journal.jsonl を読み込んでイベント列を返す (リプレイ入力)。
func ReadJournal(projDir, runID string) ([]run.Event, error) {
	path := JournalPath(projDir, runID)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("run %s: %w", runID, err)
	}
	defer func() { _ = f.Close() }()

	var events []run.Event
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var ev run.Event
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, line, err)
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return events, nil
}

// ListRunIDs は実行済みの run_id を昇順で返す。ディレクトリが無い場合は空を返す。
func ListRunIDs(projDir string) ([]string, error) {
	root := filepath.Join(projDir, project.DataDirName, runsDirName)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := run.ParseRunID(e.Name()); err != nil {
			continue
		}
		ids = append(ids, e.Name())
	}
	sort.Strings(ids)
	return ids, nil
}

// LatestRunID は最新 (run_id 昇順の末尾 = 採番時刻が最も新しい) の run_id を返す。
func LatestRunID(projDir string) (string, error) {
	ids, err := ListRunIDs(projDir)
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("no runs in %s", filepath.Join(projDir, project.DataDirName, runsDirName))
	}
	return ids[len(ids)-1], nil
}
