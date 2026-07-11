# example: daily-balance（日次残高バッチ）

stfw の**組込みプラグインエコシステム**を実プロジェクトに近い形で示す、実行可能なサンプルです。
業務日付（bizdate）をまたいで口座残高の繰越を検証します。

- **Arrange**（準備）→ **Act**（実行）→ **Collect**（収集）→ **Assert**（検証）の 4 フェーズを、
  組込みプラグインと、それらを部品として再利用するカスタムプラグイン
  （importMasterData / updateBizdate）で組み立てます。
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

> ⚠️ `sut/schema.sql` は postgres コンテナの**初期化時にだけ**実行されます。スキーマ変更を含む
> 更新を取り込んだ後 (git pull 等) は、`./run.sh --down` で DB を作り直してから `./run.sh` を
> 実行してください。既存コンテナのままだと新テーブル (biz_calendar 等) が存在せず失敗します。

### 実行状況の可視化

- **HTML レポート**（stfw 内蔵）: run の階層ごとの Success/Error を nginx で配信（`http://localhost:8088`）。
- **OTLP トレース**（Jaeger）: compose の `stfw` サービスに `OTEL_EXPORTER_OTLP_ENDPOINT: http://jaeger:4318`
  を設定済み。1 run = 1 トレースとして run→scenario→bizdate→process→step のスパンツリーを `http://localhost:16686`
  で閲覧できる。
- **k6 レポート**（invokeRest の evidence）: HTML レポート `_30_act_invokeRest/evidence/report.html`
  を出力する。負荷テスト等で十分なデータがあれば k6 の web dashboard レポートを、無ければ
  （本例の単発 Act のように）`summary.json` から生成した自己完結 HTML をフォールバックで出す。

## シナリオの流れ

初期残高 `acc-001=1000` / `acc-002=2000` に対し、2 業務日で取引を反映します。
データ準備（`_010`）と業務日付ごとの実行（`_020` / `_030`）を階層で分離しています。

プロセスのグループ名はフェーズ（arrange / act / collect / assert）に揃えてあり、
ディレクトリ名を見るだけで A→A→C→A の流れが読めます。

| bizdate | プロセス | プラグイン | フェーズ | 内容 |
|---|---|---|---|---|
| `_010_20240101` | `_10_arrange` | clearPostgres | Arrange | users / accounts / transactions を truncate |
| | `_15_arrange` | **importMasterData**（カスタム） | Arrange | 口座名義マスタ users を config 内の CSV から投入 |
| | `_20_arrange` | importPostgres | Arrange | 初期残高 CSV を投入 |
| `_020_20240101` | `_10_arrange` | **updateBizdate**（カスタム） | Arrange | SUT の業務日付（biz_calendar）を 20240101 へ更新 |
| | `_30_act` | invokeRest | Act | API へ取引 POST（acc-001 +500 / acc-002 +300） |
| | `_40_collect` | exportPostgres | Collect | 残高と取引履歴を `evidence/appdb/{accounts,transactions}.csv` へ |
| | `_50_assert` | compare | Assert | 期待残高（1500 / 2300）と取引の業務日付（20240101）を突合 |
| `_030_20240102` | `_10_arrange` | **updateBizdate**（カスタム） | Arrange | SUT の業務日付を 20240102 へ更新（残高は引き継ぐ） |
| | `_30_act` | invokeRest | Act | 前日残高に対して取引（acc-001 -200 / acc-002 +100） |
| | `_40_collect` | exportPostgres | Collect | 累積残高と取引履歴を収集 |
| | `_50_assert` | compare | Assert | **繰越**した累積残高（1300 / 2400）と取引の業務日付（20240102）を突合 |

`_020` 以降は reset / seed を行いません。**前業務日の残高を引き継いだまま updateBizdate で
業務日付だけを進める**ことで「業務日付をまたぐ」意味を示します。取引の業務日付は API の
payload では渡さず、SUT が業務日付テーブル `biz_calendar` から解決します。

## カスタムプラグイン（importMasterData / updateBizdate）

この example には**カスタムプロセスプラグイン**の実装例が 2 つあります。どちらも
**「カスタムプラグインは組込みプラグインを部品として再利用できる」**ことの実演です。

### importMasterData — 共通マスタデータの投入

`_010` の `_15_arrange` は `importMasterData`
（[`stfw/plugins/process/importMasterData/`](stfw/plugins/process/importMasterData/)）の実装例です。
**複数シナリオで共有するテスト共通のマスタ/参照データ**（ここでは口座名義 `users`）を投入します。

