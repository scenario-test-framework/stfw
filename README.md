# stfw

scenario test framework

## ⚠️ v1.0 リアーキテクティング進行中

- 旧実装（Bash + digdag, v0.2 系）は **タグ [`v0.2.0`](../../tree/v0.2.0) で凍結**しました
- 現在 Go 単一バイナリへの全面再実装を進めています（digdag 廃止・逐次実行エンジン内包・Docker/compose 配布）
- as-is の要求・要件は `docs/usdm/latest/` / `docs/rdra/latest/` / `docs/harvest/latest/` に抽出済みです
- 旧実装の仕様参照は `git show v0.2.0:src/...` を使ってください

### v1.0 で維持する互換境界

1. ディレクトリ規約: `scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/`
2. プロセスプラグイン実行契約: 環境変数 + リターンコード（0/3/6, 任意言語）
3. webhook payload JSON スキーマ
