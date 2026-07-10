# example: daily-balance（日次残高バッチ）

stfw の**組込みプラグインエコシステム**を実プロジェクトに近い形で示す、実行可能なサンプルです。
業務日付（bizdate）をまたいで口座残高の繰越を検証します。

- **Arrange**（準備）→ **Act**（実行）→ **Collect**（収集）→ **Assert**（検証）の 4 フェーズを、
  組込みプラグインだけで組み立てます。
- テスト対象システム（SUT）として、postgres に残高を書き込む小さな REST API を同梱しています
  （`sut/`）。外部インフラ不要で end-to-end に動きます。

> 各手順の意図・ディレクトリ規約・プラグイン契約の解説は [`../../docs/GUIDE.md`](../../docs/GUIDE.md) を参照。
> 本 README は「動かし方」に絞っています。

## 前提

- Docker / Docker Compose

## 実行

```sh
./run.sh
```

`run.sh` は次を順に行います。

1. 依存サービス（postgres + トイ API = SUT、Jaeger = トレース送信先）を起動
2. secret を準備（age 鍵生成 + DB パスワード登録）
3. プラグインの外部バイナリ（k6 / compare-files）を install
4. `stfw run daily-balance` を実行
5. HTML レポート配信（nginx）を起動 → http://localhost:8088
6. 実行トレースを Jaeger UI で閲覧 → http://localhost:16686（service=stfw）

後片付け:

```sh
./run.sh --down
```

> ポートを変えたい場合は `STFW_REPORT_PORT=9000 STFW_JAEGER_PORT=16690 ./run.sh`。

### 実行状況の可視化

- **HTML レポート**（stfw 内蔵）: run の階層ごとの Success/Error を nginx で配信（`http://localhost:8088`）。
- **OTLP トレース**（Jaeger）: compose の `stfw` サービスに `OTEL_EXPORTER_OTLP_ENDPOINT: http://jaeger:4318`
  を設定済み。1 run = 1 トレースとして run→bizdate→process→step のスパンツリーを `http://localhost:16686`
  で閲覧できる。
- **k6 レポート**（invokeRest の evidence）: k6 の end-of-test サマリから自己完結の HTML
  （`_30_act_invokeRest/evidence/report.html`）を生成する。`summary.json` も併置。

## シナリオの流れ

初期残高 `acc-001=1000` / `acc-002=2000` に対し、2 業務日で取引を反映します。

プロセスのグループ名はフェーズ（arrange / act / collect / assert）に揃えてあり、
ディレクトリ名を見るだけで A→A→C→A の流れが読めます。

| bizdate | プロセス | プラグイン | フェーズ | 内容 |
|---|---|---|---|---|
| `_10_20240101` | `_10_arrange` | clearPostgres | Arrange | users / accounts / transactions を truncate |
| | `_15_arrange` | **importMasterData**（カスタム） | Arrange | 口座名義マスタ users を config 内の CSV から投入 |
| | `_20_arrange` | importPostgres | Arrange | 初期残高 CSV を投入 |
| | `_30_act` | invokeRest | Act | API へ取引 POST（acc-001 +500 / acc-002 +300） |
| | `_40_collect` | exportPostgres | Collect | 残高を `evidence/appdb/accounts.csv` へ |
| | `_50_assert` | compare | Assert | 期待残高（1500 / 2300）と突合 |
| `_20_20240102` | `_10_act` | invokeRest | Act | 前日残高に対して取引（acc-001 -200 / acc-002 +100） |
| | `_20_collect` | exportPostgres | Collect | 残高を収集 |
| | `_30_assert` | compare | Assert | **繰越**した累積残高（1300 / 2400）と突合 |

Day2 は reset / seed を行いません。**Day1 の残高を引き継ぐ**ことで「業務日付をまたぐ」意味を示します。

## カスタムプラグイン（importMasterData）

`_15_arrange` は**カスタムプロセスプラグイン** `importMasterData`
（[`stfw/plugins/process/importMasterData/`](stfw/plugins/process/importMasterData/)）の実装例です。
**複数シナリオで共有するテスト共通のマスタ/参照データ**（ここでは口座名義 `users`）を投入します。

共通データは特定シナリオに属さず、プロジェクト共通の場所に集約します。

```
stfw/config/plugins/process/importMasterData/data/appdb/users.csv   # ← 全シナリオで共有
```

各シナリオは `importMasterData` プロセスの config で `tables: [users]` と宣言するだけで、
この共通データを取り込めます。

観点は **「カスタムプラグインは組込みプラグインを部品として再利用できる」** こと。
DB 投入は再実装せず、次の 2 段で実現しています。

1. **ファイル操作**: 組込み `importPostgres` が読む（プロセス配下の）`data/appdb/users.csv` から、
   共通データ `config/plugins/process/importMasterData/data/appdb/users.csv` へ **symlink を張る**
   （実体はコピーせず共通データを唯一の正とする）。プロセス配下の `data/` は gitignore 済み。
2. **委譲**: 委譲先の組込み `importPostgres` を先頭で確保（無ければ早期エラー）し、その `execute` を
   接続系の env を訳して呼び出す。

`plugins/process/{type}/` に置くだけで組込みより優先して解決されます（プラグイン契約の詳細は
プラグインの [README](stfw/plugins/process/importMasterData/README.md) と
[`../../docs/GUIDE.md`](../../docs/GUIDE.md) を参照）。

## わざと失敗させてみる

Assert が本当に効いていることを確かめるには、期待値を書き換えて再実行します。

