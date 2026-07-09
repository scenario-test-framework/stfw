# importPostgres プラグイン

ヘッダー付き CSV を PostgreSQL のテーブルへインポートする Arrange フェーズの組込みプラグイン。

## 前提

- 必要コマンド: `psql` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止)。
  詳細は exportPostgres の README を参照 (同じ接続モデル)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 DB ホストを解決する inventory グループ名 |
| `database` | ○ | 対象データベース名 |
| `user` | ○ | 接続ユーザー |
| `tables[]` | ○ | インポート対象テーブル名のリスト |
| `port` | - | PostgreSQL ポート (既定 5432) |

> ⚠️ `tables` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

## 入力 (テスト作者が用意・git 管理)

```
{process}/data/{database}/{table}.csv    # exportPostgres の出力形式 (ヘッダー付き・NULL は \N)
```

exportPostgres でエクスポートした CSV をそのまま投入できる (ラウンドトリップ)。

## 動作

psql の COPY (クライアント側ファイル入力) で CSV を投入する。
既存データは削除しない (初期化が必要な場合は clearPostgres を先行プロセスに置く)。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全テーブルのインポート成功 |
| 6 | いずれかのテーブルで失敗 (入力 CSV 不在・接続失敗を含む) |
