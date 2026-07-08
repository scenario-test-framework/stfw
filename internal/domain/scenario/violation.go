package scenario

import "fmt"

// ViolationLevel は規約違反の深刻度。
type ViolationLevel string

const (
	// ViolationError は実行不能な規約違反。validate は exit 6 で終了する。
	ViolationError ViolationLevel = "error"
	// ViolationWarn は実行可能だが望ましくない状態。validate は exit 0 のまま警告する。
	ViolationWarn ViolationLevel = "warn"
)

// Violation はディレクトリ規約違反 1 件。
type Violation struct {
	// Path は違反対象の識別子。ディレクトリ規約違反ではプロジェクトルートからの
	// 相対パス、プラグイン依存違反 (requires) のようにディレクトリに紐づかない
	// 違反ではプロセスタイプ名を入れる。
	Path    string
	Level   ViolationLevel
	Message string
}

// String は表示用の 1 行文字列を返す。
func (v Violation) String() string {
	return fmt.Sprintf("[%s] %s: %s", v.Level, v.Path, v.Message)
}

// Violations は規約違反のコレクション。
type Violations []Violation

// HasError はエラーレベルの違反を含むかを返す。
func (vs Violations) HasError() bool {
	for _, v := range vs {
		if v.Level == ViolationError {
			return true
		}
	}
	return false
}

// Count はレベル別の件数 (エラー, 警告) を返す。
func (vs Violations) Count() (errors, warns int) {
	for _, v := range vs {
		switch v.Level {
		case ViolationError:
			errors++
		case ViolationWarn:
			warns++
		}
	}
	return errors, warns
}
