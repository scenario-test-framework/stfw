<p align="center">
  <img src="docs/assets/ogp.png" alt="stfw — scenario test framework" width="760">
</p>

<p align="center">
  <a href="https://github.com/scenario-test-framework/stfw/actions/workflows/ci.yml"><img src="https://github.com/scenario-test-framework/stfw/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/scenario-test-framework/stfw/releases"><img src="https://img.shields.io/github/v/release/scenario-test-framework/stfw" alt="Release"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/scenario-test-framework/stfw" alt="Go"></a>
  <a href="https://github.com/scenario-test-framework/stfw/pkgs/container/stfw"><img src="https://img.shields.io/badge/ghcr.io-stfw-2496ED?logo=docker&logoColor=white" alt="Container"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/scenario-test-framework/stfw" alt="License"></a>
</p>

<p align="center">
  <a href="README.md">English</a> | <b>日本語</b>
</p>

# stfw

scenario test framework — 業務日付をまたぐシナリオテストをディレクトリ規約で記述し、単一バイナリで自動実行する CLI。

- **単一バイナリ**: Go 製。実行エンジン内包（digdag 等の外部ワークフローエンジン・JVM は不要）
- **規約ベース**: `scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/` の階層にスクリプトを置くだけ
- **順序保証・エラー時停止**: ファイル名昇順で逐次実行し、エラー以降のステップは `Blocked` として記録
- **可視化**: JSONL 実行ジャーナル + `stfw status` + 静的 HTML レポート
- **オブザーバビリティ**: run / scenario / bizdate / process / step の実行状況を OTLP トレースとしてエクスポート (Jaeger / Grafana Tempo / Datadog 等でそのまま可視化)
- **組込みプラグイン**: Arrange → Act → Collect → Assert のシナリオテストを部品の組み合わせで記述 (RDBMS / Redis / ssh / scp / k6 / ファイル突合)
- **ハウスキープ**: `stfw run` の開始時に保存期間 (`stfw.housekeep.retention_days`) を過ぎた実行結果を自動削除

> ℹ️ 内部ドキュメント（`docs/`）とコード内コメントは日本語です。

## インストール

### バイナリ (Linux / macOS / Windows)

- Linux / macOS

  ```sh
  curl -fsSL https://raw.githubusercontent.com/scenario-test-framework/stfw/master/install.sh | bash
  stfw --version
  ```

  `install.sh` は OS / arch を自動判定し、デフォルトで最新リリースをインストールします。
  バージョン固定やインストール先の変更も可能です。

  ```sh
  curl -fsSL https://raw.githubusercontent.com/scenario-test-framework/stfw/master/install.sh | \
    STFW_VERSION=X.Y.Z STFW_BINDIR=$HOME/.local/bin bash
  ```

  アンインストール:

  ```sh
  curl -fsSL https://raw.githubusercontent.com/scenario-test-framework/stfw/master/uninstall.sh | bash
  ```

  インストール先を変更していた場合は同じディレクトリを指定します。

  ```sh
  curl -fsSL https://raw.githubusercontent.com/scenario-test-framework/stfw/master/uninstall.sh | \
    STFW_BINDIR=$HOME/.local/bin bash
  ```

