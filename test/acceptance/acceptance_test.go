// Package acceptance は integration_test.sh (v0.2) を翻訳した受け入れテスト。
// testscript で実際の stfw CLI をエンドツーエンドに検証する。
package acceptance

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/scenario-test-framework/stfw/internal/presentation/cli"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"stfw": cli.Execute,
	}))
}

func TestAcceptance(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			"normjournal": cmdNormJournal,
			"latestrun":   cmdLatestRun,
			"execcode":    cmdExecCode,
		},
	})
}

// cmdExecCode はコマンドを実行し、終了コードが期待値と一致することを検証する
// (`! exec` は非 0 しか検証できないため、exit code 6 等の互換確認に使う)。
//
//	使い方: execcode <want> <command> [args...]
func cmdExecCode(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) < 2 {
		ts.Fatalf("usage: execcode <want> <command> [args...]")
	}
	want, err := strconv.Atoi(args[0])
	if err != nil {
		ts.Fatalf("execcode: invalid exit code %q", args[0])
	}

	got := 0
	if execErr := ts.Exec(args[1], args[2:]...); execErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(execErr, &exitErr) {
			ts.Fatalf("execcode: %v", execErr)
		}
		got = exitErr.ExitCode()
	}
	if got != want {
		ts.Fatalf("exit code = %d, want %d", got, want)
	}
}

// runIDPattern は run_id (`_{yyyymmddhhmmss}_{pid}`) の出現箇所。
var runIDPattern = regexp.MustCompile(`_\d{14}_\d+`)

// cmdNormJournal は最新 run のジャーナルを正規化して outfile へ書き出す。
//
//	使い方: normjournal <projdir> <outfile>
//
// 正規化規則 (ゴールデン比較用):
//   - ts / start_ts / end_ts を除去する (実行時刻に依存するため)
//   - run_id を RUN_ID へ置換する (採番時刻・pid に依存するため)
//   - キーはアルファベット順で再整列する
func cmdNormJournal(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 2 {
		ts.Fatalf("usage: normjournal <projdir> <outfile>")
	}
	projDir := ts.MkAbs(args[0])

	runID, err := repository.LatestRunID(projDir)
	ts.Check(err)
	raw, err := os.ReadFile(repository.JournalPath(projDir, runID))
	ts.Check(err)

	var out bytes.Buffer
	for _, line := range bytes.Split(raw, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal(line, &event); err != nil {
			ts.Fatalf("journal line %q: %v", line, err)
		}
		delete(event, "ts")
		delete(event, "start_ts")
		delete(event, "end_ts")
		normalized, err := json.Marshal(event)
		ts.Check(err)
		out.Write(runIDPattern.ReplaceAll(normalized, []byte("RUN_ID")))
		out.WriteByte('\n')
	}
	ts.Check(os.WriteFile(ts.MkAbs(args[1]), out.Bytes(), 0o644))
}

// cmdLatestRun は最新 run の run_id を環境変数へ設定する
// (HTML レポートのパス runs/{run_id}.html の検証用)。
//
//	使い方: latestrun <projdir> <envvar>
func cmdLatestRun(ts *testscript.TestScript, neg bool, args []string) {
	if neg || len(args) != 2 {
		ts.Fatalf("usage: latestrun <projdir> <envvar>")
	}
	runID, err := repository.LatestRunID(ts.MkAbs(args[0]))
	ts.Check(err)
	ts.Setenv(args[1], runID)
}
