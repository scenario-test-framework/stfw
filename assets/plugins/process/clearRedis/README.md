# clearRedis プラグイン

Redis のキー (パターン一致) を削除する Arrange フェーズ (初期化) の組込みプラグイン。

## 前提

- 必要コマンド: `redis-cli` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止)。
  詳細は exportRedis の README を参照 (同じ接続モデル)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 Redis ホストを解決する inventory グループ名 |
| `user` | ○ | 認証ユーザー (secret の解決キー。requirepass のみの Redis 6+ では `default`) |
| `key_patterns[].match` | ○ | 削除対象キーの SCAN glob パターン |
| `port` | - | Redis ポート (既定 6379) |
| `db` | - | DB 番号 0-15 (既定 0) |

> ⚠️ `key_patterns` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

```yaml
stfw:
  process:
    clearRedis:
      host_group: cache
      user: appuser
      key_patterns:
        - match: "session:*"
```

## 動作

match ごとに SCAN で列挙したキーを DEL する。データ投入 (importRedis) の先行プロセスとして使う。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全パターンの削除成功 |
| 6 | いずれかで失敗 |
