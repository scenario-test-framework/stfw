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

1. 依存サービス（postgres + トイ API = SUT）を起動
2. secret を準備（age 鍵生成 + DB パスワード登録）
3. プラグインの外部バイナリ（k6 / compare-files）を install
4. `stfw run daily-balance` を実行
5. HTML レポート配信（nginx）を起動 → http://localhost:8088

後片付け:

```sh
./run.sh --down
```

> レポートのポートを変えたい場合は `STFW_REPORT_PORT=9000 ./run.sh`。

## シナリオの流れ

初期残高 `acc-001=1000` / `acc-002=2000` に対し、2 業務日で取引を反映します。

プロセスのグループ名はフェーズ（arrange / act / collect / assert）に揃えてあり、
ディレクトリ名を見るだけで A→A→C→A の流れが読めます。

| bizdate | プロセス | プラグイン | フェーズ | 内容 |
|---|---|---|---|---|
| `_10_20240101` | `_10_arrange` | clearPostgres | Arrange | accounts / transactions を truncate |
| | `_20_arrange` | importPostgres | Arrange | 初期残高 CSV を投入 |
| | `_30_act` | invokeRest | Act | API へ取引 POST（acc-001 +500 / acc-002 +300） |
| | `_40_collect` | exportPostgres | Collect | 残高を `evidence/appdb/accounts.csv` へ |
| | `_50_assert` | compare | Assert | 期待残高（1500 / 2300）と突合 |
| `_20_20240102` | `_10_act` | invokeRest | Act | 前日残高に対して取引（acc-001 -200 / acc-002 +100） |
| | `_20_collect` | exportPostgres | Collect | 残高を収集 |
| | `_30_assert` | compare | Assert | **繰越**した累積残高（1300 / 2400）と突合 |

Day2 は reset / seed を行いません。**Day1 の残高を引き継ぐ**ことで「業務日付をまたぐ」意味を示します。

## わざと失敗させてみる

Assert が本当に効いていることを確かめるには、期待値を書き換えて再実行します。

```sh
# 例: Day1 の期待残高を 1500 → 9999 に変える
vi stfw/scenario/daily-balance/_10_20240101/_50_assert_compare/expect/_40_collect_exportPostgres/appdb/accounts.csv
docker compose run --rm stfw run daily-balance   # compare が差分を検出し Error 終了
```

## ドキュメント / spec（ラウンドトリップ）

シナリオのツリー（=正）から、人が読む**ドキュメント**と、機械可読な **spec** を投影できます。

| コマンド | 生成物 | 用途 |
|---|---|---|
| `stfw scenario doc daily-balance` | [`stfw/docs/daily-balance.doc.md`](stfw/docs/daily-balance.doc.md) | フェーズ推定・**要求トレーサビリティ表**つきの Markdown |
| `stfw scenario spec daily-balance` | [`stfw/docs/daily-balance.spec.yml`](stfw/docs/daily-balance.spec.yml) | ツリーと可逆な YAML（往復の出口） |
| `stfw scenario scaffold <spec.yml>` | ディレクトリ骨格 | spec からツリーを再生成（往復の入口） |

- doc / spec は各階層の `metadata.yml`（`description` / `requirement_specifications`）と
  `config/config.yml` を読み取ります。本例では assert プロセスに `REQ-01` / `REQ-02` を紐づけ、
  doc の「要求トレーサビリティ」表に「どの要求をどの process が検証するか」が出力されます。
- `spec → scaffold → spec` は**完全一致**（骨格：seq / bizdate / group / type / description /
  requirement_specifications / config.yml のサブツリー）。data CSV・script・expect などの葉は
  対象外です。

同梱の `stfw/docs/*.md` / `stfw/docs/*.spec.yml` は上記コマンドで生成した実出力です。手元で再生成するには（stfw サービスの作業ディレクトリ `/work` が `stfw/` にマウントされます）:

```sh
docker compose run --rm stfw scenario doc  daily-balance --out docs/daily-balance.doc.md
docker compose run --rm stfw scenario spec daily-balance --out docs/daily-balance.spec.yml
```

> 詳細は [`../../docs/GUIDE.md` §8](../../docs/GUIDE.md) を参照。

## ディレクトリ

```
examples/daily-balance/
├── run.sh              # 一発実行スクリプト
├── compose.yaml        # postgres + api(SUT) + stfw(:full) + nginx
├── sut/                # テスト対象システム (トイ REST API + スキーマ)
│   ├── main.go
│   ├── schema.sql
│   └── Dockerfile
└── stfw/               # stfw プロジェクト (stfw init 相当 + シナリオ)
    ├── stfw.yml
    ├── config/inventory/local.yml
    ├── docs/               # ラウンドトリップ生成物 (doc / spec の実出力例)
    │   ├── daily-balance.doc.md
    │   └── daily-balance.spec.yml
    └── scenario/daily-balance/
        ├── _10_20240101/   # Day1: arrange→arrange→act→collect→assert
        └── _20_20240102/   # Day2: act→collect→assert (繰越)
```

接続情報は config に直書きせず、inventory（ホスト解決）と secret（パスワード）から解決します。
`run.sh` が生成する `stfw/config/encrypt/`・`stfw/config/passwd/`（デモ用鍵・クレデンシャル）は
git 管理外です。
