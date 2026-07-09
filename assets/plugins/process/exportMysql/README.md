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
