# invokeWeb プラグイン

画面への取引入力と画面操作検証を [grafana k6](https://github.com/grafana/k6) の**ブラウザモード**で実行する Act フェーズの組込みプラグイン。

## 前提

- プロビジョニング: `stfw plugin install invokeWeb` (通常は `stfw init` が自動実行) が
  実行ホストの os_arch 版 `k6` をダウンロードし `.stfw/cache/plugins/invokeWeb/` へキャッシュする
- **実行には Chromium が必要**。コンテナ利用時は全ランタイム同梱の
  `ghcr.io/scenario-test-framework/stfw:full` イメージを使う
  (`K6_BROWSER_EXECUTABLE_PATH` / `K6_BROWSER_ARGS=no-sandbox` 設定済み)。
  ローカル実行では Chrome / Chromium を導入し、必要なら
  `K6_BROWSER_EXECUTABLE_PATH` で実行パスを指定する
- 接続先を注入する場合は inventory / secret を使う (config への直書きは禁止)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `script` | - | k6 ブラウザモードのテストスクリプトのパス (既定 `script.js`) |
| `host_group` | - | inventory グループ名。先頭ホストを `__ENV.stfw_target_host` へ注入 |
| `user` | - | secret `{host}-{user}` のパスワードを `__ENV.stfw_target_password` へ注入 (`host_group` 必須) |
| `env[]` | - | k6 へ渡す `"KEY=VALUE"` のリスト (`-e` で注入) |
| `k6_version` | - | 取得する k6 のリリースタグ (既定 v2.1.0) |

## 動作

キャッシュ済み k6 バイナリで `k6 run` を実行する。ブラウザは既定でヘッドレス
(`K6_BROWSER_HEADLESS=true`。環境変数で上書き可)。パスワードは argv に載せず
環境変数として渡す (ログは Masker が自動マスク)。

k6 スクリプトは `k6/browser` モジュールを使い、scenario options に
`options: { browser: { type: 'chromium' } }` を指定する。

## 出力

```
{process}/evidence/summary.json    # k6 の end-of-test サマリ (--summary-export)
```

## 終了コード

| コード | 意味 |
|---|---|
| 0 | k6 成功 |
| 6 | k6 が非 0 (閾値失敗 99 / 各種エラー 100+ / 外部中断 105 を含む)。Act の失敗はステップ失敗 |

## 補足

- `k6_version` を変更したときは `.stfw/cache/plugins/invokeWeb/` を削除してから再 install する
  (invokeRest とはキャッシュが独立)
