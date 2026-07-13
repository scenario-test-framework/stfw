# stfw 参照ドキュメント URL 一覧

フェッチできない環境では、この一覧をそのままユーザーに提示する。
`{ref}` は インストール済み stfw のバージョンタグ (例 `v1.3.0`)。`-dev` 版や不明なら `master`。

raw ベース URL: `https://raw.githubusercontent.com/scenario-test-framework/stfw/{ref}/`
ブラウザ用 URL: `https://github.com/scenario-test-framework/stfw/blob/{ref}/`

## 必読

| 資料 | パス |
|---|---|
| シナリオ作成ガイド (4 フェーズ・実例・spec/scaffold) | `docs/GUIDE.md` |
| 実装契約仕様書 (ディレクトリ規約 §3 / プラグイン契約 §4 / 設定 §8 / secret・inventory §9 / spec スキーマ §12) | `docs/AS-BUILT.md` |

## プラグイン README (使うものだけ)

パス: `assets/plugins/process/{type}/README.md`

| フェーズ | type |
|---|---|
| Arrange | `clearMysql` / `clearPostgres` / `clearRedis` / `importMysql` / `importPostgres` / `importRedis` / `scpPut` |
| Act | `invokeRest` / `invokeWeb` / `sshExec` |
| Collect | `collectLog` / `collectFile` / `exportMysql` / `exportPostgres` / `exportRedis` |
| Assert | `compare` |
| 汎用 | `scripts` |

## 実行可能サンプル (examples/daily-balance)

個別フェッチよりも shallow clone を推奨 (ツリー全体が実例のため):

```bash
git clone --depth 1 --branch {ref} https://github.com/scenario-test-framework/stfw
```

ブラウザで見る場合の起点:

| 内容 | パス |
|---|---|
| example の解説 (README) | `examples/daily-balance/README.md` |
| シナリオツリー実物 | `examples/daily-balance/stfw/scenario/daily-balance/` |
| inventory 実物 | `examples/daily-balance/stfw/config/inventory/local.yml` |
| プロジェクト共通のプラグイン設定 | `examples/daily-balance/stfw/config/plugins/process/` |
| 比較レイアウト実物 | `examples/daily-balance/stfw/config/plugins/process/compare/compare_layout/` |
| カスタムプラグイン実例 (組込みへの委譲) | `examples/daily-balance/stfw/plugins/process/updateBizdate/` |
| カスタムプラグイン実例 (共通データ投入) | `examples/daily-balance/stfw/plugins/process/importMasterData/` |

## compare-files (比較レイアウト)

| 内容 | URL |
|---|---|
| 比較レイアウトリファレンス | `https://github.com/scenario-test-framework/compare-files/blob/master/docs/compare_layout.md` |
| compare-layout スキル (生成はこれに委譲) | `https://github.com/scenario-test-framework/compare-files/tree/master/.claude/skills/compare-layout` |
| インストール | `npx skills add scenario-test-framework/compare-files --skill compare-layout -a claude-code` |
