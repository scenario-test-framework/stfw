# exportRedis プラグイン

Redis のキーをヘッダー付き CSV (`key,type,ttl,value`) でエクスポートする Collect フェーズの組込みプラグイン。

## 前提

- 必要コマンド: `redis-cli` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止):
  - ホスト: inventory グループ (`host_group`)。複数解決時は先頭ホストを使用し警告する
  - パスワード: secret (`{host}-{user}`)。`REDISCLI_AUTH` 環境変数で渡すため argv に露出しない

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 Redis ホストを解決する inventory グループ名 |
| `user` | ○ | 認証ユーザー (secret の解決キー。ACL 未使用/requirepass のみの Redis 6+ では `default` を指定) |
| `key_patterns[].name` | ○ | 出力ファイル名 (`evidence/{host}/{name}.csv`) |
| `key_patterns[].match` | ○ | SCAN の glob パターン |
| `port` | - | Redis ポート (既定 6379) |
| `db` | - | DB 番号 0-15 (既定 0) |

> ⚠️ `key_patterns` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

```yaml
stfw:
  process:
    exportRedis:
      host_group: cache
      user: appuser
      key_patterns:
        - name: session
          match: "session:*"
```

## 動作

match ごとに SCAN でキーを列挙し、各キーの type / ttl / 値を取得して CSV 化する。
value の正規化 (compare での比較安定性のため):

| type | value 列の表現 |
|---|---|
| string | 生値 |
| list | 順序保持の JSON 配列 |
| set | ソート済み JSON 配列 |
| hash | field 昇順の JSON オブジェクト |
| zset | `[[member,score],...]` member 昇順の JSON |

## 出力

```
{process}/evidence/{host}/{name}.csv    # ヘッダー key,type,ttl,value
```

パターン内の全キー取得が成功したときのみファイルを公開する (途中失敗で不完全な CSV を残さない)。
importRedis とラウンドトリップ可能。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全パターンのエクスポート成功 |
| 6 | いずれかで失敗 |

## 既知の制約

- 値は UTF-8 テキストのみ対象 (バイナリ値は非対応)
- 改行を含むキーは非対応 (SCAN 出力の行区切り前提)
- SCAN 中に変更されるキーの整合は保証しない (テスト用途では静止点で実行する)
