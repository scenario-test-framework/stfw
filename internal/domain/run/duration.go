package run

import (
	"fmt"
	"time"
)

// ElapsedString は開始・終了タイムスタンプ (TSFormat) の経過時間を
// v0.2 の private.calc_processing_time と同じ `HH:MM:SS` 形式で返す
// (時は 2 桁ゼロ埋め。24 時間を超える場合はそのまま加算表記)。
func ElapsedString(startTS, endTS string) (string, error) {
	start, err := time.Parse(TSFormat, startTS)
	if err != nil {
		return "", fmt.Errorf("start_ts %s: %w", startTS, err)
	}
	end, err := time.Parse(TSFormat, endTS)
	if err != nil {
		return "", fmt.Errorf("end_ts %s: %w", endTS, err)
	}
	sec := int(end.Sub(start) / time.Second)
	if sec < 0 {
		return "", fmt.Errorf("end_ts %s is before start_ts %s", endTS, startTS)
	}
	return fmt.Sprintf("%02d:%02d:%02d", sec/3600, sec/60%60, sec%60), nil
}
