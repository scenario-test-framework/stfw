# exportMysql プラグイン

MySQL のテーブルをヘッダー付き CSV でエクスポートする Collect フェーズの組込みプラグイン。

## 前提

- 必要コマンド: `mysql` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止):
  - ホスト: inventory グループ (`host_group`)。複数解決時は先頭ホストを使用し警告する
  - パスワード: secret (`{host}-{user}`)。復号値は Masker がログから自動マスクする

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 DB ホストを解決する inventory グループ名 |
| `database` | ○ | 対象データベース名 |
| `user` | ○ | 接続ユーザー |
| `tables[]` | ○ | エクスポート対象テーブル名のリスト |
| `port` | - | MySQL ポート (既定 3306) |
| `charset` | - | 接続の文字コード (既定 `binary` = 変換なし)。下の「文字コード」を参照 |

> ⚠️ `tables` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

```yaml
stfw:
  process:
    exportMysql:
      host_group: db
      database: appdb
      user: appuser
      tables:
        - orders
        - order_items
```

## 動作

`mysql --batch` で SELECT した結果 (タブ区切り) を内部ヘルパ
`stfw plugin mysql-tsv-to-csv` で RFC4180 CSV へ変換して出力する。

## 文字コード

CSV は UTF-8 を前提にする。接続の文字コードは設定 `charset` で選ぶ (exportMysql / importMysql / clearMysql 共通)。

| `charset` | 動き | 向いている DB |
|---|---|---|
| `binary` (既定) | 変換しない。格納されているバイト列をそのまま CSV と受け渡す | utf8mb4 列。latin1 列に UTF-8 のバイト列を入れている既存の DB |
| `utf8mb4` | サーバーが列の文字コードと UTF-8 を変換する | latin1・Shift_JIS など、その文字コードで正しく格納している列 |

- `binary` で UTF-8 以外の列を export すると、その文字コードのまま出力される (CSV が UTF-8 にならない)。
- `utf8mb4` で latin1 列に UTF-8 のバイト列を入れている DB を扱うと、import では日本語が `?` になり、export では二重に変換されて化ける。
- v1.6.12 までは指定しておらず、mysql クライアントの既定 (stfw のイメージでは 1 バイト文字コード) で話していた。
  そのため utf8mb4 列の日本語が export で `?` に化けていた (#52)。

## 出力

```
{process}/evidence/{database}/{table}.csv    # 1 行目ヘッダー (カラム名)
```

NULL は `\N` で表現する (空文字と区別)。importMysql とラウンドトリップ可能。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全テーブルのエクスポート成功 |
| 6 | いずれかのテーブルで失敗 (設定不備・接続失敗を含む) |

## 既知の制約

- `mysql --batch` は SQL の NULL を文字列 `NULL` として出力するため、
  **SQL NULL と文字列 "NULL" は区別できない** (両者を `\N` として扱う)
