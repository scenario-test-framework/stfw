# importRedis プラグイン

exportRedis 形式の CSV (`key,type,ttl,value`) を Redis へインポートする Arrange フェーズの組込みプラグイン。

## 前提

- 必要コマンド: `redis-cli` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止)。
  詳細は exportRedis の README を参照 (同じ接続モデル)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 Redis ホストを解決する inventory グループ名 |
| `user` | ○ | 認証ユーザー (secret の解決キー。requirepass のみの Redis 6+ では `default`) |
| `key_patterns[].name` | ○ | 入力ファイル名 (`data/{host}/{name}.csv`) |
| `port` | - | Redis ポート (既定 6379) |
| `db` | - | DB 番号 0-15 (既定 0) |

> ⚠️ `key_patterns` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

## 入力 (テスト作者が用意・git 管理)

```
{process}/data/{host}/{name}.csv    # exportRedis の出力形式 (ヘッダー key,type,ttl,value)
```

exportRedis でエクスポートした CSV をそのまま投入できる (ラウンドトリップ)。

## 動作

内部ヘルパ `stfw plugin redis-decode` で CSV を redis-cli コマンド列へ変換し、
redis-cli の標準入力へ渡して投入する。各キーは **DEL 後に再作成**し、
type に応じて SET / RPUSH / SADD / HSET / ZADD を使う。`ttl > 0` は EXPIRE を設定する。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全ファイルのインポート成功 |
| 6 | いずれかで失敗 (入力 CSV 不在・接続失敗を含む) |

## 既知の制約

- 値は UTF-8 テキストのみ対象 (バイナリ値は非対応)
