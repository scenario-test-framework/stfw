# importMasterData（カスタムプラグインの実装例）

`config` に持たせた CSV を、データストアのマスタ/参照テーブルへ投入する Arrange 用の
**カスタムプロセスプラグイン**です。件数が少なく scenario と一緒に管理したいマスタデータ
（口座名義・区分値・支店マスタなど）を、別ファイルに切り出さず config へ同梱するために使います。

このプラグインの主眼は **「カスタムプラグインは組込みプラグインを部品として再利用できる」**
ことの実演です。DB 投入ロジックは再実装せず、次の 2 段で実現します。

1. **ファイル操作**: config の CSV を、組込み `importPostgres` が読む
   `data/{database}/{table}.csv` へ書き出す。
2. **委譲**: 組込み `importPostgres` の `bin/run/execute` を、接続系の env を訳して呼び出す
   （`stfw plugin install importPostgres` で `.stfw/` 配下へ materialize されたものを exec）。

## 設定

`config/config.yml` の `stfw.process.importMasterData`:

| キー | 説明 |
|---|---|
| `host_group` / `port` / `database` / `user` | 接続系（組込み `importPostgres` と同じ。接続情報は inventory + secret で解決し config に直書きしない） |
| `tables[].name` | 投入先テーブル名 |
| `tables[].csv` | ヘッダー付き CSV 本文（`importPostgres` と同じ形式。NULL は `\N`） |

```yaml
stfw:
  process:
    importMasterData:
      host_group: db
      database: appdb
      user: appuser
      tables:
        - name: users
          csv: |
            id,name,email
            acc-001,Alice,alice@example.com
            acc-002,Bob,bob@example.com
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
  `STFW_PROJ_DIR_DATA` など実行コンテキスト）。
- 出力はリターンコード（`0`=Success / `3`=Warn / `6`=Error）。
- `plugins/process/` に置くだけで `stfw run` / `stfw validate` が解決します（組込みより優先）。

> 注: config の値は env 展開（`$VAR`）を通るため、CSV 本文に `$` を含めないでください。
