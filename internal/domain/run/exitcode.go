package run

// ExitCode はプロセスプラグイン実行契約の終了コード。
// v0.2 (Bash 版 setenv) の EXITCODE_* と互換。
type ExitCode int

const (
	ExitSuccess ExitCode = 0
	ExitWarn    ExitCode = 3
	ExitError   ExitCode = 6
)

// Int は os.Exit へ渡す値を返す。
func (c ExitCode) Int() int { return int(c) }
