package run

import (
	"testing"
	"time"
)

func TestCompareRunID(t *testing.T) {
	mustParse := func(t *testing.T, s string) RunID {
		t.Helper()
		id, err := ParseRunID(s)
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	tests := []struct {
		name string
		a, b string
		want int
	}{
		// 文字列比較では "9855" > "10006" となり新旧が逆転する (CI で発生した組み合わせ)
		{"CompareRunID_同一採番時刻でpidの桁数が異なる場合_pidの数値順であること", "_20261117235333_9855", "_20261117235333_10006", -1},
		{"CompareRunID_採番時刻が異なる場合_pidに関わらず採番時刻順であること", "_20261117235333_99999", "_20261117235334_1", -1},
		{"CompareRunID_pidに先頭ゼロがある場合_数値として比較されること", "_20261117235333_0099", "_20261117235333_100", -1},
		{"CompareRunID_同一のrun_idの場合_0であること", "_20261117235333_42", "_20261117235333_42", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			a, b := mustParse(t, tt.a), mustParse(t, tt.b)
			// Act
			got, reversed := CompareRunID(a, b), CompareRunID(b, a)
			// Assert
			if got != tt.want || reversed != -tt.want {
				t.Errorf("CompareRunID(%s, %s) = %d (reversed %d), want %d", tt.a, tt.b, got, reversed, tt.want)
			}
		})
	}
}

func TestNextIssueTime(t *testing.T) {
	now := time.Date(2026, 11, 17, 23, 53, 33, 700_000_000, time.Local)
	sec := now.Truncate(time.Second)
	tests := []struct {
		name   string
		latest time.Time // ゼロ値 = 既存 run なし
		want   time.Time
	}{
		{"NextIssueTime_既存runが無い場合_現在時刻の秒単位であること", time.Time{}, sec},
		{"NextIssueTime_最新runが現在より前の場合_現在時刻であること", sec.Add(-time.Second), sec},
		{"NextIssueTime_最新runが同一秒の場合_最新runの1秒後であること", sec, sec.Add(time.Second)},
		{"NextIssueTime_最新runが24時間以内の未来の場合_最新runの1秒後であること", sec.Add(24 * time.Hour), sec.Add(24*time.Hour + time.Second)},
		{"NextIssueTime_最新runが24時間を超える未来の場合_現在時刻であること", sec.Add(24*time.Hour + time.Second), sec},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			latest := tt.latest
			// Act
			got := NextIssueTime(now, latest)
			// Assert
			if !got.Equal(tt.want) {
				t.Errorf("NextIssueTime(%v, %v) = %v, want %v", now, latest, got, tt.want)
			}
		})
	}

	t.Run("NextIssueTime_同一秒に連続して多数採番する場合_採番時刻が単調に増加し続けること", func(t *testing.T) {
		// Arrange
		// 1 秒未満の run を連続実行すると繰り上げで実時刻より先行する。
		// 先行が上限に届いて現在時刻へ戻ると新旧が逆転するため、実運用の回数では戻らないこと
		const runs = 1000
		latest := time.Time{}
		// Act
		var issued []time.Time
		for i := 0; i < runs; i++ {
			latest = NextIssueTime(now, latest)
			issued = append(issued, latest)
		}
		// Assert
		for i := 1; i < len(issued); i++ {
			if !issued[i].After(issued[i-1]) {
				t.Fatalf("%d 回目の採番時刻 %v が直前 %v より後でない", i+1, issued[i], issued[i-1])
			}
		}
	})
}