共通データは特定シナリオに属さず、プロジェクト共通の場所に集約します。

```
stfw/config/plugins/process/importMasterData/data/appdb/users.csv   # ← 全シナリオで共有
```

各シナリオは `importMasterData` プロセスの config で `tables: [users]` と宣言するだけで、
この共通データを取り込めます。

DB 投入は再実装せず、次の 2 段で実現しています。

1. **ファイル操作**: 組込み `importPostgres` が読む（プロセス配下の）`data/appdb/users.csv` から、
   共通データ `config/plugins/process/importMasterData/data/appdb/users.csv` へ **symlink を張る**
   （実体はコピーせず共通データを唯一の正とする）。プロセス配下の `data/` は gitignore 済み。
2. **委譲**: 委譲先の組込み `importPostgres` を先頭で確保（無ければ早期エラー）し、その `execute` を
   接続系の env を訳して呼び出す。

### updateBizdate — SUT の業務日付を進める

`_020` / `_030` の `_10_arrange` は `updateBizdate`
（[`stfw/plugins/process/updateBizdate/`](stfw/plugins/process/updateBizdate/)）の実装例です。
SUT の業務日付テーブル `biz_calendar`（単一行 `id='system'`）を、**実行中の bizdate ディレクトリの
業務日付**へ更新します。

観点は **stfw の実行コンテキスト env の活用**です。stfw は bizdate 階層以下のプロセスへ
`stfw_bizdate`（YYYYMMDD）を注入するため、シナリオ側に日付をハードコードする必要がなく、
**bizdate ディレクトリ名が業務日付の唯一の正**になります。

1. **CSV 生成**: `stfw_bizdate` から `data/appdb/biz_calendar.csv`（プロセス配下）を生成する。
   プロセス配下の `data/` は gitignore 済み。
2. **委譲**: 組込み `clearPostgres` で `biz_calendar` を初期化（`COPY` は追記のため）し、
   組込み `importPostgres` の `execute` を接続系の env を訳して呼び出す。

`plugins/process/{type}/` に置くだけで組込みより優先して解決されます（プラグイン契約の詳細は
各プラグインの README（[importMasterData](stfw/plugins/process/importMasterData/README.md) /
[updateBizdate](stfw/plugins/process/updateBizdate/README.md)）と
[`../../docs/GUIDE.md`](../../docs/GUIDE.md) を参照）。

## わざと失敗させてみる

Assert が本当に効いていることを確かめるには、期待値を書き換えて再実行します。

```sh
# 例: Day1 の期待残高を 1500 → 9999 に変える
vi stfw/scenario/daily-balance/_020_20240101/_50_assert_compare/expect/_40_collect_exportPostgres/appdb/accounts.csv
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
  `config/config.yml` を読み取ります。本例では assert プロセスに `REQ-01`〜`REQ-03` を紐づけ、
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
│   ├── schema.sql      # biz_calendar(業務日付) / users(マスタ) / accounts(残高) / transactions
│   └── Dockerfile
└── stfw/               # stfw プロジェクト (stfw init 相当 + シナリオ)
    ├── stfw.yml                                             # stfw.db.* = DB 接続先の単一ソース
    ├── config/
    │   ├── inventory/local.yml
    │   └── plugins/process/                                 # プロセスプラグインの共通設定
    │       ├── {clear,import,export}Postgres/config.yml     #   接続系 (${stfw_db_*} 参照)
    │       ├── updateBizdate/config.yml                     #   DB 接続系を共通化
    │       ├── importMasterData/
    │       │   ├── config.yml                               #   DB 接続系を共通化
    │       │   └── data/appdb/users.csv                     #   シナリオ共通のマスタデータ
    │       └── compare/compare_layout/                      #   シナリオ共通の比較レイアウト
    │           ├── accounts.json                            #   残高 CSV (id をキーに突合)
    │           └── transactions.json                        #   取引履歴 CSV (連番 id を除外)
    ├── plugins/            # カスタムプラグイン
    │   └── process/
    │       ├── importMasterData/   # 共通データ → 組込み importPostgres へ委譲
    │       └── updateBizdate/      # stfw_bizdate → biz_calendar 更新 (clear/import へ委譲)
    ├── docs/               # リバース生成物 (spec + doc の実出力例)
    │   ├── daily-balance.yml
    │   └── daily-balance.md
    └── scenario/daily-balance/
        ├── _010_20240101/  # データ準備: clear→importMasterData→import (取引は流さない)
        ├── _020_20240101/  # Day1: updateBizdate→act→collect→assert
        └── _030_20240102/  # Day2: updateBizdate→act→collect→assert (繰越)
```

