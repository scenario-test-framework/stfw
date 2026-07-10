# invokeRest プラグイン

API への取引入力とレスポンス検証を [grafana k6](https://github.com/grafana/k6) で実行する Act フェーズの組込みプラグイン。

## 前提

- プロビジョニング: `stfw plugin install invokeRest` (通常は `stfw init` が自動実行) が
  実行ホストの os_arch 版 `k6` をダウンロードし `.stfw/cache/plugins/invokeRest/` へ
  キャッシュする (install 時に `curl` と `tar` (linux) / `unzip` (macOS) が必要)
- 接続先を注入する場合は inventory / secret を使う (config への直書きは禁止)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `script` | - | k6 テストスクリプトのパス (プロセスディレクトリ基準。既定 `script.js`) |
| `host_group` | - | inventory グループ名。先頭ホストを `__ENV.stfw_target_host` へ注入 |
| `user` | - | secret `{host}-{user}` のパスワードを `__ENV.stfw_target_password` へ注入 (`host_group` 必須) |
| `env[]` | - | k6 へ渡す `"KEY=VALUE"` のリスト (`-e` で注入) |
| `k6_version` | - | 取得する k6 のリリースタグ (既定 v2.1.0) |

```yaml
stfw:
  process:
    invokeRest:
      script: script.js
      host_group: api
      user: appuser
      env:
        - "VUS=5"
        - "DURATION=10s"
```

k6 スクリプト側からは `__ENV.stfw_target_host` / `__ENV.stfw_target_password` で参照する。

## 動作

キャッシュ済み k6 バイナリで `k6 run` を実行する。パスワードは argv に載せず
環境変数として渡す (k6 が `__ENV` へ取り込む。ログは Masker が自動マスク)。

## 出力

```
{process}/evidence/summary.json    # k6 の end-of-test サマリ (--summary-export)
{process}/evidence/report.html     # summary.json から生成する自己完結の HTML レポート
```

## 終了コード

| コード | 意味 |
|---|---|
| 0 | k6 成功 |
| 6 | k6 が非 0 (閾値失敗 99 / 各種エラー 100+ / 外部中断 105 を含む)。Act の失敗はステップ失敗 |

## 補足

- `k6_version` を変更したときは `.stfw/cache/plugins/invokeRest/` を削除してから再 install する
