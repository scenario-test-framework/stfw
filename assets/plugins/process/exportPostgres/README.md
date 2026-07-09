# exportPostgres プラグイン

PostgreSQL のテーブルをヘッダー付き CSV でエクスポートする Collect フェーズの組込みプラグイン。

## 前提

- 必要コマンド: `psql` (実行前に PATH ゲートされる)
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
| `port` | - | PostgreSQL ポート (既定 5432) |

> ⚠️ `tables` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

```yaml
stfw:
  process:
    exportPostgres:
      host_group: db
      database: appdb
      user: appuser
      tables:
        - orders
        - order_items
```

## 動作

psql の COPY (クライアント側ファイル出力) で RFC4180 CSV を直接書き出す。

## 出力

```
{process}/evidence/{database}/{table}.csv    # 1 行目ヘッダー (カラム名)
```

NULL は `\N` で表現する (空文字と区別)。importPostgres とラウンドトリップ可能。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全テーブルのエクスポート成功 |
| 6 | いずれかのテーブルで失敗 (設定不備・接続失敗を含む) |
