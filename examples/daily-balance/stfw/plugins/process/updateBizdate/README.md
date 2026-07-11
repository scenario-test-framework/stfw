# updateBizdate（カスタムプラグインの実装例）

**SUT の業務日付テーブル `biz_calendar`** を、実行中の業務日付階層の bizdate へ更新する
Arrange 用の**カスタムプロセスプラグイン**です。各 bizdate ディレクトリの先頭（arrange）に
置くことで、「SUT の業務日付を進めてから当日の取引を流す」という業務システムの日次運転を
ディレクトリ規約だけで表現できます。

## 仕組み（stfw の実行コンテキスト env の活用）

stfw は bizdate 階層以下のプロセスへ **`stfw_bizdate`（YYYYMMDD）** を注入します。
本プラグインはこの値から `biz_calendar` の投入 CSV（単一行 `id='system'`）を生成し、
SUT が参照する業務日付を書き換えます。

```
_020_20240101/_10_arrange_updateBizdate/   ← stfw_bizdate=20240101 が注入される
_030_20240102/_10_arrange_updateBizdate/   ← stfw_bizdate=20240102 が注入される
```

シナリオ側に日付のハードコードは不要で、bizdate ディレクトリ名が唯一の正になります。

## 設計（組込みプラグインの再利用）

importMasterData と同じく、主眼は **「カスタムプラグインは組込みプラグインを部品として
再利用できる」** ことの実演です。DB 反映ロジックは再実装せず、次の 2 段で実現します。

1. **CSV 生成**: `stfw_bizdate` から `data/{database}/biz_calendar.csv`（プロセス配下）を
   生成する。プロセス配下の `data/` は実行時生成物として gitignore する。
2. **委譲**: 組込み `clearPostgres` で `biz_calendar` を初期化（`COPY` は追記のため）し、
   組込み `importPostgres` の `bin/run/execute` を接続系の env を訳して呼び出す
   （委譲先は先頭で確保し、用意できなければ早期にエラー終了する）。

## 設定

`config/config.yml` の `stfw.process.updateBizdate`:

| キー | 説明 |
|---|---|
| `host_group` / `port` / `database` / `user` | 接続系（組込み `clearPostgres` / `importPostgres` と同じ。接続情報は inventory + secret で解決し config に直書きしない） |

更新値は `stfw_bizdate`・対象テーブルは `biz_calendar` 固定のため、これ以外の設定はありません。
この example では接続系を `config/plugins/process/updateBizdate/config.yml` で共通化しており、
各プロセスの `config/config.yml` には設定キーがありません。

## プラグインの作り方（このリポジトリのプラグイン契約）

```
plugins/process/updateBizdate/
├── plugin.yml              # requires: [psql]（前提コマンド）
├── config.yml              # 既定設定（プロセスの config/config.yml で上書き）
├── bin/install/is_installed # 前提コマンドが揃えば "true" を出力
├── bin/install/install      # 外部バイナリの provisioning（本例は不要 → exit 0）
└── bin/run/{pre_execute,execute,post_execute}   # 実行フェーズ
```

- 入力は env（`stfw_process_updateBizdate_*` = config のフラット化、`stfw_bizdate` /
  `stfw_process_dir` / `STFW_PROJ_DIR_DATA` など実行コンテキスト）。
- 出力はリターンコード（`0`=Success / `3`=Warn / `6`=Error）。
- `plugins/process/` に置くだけで `stfw run` / `stfw validate` が解決します（組込みより優先）。
