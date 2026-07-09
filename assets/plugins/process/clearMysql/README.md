# clearMysql プラグイン

MySQL のテーブルを truncate する Arrange フェーズ (初期化) の組込みプラグイン。

## 前提

- 必要コマンド: `mysql` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止)。
  詳細は exportMysql の README を参照 (同じ接続モデル)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 DB ホストを解決する inventory グループ名 |
| `database` | ○ | 対象データベース名 |
| `user` | ○ | 接続ユーザー |
| `tables[]` | ○ | truncate 対象テーブル名のリスト |
| `port` | - | MySQL ポート (既定 3306) |

> ⚠️ `tables` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

```yaml
stfw:
  process:
    clearMysql:
      host_group: db
      database: appdb
      user: appuser
      tables:
        - orders
```

## 動作

各テーブルを `TRUNCATE TABLE` する。データ投入 (importMysql) の先行プロセスとして使う。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全テーブルの truncate 成功 |
| 6 | いずれかのテーブルで失敗 |
