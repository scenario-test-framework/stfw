package evidence

import "testing"

func TestHostFilePath(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		src     string
		want    string
		wantErr bool
	}{
		{name: "絶対パスを再現", host: "web01", src: "/var/log/app/access.log", want: "evidence/web01/var/log/app/access.log"},
		{name: "末尾スラッシュを正規化", host: "web01", src: "/var/log/", want: "evidence/web01/var/log"},
		{name: "相対パスは不可", host: "web01", src: "var/log/app.log", wantErr: true},
		{name: "ルートは不可", host: "web01", src: "/", wantErr: true},
		{name: "host に区切りは不可", host: "web/01", src: "/var/log/a.log", wantErr: true},
		{name: "host は空不可", host: "", src: "/var/log/a.log", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := HostFilePath(tc.host, tc.src)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("エラーを期待したが nil (got=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want=%q got=%q", tc.want, got)
			}
		})
	}
}

func TestHostFilePathTraversalStaysUnderEvidence(t *testing.T) {
	// 絶対パスの Clean は `..` がルートを越えないため、常に evidence/{host}/ 配下に収まる。
	got, err := HostFilePath("web01", "/var/log/../../../etc/passwd")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if got != "evidence/web01/etc/passwd" {
		t.Fatalf("パストラバーサルが evidence 配下に収まっていない: %q", got)
	}
}

func TestDatabaseTablePath(t *testing.T) {
	got, err := DatabaseTablePath("appdb", "orders")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if got != "evidence/appdb/orders.csv" {
		t.Fatalf("want=evidence/appdb/orders.csv got=%q", got)
	}

	if _, err := DatabaseTablePath("appdb", "ord/ers"); err == nil {
		t.Fatal("table に区切りを含む場合はエラーを期待")
	}
	if _, err := DatabaseTablePath("", "orders"); err == nil {
		t.Fatal("database が空の場合はエラーを期待")
	}
}