接続情報は config に直書きせず、inventory（ホスト解決）と secret（パスワード）から解決します。
`run.sh` が生成する `stfw/config/encrypt/`・`stfw/config/passwd/`（デモ用鍵・クレデンシャル）は
git 管理外です。

## 比較レイアウトの共通化

エクスポート CSV の突合は、**シナリオ共通の比較レイアウト**
（`stfw/config/plugins/process/compare/compare_layout/*.json`）で項目単位・キー対応付けに
しています。行全体のテキスト比較と違い、物理的な行順に依存しません。

- **transactions.json**: 連番 `id`（BIGSERIAL。`TRUNCATE` でシーケンスがリセットされず
  再実行のたびにずれる）を `criteria: "Ignore"` で比較除外し、`account_id` + `bizdate` を
  `compareKey`（行の対応付けキー）、`amount` を `criteria: "Equal"` で厳密比較
- **accounts.json**: `id` を `compareKey`、`balance` を `criteria: "Equal"`

compare プラグインはこのディレクトリを `COMPAREFILES_CLASSPATH` として compare-files へ渡します。
レイアウトは後勝ちマージのため、プロセス固有の上書きが必要なら従来どおり各プロセスの
`config/compare_layout/` に同じ `fileRegexPattern` で定義できます。起動設定
（`compare_files.{json,yaml,yml}`）はプロセスローカル `config/` が最優先です（AS-BUILT §4.11）。
レイアウトの書き方は compare-files リポジトリの
[比較レイアウトリファレンス](https://github.com/scenario-test-framework/compare-files/blob/main/docs/compare_layout.md)
を参照してください。

## 設定の共通化（プロジェクト共通 config とプロセス差分）

プラグインの設定は「プラグイン既定 → プロジェクト共通 `config/plugins/process/{type}/config.yml`
→ プロセス `config/config.yml`」の 3 層チェーンで解決されます（全プラグイン共通の仕組み）。
**シナリオを跨いで同じ設定はプロジェクト共通に置き、プロセスには差分だけを書く**のがルールです。

この example では次を共通化しています。

| type | 共通化した設定 | プロセス側の差分 |
|---|---|---|
| `clearPostgres` / `importPostgres` / `exportPostgres` | 接続系（host_group / port / database / user） | `tables` のみ |
| `importMasterData` / `updateBizdate`（カスタム） | 接続系 | `tables` のみ / 差分なし |
| `invokeRest` | `host_group: api` | 差分なし（スクリプトは既定の `script.js`） |
| `compare` | `compare_files_version` | 差分なし |

特に DB 接続系は、各シナリオのプロセス config に書くとユーザーを変えたときに全プロセスを
直す必要があり、直し忘れやすくなります。そこで **接続系はプロジェクト共通の上書き設定に
一本化** し、各プロセスの `config/config.yml` は `tables` だけを書きます。

```
stfw.yml                                    ← host_group / database / user を単一ソースで宣言 (stfw.db.*)
config/plugins/process/{type}/config.yml   ← ${stfw_db_*} 参照を type ごとに共通定義
scenario/.../{process}/config/config.yml   ← tables のみ (差分がなければ {})
```

- 設定の上書きチェーン（§8.1）: プラグイン既定 → `config/plugins/process/{type}/` →
  各プロセス `config/config.yml`。共通値は真ん中の層に置きます。
- `host_group` / `database` / `user` の値そのものは **`stfw.yml` の `stfw.db.*` が単一ソース**です。
  共通設定は `${stfw_db_host_group}` / `${stfw_db_database}` / `${stfw_db_user}` で参照します。
  `stfw run` 開始時に stfw.yml の値が環境へ export され、config チェーンの `${...}` 展開で
  解決されます（AS-BUILT §8.2）。
- `port` は**プラグイン既定（5432）と同値なら書きません**。既定値の再記載は重複です
  （変える場合だけ共通設定で上書きします）。

```yaml
# stfw/stfw.yml — 接続先 DB の identity をここ 1 か所で管理 (探しやすい)
stfw:
  db:
    host_group: db
    database: appdb
    user: appuser
```

> SUT 側（`compose.yaml` の postgres/api）は「テスト対象そのものの DB 定義」なので compose に
> 置きます。`stfw.yml` の `stfw.db.*` はそこへ接続する側の宣言で、値を一致させます。
> 接続先ホスト・パスワードは `stfw.yml` に書かず、inventory + secret で解決します（禁止契約）。
