# importMasterData（カスタムプラグインの実装例）

**複数シナリオで共有するテスト共通のマスタ/参照データ**をデータストアへ投入する
Arrange 用の**カスタムプロセスプラグイン**です。口座名義・区分値・支店マスタなど、
特定シナリオに属さない共通データを 1 か所にまとめて管理し、各シナリオはそれを
「取り込む」だけにできます。

## 共通データの置き場

```
config/plugins/importMasterData/data/{database}/{table}.csv
```

シナリオ非依存のプロジェクト共通の場所です。任意のシナリオの `importMasterData`
プロセスが `tables` にテーブル名を宣言するだけで、この共通データを取り込めます。

```
config/plugins/importMasterData/data/
└── appdb/
    └── users.csv        # ← 全シナリオで共有する口座名義マスタ
```

## 設計（組込みプラグインの再利用）

このプラグインの主眼は **「カスタムプラグインは組込みプラグインを部品として再利用できる」**
ことの実演です。DB 投入ロジックは再実装せず、次の 2 段で実現します。

1. **ファイル操作**: 共通データ `config/plugins/importMasterData/data/{db}/{table}.csv` を、
   組込み `importPostgres` が読む `data/{db}/{table}.csv`（プロセス配下）へコピーする。
2. **委譲**: 組込み `importPostgres` の `bin/run/execute` を、接続系の env を訳して呼び出す
   （`stfw plugin install importPostgres` で `.stfw/` 配下へ materialize されたものを exec）。

## 設定

`config/config.yml` の `stfw.process.importMasterData`:

| キー | 説明 |
|---|---|
| `host_group` / `port` / `database` / `user` | 接続系（組込み `importPostgres` と同じ。接続情報は inventory + secret で解決し config に直書きしない） |
| `tables` | 取り込む共通データ（テーブル名）のリスト。`config/plugins/importMasterData/data/{database}/{table}.csv` を解決する |

```yaml
stfw:
  process:
    importMasterData:
      host_group: db
      database: appdb
      user: appuser
      tables:
        - users
```

## プラグインの作り方（このリポジトリのプラグイン契約）

```
plugins/process/importMasterData/
├── plugin.yml              # requires: [psql]（前提コマンド）
├── config.yml              # 既定設定（プロセスの config/config.yml で上書き）
├── bin/install/is_installed # 前提コマンドが揃えば "true" を出力
├── bin/install/install      # 外部バイナリの provisioning（本例は不要 → exit 0）
└── bin/run/{pre_execute,execute,post_execute}   # 実行フェーズ
```

- 入力は env（`stfw_process_importMasterData_*` = config のフラット化、`stfw_process_dir` /
  `STFW_PROJ_DIR_CONFIG` / `STFW_PROJ_DIR_DATA` など実行コンテキスト）。
- 出力はリターンコード（`0`=Success / `3`=Warn / `6`=Error）。
- `plugins/process/` に置くだけで `stfw run` / `stfw validate` が解決します（組込みより優先）。