- Windows (PowerShell)

  ```ps1
  & ([scriptblock]::Create((irm https://raw.githubusercontent.com/scenario-test-framework/stfw/master/install.ps1)))
  stfw --version
  ```

  `install.ps1` はアーキテクチャを自動判定し、デフォルトで最新リリースを `$HOME\bin` に配置して
  ユーザー PATH に追加します。バージョン固定やインストール先の変更も可能です。

  ```ps1
  & ([scriptblock]::Create((irm https://raw.githubusercontent.com/scenario-test-framework/stfw/master/install.ps1))) `
    -Version X.Y.Z `
    -BinDir "$HOME\bin"
  ```

  アンインストール:

  ```ps1
  & ([scriptblock]::Create((irm https://raw.githubusercontent.com/scenario-test-framework/stfw/master/uninstall.ps1)))
  ```

  インストール先を変更していた場合は同じディレクトリを指定します。

  ```ps1
  & ([scriptblock]::Create((irm https://raw.githubusercontent.com/scenario-test-framework/stfw/master/uninstall.ps1))) `
    -BinDir "$HOME\bin"
  ```

### Docker

```console
docker pull ghcr.io/scenario-test-framework/stfw:latest
docker run --rm -v "$PWD":/work ghcr.io/scenario-test-framework/stfw:latest --version
```

組込みプラグイン (RDBMS / Redis / ssh 系 / invokeWeb) を使う場合は、全ランタイム同梱版の
**`stfw:full`** を使います (mysql / psql / redis-cli / sshpass / Chromium を同梱):

```console
docker pull ghcr.io/scenario-test-framework/stfw:full
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

`compose.yaml` は stfw + nginx + reports-init (volume 所有権初期化の one-shot) の構成です。実行レポートを共有 volume 経由で nginx が配信します。

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

プロセスタイプは次の契約で拡張できます:

- 入力: 環境変数（`stfw_*` = 設定のフラット化 + `STFW_PROJ_DIR` などの実行コンテキスト）
- 出力: リターンコード（`0` = Success / `3` = Warn / `6` = Error）
- 実装言語は任意（実行可能ファイルであればよい）

### 組込みプロセスプラグイン

Arrange → Act → Collect → Assert のシナリオテストを部品の組み合わせで記述できます。
接続先は inventory グループ・パスワードは secret から解決し、設定への直書きを禁止します。

| フェーズ | プラグイン | 説明 |
|---|---|---|
| 任意 | `scripts` | 任意スクリプトの昇順逐次実行 (Go ネイティブ) |
| Arrange | `importMysql` / `importPostgres` / `importRedis` | CSV からデータストアへ投入 |
| Arrange | `clearMysql` / `clearPostgres` / `clearRedis` | データストアの初期化 |
| Arrange | `scpPut` | ローカルファイルをリモートへ原子的に配置 (scp + atomic rename) |
| Act | `invokeRest` / `invokeWeb` | grafana k6 による API 取引入力・ブラウザ操作 |
| Act | `sshExec` | リモートスクリプトの一括実行 (ssh) |
| Collect | `collectFile` / `collectLog` | リモートからのエビデンス収集 (時刻フィルタつき) |
| Collect | `exportMysql` / `exportPostgres` / `exportRedis` | データストアの CSV エクスポート |
| Assert | `compare` | 期待値とエビデンスのディレクトリ突合 (compare-files) |

実プロジェクトに近い**動く例**と、組み方の**通し解説**は次を参照してください:

- [examples/daily-balance](examples/daily-balance/) — 業務日付をまたぐ日次残高バッチの実行可能サンプル。postgres + トイ REST API を同梱し `./run.sh` で end-to-end 実行できます（Arrange→Act→Collect→Assert を組込みプラグインだけで構成）
- [docs/GUIDE.md](docs/GUIDE.md) — シナリオ作成ガイド（4 フェーズの組み方・接続情報・エビデンス規約の通し解説）

詳細な契約・設定は [docs/AS-BUILT.md](docs/AS-BUILT.md) §4 を参照してください。

## 主なコマンド

| コマンド | 説明 |
|---|---|
| `stfw init` | プロジェクト初期化 |
| `stfw new scenario/bizdate/process` | 階層の scaffold 生成 |
| `stfw scenario reverse <name> [-o dir]` | シナリオから spec (`.yml`) + doc (`.md`) をセット生成 (tree → spec + doc、既定出力先 `docs/`) |
| `stfw scenario scaffold <spec.yml> [--sync]` | spec からシナリオ骨格を生成 (spec → tree、往復の入口)。既存シナリオは `--sync` で差分同期 (追加/維持/削除) |
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
開発規約は [CLAUDE.md](CLAUDE.md) を参照してください。

## License

[Apache License 2.0](LICENSE)。
