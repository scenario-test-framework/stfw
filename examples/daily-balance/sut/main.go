// daily-balance-sut は example 用のトイ「テスト対象システム (SUT)」。
// 口座残高を PostgreSQL で管理する最小の REST API。
//
//	POST /transactions  {"account_id","amount"}  取引を記録し残高を加減算
//	GET  /accounts                               口座残高一覧 (デバッグ用)
//	GET  /healthz                                ヘルスチェック
//
// 取引の業務日付は payload では受け取らず、業務日付テーブル biz_calendar (単一行
// id='system') から解決する。biz_calendar はテスト側のカスタムプラグイン updateBizdate
// が業務日付ごとに更新する。
//
// これは stfw の example を end-to-end で動かすためのダミー実装であり、
// 認証・バリデーション・トランザクション分離などは意図的に最小限にしている。
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type transaction struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
}

type account struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "self HTTP health probe (for container HEALTHCHECK)")
	flag.Parse()
	if *healthcheck {
		resp, err := http.Get("http://127.0.0.1:8080/healthz")
		if err != nil {
			log.Fatalf("healthcheck: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("healthcheck: status %d", resp.StatusCode)
		}
		return
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://appuser:apppass@postgres:5432/appdb?sslmode=disable"
	}

	db := mustConnect(dsn)
	defer db.Close()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var tx transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if tx.AccountID == "" {
			http.Error(w, "account_id is required", http.StatusBadRequest)
			return
		}
		// 業務日付は biz_calendar (updateBizdate が更新する単一行) から解決する。
		// 未設定は運転前提の不備なので取引を受け付けない。
		var bizdate string
		if err := db.QueryRow(
			`SELECT bizdate FROM biz_calendar WHERE id = 'system'`,
		).Scan(&bizdate); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "business date is not set (biz_calendar)", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := db.Exec(
			`INSERT INTO transactions (account_id, amount, bizdate) VALUES ($1, $2, $3)`,
			tx.AccountID, tx.Amount, bizdate,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := db.Exec(
			`UPDATE accounts SET balance = balance + $1 WHERE id = $2`,
			tx.Amount, tx.AccountID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			http.Error(w, "unknown account_id: "+tx.AccountID, http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	http.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, balance FROM accounts ORDER BY id`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var accounts []account
		for rows.Next() {
			var a account
			if err := rows.Scan(&a.ID, &a.Balance); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			accounts = append(accounts, a)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(accounts)
	})

	addr := ":8080"
	log.Printf("daily-balance-sut listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// mustConnect は DB へ接続し、起動直後に postgres が未 ready でも
// 数十秒はリトライする (compose の起動順ゆらぎ対策)。
func mustConnect(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return db
		}
		log.Printf("waiting for db (%d/30): %v", i+1, err)
		time.Sleep(time.Second)
	}
	log.Fatalf("connect db: %v", err)
	return nil
}
