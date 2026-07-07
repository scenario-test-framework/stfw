# stfw

scenario test framework — 業務日付をまたぐシナリオテストをディレクトリ規約で記述し、単一バイナリで自動実行する CLI。

- **単一バイナリ**: Go 製。実行エンジン内包（digdag 等の外部ワークフローエンジン・JVM は不要）
- **規約ベース**: `scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/` の階層にスクリプトを置くだけ
- **順序保証・エラー時停止**: ファイル名昇順で逐次実行し、エラー以降のステップは `Blocked` として記録
- **可視化**: JSONL 実行ジャーナル + `stfw status` + 静的 HTML レポート
- **通知**: 各階層の開始・終了を webhook (JSON POST) で外部へ通知

> ⚠️ v0.2 系（Bash + digdag 実装）はタグ [`v0.2.0`](../../tree/v0.2.0) で凍結しました。
> 移行手順と非互換事項は [docs/MIGRATION.md](docs/MIGRATION.md) を参照してください。

## インストール

### バイナリ (Linux / macOS / Windows)

[GitHub Releases](../../releases) からアーカイブを取得し、`stfw` を PATH に置きます。

### Docker

```console
docker pull ghcr.io/scenario-test-framework/stfw:latest
docker run --rm -v "$PWD":/work ghcr.io/scenario-test-framework/stfw:latest --version
```

## Quick Start

```console
$ mkdir myproject && cd myproject
$ stfw init                        # プロジェクト初期化 (sample シナリオ付き)
$ stfw run sample                  # シナリオ実行
$ stfw status                      # 実行結果ツリーの表示
$ stfw report                      # HTML レポート再生成 (.stfw/reports/)
```

シナリオの追加:

```console
$ stfw new scenario release_test               # シナリオ
$ cd scenario/release_test
$ stfw new bizdate 10 20260701                 # 業務日付 (連番 + YYYYMMDD)
$ cd _10_20260701
$ stfw new process 10 web scripts              # プロセス (連番 + グループ + タイプ)
$ stfw validate release_test                   # 規約の静的検証
```

## Docker Compose (HTML レポート配信つき)

`compose.yaml` は stfw + nginx の 2 サービス構成です。実行レポートを共有 volume 経由で nginx が配信します。

```console
$ docker compose up -d nginx                 # レポート配信を起動
$ docker compose run --rm stfw init          # プロジェクト初期化
$ docker compose run --rm stfw run sample    # シナリオ実行
$ open http://localhost:8080                 # ブラウザでレポートを閲覧
```

## ディレクトリ規約

```
myproject/
├── stfw.yml                     # プロジェクト設定 (デフォルト設定を上書き)
├── config/
│   ├── inventory/staging.yml    # テスト対象ホストのグループ定義
│   ├── encrypt/                 # 暗号化キー (stfw secret keygen)
│   └── passwd/                  # 暗号化済み資格情報 (stfw secret set)
├── plugins/                     # 階層フック・独自プロセスプラグイン
│   └── {run,scenario,bizdate,process}/_common/{setup,teardown}/
└── scenario/
    └── {scenario}/              # シナリオ
        └── _{seq}_{bizdate}/    # 業務日付 (昇順に実行)
            └── _{seq}_{group}_{type}/   # プロセス (昇順に実行)
                └── scripts/     # ステップ (昇順に逐次実行, エラーで停止)
```

## プロセスプラグイン契約

プロセスタイプ（組込みは `scripts`）は次の契約で拡張できます:

- 入力: 環境変数（`stfw_*` = 設定のフラット化 + `STFW_PROJ_DIR` などの実行コンテキスト）
- 出力: リターンコード（`0` = Success / `3` = Warn / `6` = Error）
- 実装言語は任意（実行可能ファイルであればよい）

## 主なコマンド

| コマンド | 説明 |
|---|---|
| `stfw init` | プロジェクト初期化 |
| `stfw new scenario/bizdate/process` | 階層の scaffold 生成 |
| `stfw validate [scenario...]` | ディレクトリ規約・プラグイン解決の静的検証 |
| `stfw run [--dry-run] <scenario...>` | シナリオの一括自動実行 |
| `stfw status [run_id]` | 実行結果ツリーの表示 |
| `stfw report [run_id] [--out dir]` | HTML レポート再生成 |
| `stfw inventory list/exists` | ホストグループの参照 |
| `stfw secret keygen/set/show/migrate` | 資格情報の暗号化管理 (age) |
| `stfw ssh trust <host\|group>` | SSH サーバキーの known_hosts 登録 |
| `stfw plugin list/install` | プロセスプラグイン管理 |

## 開発

```console
$ go build ./...
$ go test ./...        # 単体 + testscript 受け入れテスト
```

要求・要件の抽出資産（USDM / RDRA）は `docs/usdm/` / `docs/rdra/` / `docs/harvest/` にあります。

## License

[LICENSE](LICENSE) を参照。