```sh
# 例: Day1 の期待残高を 1500 → 9999 に変える
vi stfw/scenario/daily-balance/_10_20240101/_50_assert_compare/expect/_40_collect_exportPostgres/appdb/accounts.csv
docker compose run --rm stfw run daily-balance   # compare が差分を検出し Error 終了
```

## ドキュメント / spec（ラウンドトリップ）

シナリオのツリー（=正）から、機械可読な **spec** と人が読む**ドキュメント**をまとめて生成できます
（リバース生成）。逆に spec からツリーを再生成することもできます。

| コマンド | 生成物 | 用途 |
|---|---|---|
| `stfw scenario reverse daily-balance` | [`stfw/docs/daily-balance.yml`](stfw/docs/daily-balance.yml) + [`.md`](stfw/docs/daily-balance.md) | tree → spec (`.yml`) + doc (`.md`) をセット生成 |
| `stfw scenario scaffold <spec.yml> [--sync]` | ディレクトリ骨格 | spec からツリーを生成／差分同期（往復の入口） |

- `reverse` は各階層の `metadata.yml`（`description` / `requirement_specifications`）と
  `config/config.yml` を読み取ります。本例では assert プロセスに `REQ-01` / `REQ-02` を紐づけ、
  doc（`.md`）の「要求トレーサビリティ」表に「どの要求をどの process が検証するか」が出力されます。
- spec（`.yml`）はツリーと可逆で、`reverse → scaffold → reverse` は**完全一致**（骨格：seq /
  bizdate / group / type / description / requirement_specifications / config.yml のサブツリー）。
  data CSV・script・expect などの葉は対象外です。

同梱の `stfw/docs/daily-balance.{yml,md}` は上記コマンドで生成した実出力です。手元で再生成するには
（stfw サービスの作業ディレクトリ `/work` が `stfw/` にマウントされます。既定出力先は `docs/`）:

```sh
docker compose run --rm stfw scenario reverse daily-balance
```

> 詳細は [`../../docs/GUIDE.md` §8](../../docs/GUIDE.md) を参照。

## ディレクトリ

```
examples/daily-balance/
├── run.sh              # 一発実行スクリプト
├── compose.yaml        # postgres + api(SUT) + jaeger(トレース) + stfw(:full) + nginx
├── sut/                # テスト対象システム (トイ REST API + スキーマ)
│   ├── main.go
│   ├── schema.sql      # users(マスタ) / accounts(残高) / transactions
│   └── Dockerfile
└── stfw/               # stfw プロジェクト (stfw init 相当 + シナリオ)
    ├── stfw.yml                                             # stfw.db.* = DB 接続先の単一ソース
    ├── config/
    │   ├── inventory/local.yml
    │   └── plugins/process/                                 # プロセスプラグインの共通設定
    │       ├── {clear,import,export}Postgres/config.yml     #   接続系 (${stfw_db_*} 参照)
    │       └── importMasterData/
    │           ├── config.yml                               #   DB 接続系を共通化
    │           └── data/appdb/users.csv                     #   シナリオ共通のマスタデータ
    ├── plugins/            # カスタムプラグイン
    │   └── process/importMasterData/   # 共通データ → 組込み importPostgres へ委譲
    ├── docs/               # リバース生成物 (spec + doc の実出力例)
    │   ├── daily-balance.yml
    │   └── daily-balance.md
    └── scenario/daily-balance/
        ├── _10_20240101/   # Day1: clear→importMasterData→import→act→collect→assert
        └── _20_20240102/   # Day2: act→collect→assert (繰越)
```

接続情報は config に直書きせず、inventory（ホスト解決）と secret（パスワード）から解決します。
`run.sh` が生成する `stfw/config/encrypt/`・`stfw/config/passwd/`（デモ用鍵・クレデンシャル）は
git 管理外です。

## DB 接続系の共通化

`clearPostgres` / `importPostgres` / `exportPostgres` / `importMasterData` は、どれも同じ DB
（`database` / `user`）へ接続します。各シナリオのプロセス config にこれを書くと、ユーザーを
変えたときに全プロセスを直す必要があり、直し忘れやすくなります。そこで **接続系はプロジェクト
共通の上書き設定に一本化** し、各プロセスの `config/config.yml` は `tables` だけを書きます。

```
stfw.yml                                    ← database / user を単一ソースで宣言 (stfw.db.*)
config/plugins/process/{type}/config.yml   ← host_group / port と ${stfw_db_*} 参照を共通定義
scenario/.../{process}/config/config.yml   ← tables のみ
```

- 設定の上書きチェーン（§8.1）: プラグイン既定 → `config/plugins/process/{type}/` →
  各プロセス `config/config.yml`。共通値は真ん中の層に置きます。
- `database` / `user` の値そのものは **`stfw.yml` の `stfw.db.*` が単一ソース**です。共通設定は
  `${stfw_db_database}` / `${stfw_db_user}` で参照します。`stfw run` 開始時に stfw.yml の値が
  環境へ export され、config チェーンの `${...}` 展開で解決されます（AS-BUILT §8.2）。

```yaml
# stfw/stfw.yml — 接続先 DB の identity をここ 1 か所で管理 (探しやすい)
stfw:
  db:
    database: appdb
    user: appuser
```

> SUT 側（`compose.yaml` の postgres/api）は「テスト対象そのものの DB 定義」なので compose に
> 置きます。`stfw.yml` の `stfw.db.*` はそこへ接続する側の宣言で、値を一致させます。
> 接続先ホスト・パスワードは `stfw.yml` に書かず、inventory + secret で解決します（禁止契約）。
