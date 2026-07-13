---
name: stfw-scenario
description: stfw (scenario test framework) のシナリオテスト一式を、SUT の設計情報 (要件定義・アーキテクチャ・OpenAPI Spec・DDL 等) とテストのゴールから生成する。inventory・secret セットアップスクリプト・カスタムプラグイン・scenario tree・期待値・比較レイアウトまでを揃え、stfw validate 通過まで面倒を見る。ユーザーが「stfw のシナリオを作って」「シナリオテストを組みたい」「この API / バッチのシナリオテスト」「stfw プロジェクトをセットアップ」「業務日付をまたぐテスト」「validate が通らない」などに言及したら必ずこのスキルを使うこと。既存シナリオへの業務日付・プロセス追加、期待値・カスタムプラグインの修正にも使う。
---

# stfw シナリオ一式生成

stfw は、業務日付 (bizdate) をまたぐシナリオテストをディレクトリ規約で記述し、
単一バイナリで自動実行する CLI。このスキルは SUT の設計情報とテストゴールから、
実行に必要な一式 (inventory / secret / カスタムプラグイン / scenario tree / 期待値) を生成する。

**ゴールは `stfw validate {scenario}` の通過まで。** 実行 (`stfw run`) は SUT 環境に
依存するため、引き渡し手順を提示してユーザーに委ねる。

## 0. 前提チェック

作業を始める前に確認する:

1. **stfw がインストール済みか・バージョンは十分か**: `stfw --version`。無ければ
   <https://github.com/scenario-test-framework/stfw#installation> を案内して中断する。
   **バージョンコアが 1.3.0 未満の場合もアップグレードを案内して中断する** (本スキルは
   `scenario scaffold` / `on_mismatch` / 部分実行など v1.3.0 までの機能を前提とする。
   旧タグには `docs/GUIDE.md` や examples が無いものもあり、参照取り込みも成立しない)。
   判定は `-dev` 等のプレリリース suffix を除いたコアで行う (`1.3.0-dev` → コア `1.3.0` で
   合格)。バージョン文字列は後述の参照ドキュメントの ref 決定にも使う
   (リリース版 `1.3.0` → タグ `v1.3.0`、suffix 付き → `master`)。
2. **プロジェクトルートか**: カレントに `stfw.yml` があるか。無ければ Step 3 で `stfw init` から始める。
3. **compare (Assert) を使う見込みがあるか**: ほぼ全シナリオで使う。比較レイアウトの生成は
   **compare-layout スキルに委譲する**ため、依存チェックを行う (§依存: compare-layout)。

## 1. 参照ドキュメントの取り込み

このスキルは規約の要点だけを持つ。**詳細仕様は必ず stfw リポジトリの一次ドキュメントで
確認し、推測で書かない。** 取り込みは次の順でフォールバックする:

1. **shallow clone (推奨)**: 1 回で docs / プラグイン README / 実例が全部揃う。
   ```bash
   git clone --depth 1 --branch v{stfwバージョン} https://github.com/scenario-test-framework/stfw {作業dir}/stfw-docs
   # タグが無い (-dev 版) 場合は --branch を外して master を取る
   ```
2. **raw URL の個別フェッチ**: clone できない場合。URL は `references/urls.md` の一覧から。
3. **どちらも失敗する場合**: `references/urls.md` の URL 一覧をそのままユーザーに提示し、
   ブラウザで開ける環境での取得かローカルパスの指定を依頼する。

読む順序 (全部を読まず、必要なものだけ):

| 資料 | いつ読むか |
|---|---|
| `docs/GUIDE.md` | **必読**。4 フェーズの考え方・プラグイン一覧・通しの実例・spec/scaffold の使い方 |
| `assets/plugins/process/{type}/README.md` | 使うプラグインが決まったら、その設定キーの正確な仕様 |
| `docs/AS-BUILT.md` | 詳細契約が必要なとき。§3 ディレクトリ規約 / §4 プラグイン実行契約・公開 env / §8 設定と上書きチェーン / §9 secret・inventory / §12 spec スキーマ |
| `examples/daily-balance/` | 実物の書き方に迷ったら。カスタムプラグイン実例 (updateBizdate / importMasterData) を含む |

## 2. 入力の収集

ユーザーから次を集める。不足があれば推測せず質問する:

- **SUT の設計情報**: 要件定義、アーキテクチャ図、OpenAPI Spec、DDL、バッチ仕様など。
  ファイルパスや URL で受け取り、実物を読む。
- **テストのゴール**: 何を検証したいか (例:「日次バッチの残高繰越が 2 営業日で正しいこと」)。
- **環境情報**: 接続先ホスト (DB / API / バッチサーバ)、DB 種別 (MySQL / PostgreSQL / Redis)、
  接続ユーザー。**パスワードの値は聞かない・ファイルに書かない** (secret の仕組みで実行時に設定)。

## 3. シナリオ設計 → ユーザー確認

生成に入る前に、設計サマリを提示して合意を取る。設計の軸は 2 つ:

- **業務日付の分割**: データ準備日 / Day1 / Day2 ... に分け、繰越・日またぎをどう検証するか。
  同一 bizdate でも「準備」と「実行」で bizdate ディレクトリを分けてよい (seq が実行順)。
- **各 bizdate 内の 4 フェーズ**: Arrange (準備) → Act (実行) → Collect (収集) → Assert (検証)。
  グループ名をフェーズ名に揃えると、ディレクトリ名だけで流れが読める。

提示フォーマット例:

```
scenario/{name}/
├── _010_{bizdate}/  データ準備
│   ├── _10_arrange_clearPostgres     対象テーブルの全行削除
│   └── _20_arrange_importPostgres    初期データ投入 (data/appdb/*.csv)
├── _020_{bizdate}/  Day1
│   ├── _30_act_invokeRest            取引 POST (OpenAPI の /transactions)
│   ├── _40_collect_exportPostgres    残高・履歴を CSV 収集
│   └── _50_assert_compare            期待値と突合
└── _030_{翌bizdate}/ Day2 (前日繰越)
    └── ...
```

あわせて確認する項目:

- **プラグイン選定**: まず組込み (GUIDE §2 の一覧) で組む。組込みで表現できない処理
  (SUT 固有の業務日付更新、独自プロトコルなど) だけカスタムプラグインにする。
- **inventory グループ設計**: プラグインの `host_group` から逆算 (例: `db` / `api` / `batch`)。
- **secret 一覧**: `{host}-{user}` のペアを列挙 (DB 接続、SSH 接続)。
- **compare の運用**: 回帰テスト (差分で停止 = `on_mismatch: error` 既定) か、
  差分確認 (最後まで流して Warn 一覧 = `on_mismatch: warn`) か。

## 4. 一式の生成

### 4.1 プロジェクト初期化 (新規のみ)

```bash
stfw init --skip-plugin-init   # プラグインの外部バイナリ取得は引き渡し手順に回す
```

`stfw.yml` が既にあるプロジェクトでは実行しない (再初期化はエラー)。
init が展開する sample シナリオは、ユーザーに確認して不要なら削除する。

### 4.2 inventory

`config/inventory/{env}.yml` に、設計したグループ → ホストの対応を書く:

```yaml
stfw_inventory:
  - db:  [postgres-host]
  - api: [api-host]
```

どのファイルが読まれるかは **`stfw.yml` の `stfw.inventory`** で決まる (init 既定は
`staging.yml`。ルート直下ではなく `stfw:` 配下のキーである点に注意)。
既定と違うファイル名にする場合は `stfw.yml` も併せて更新する:

```yaml
stfw:
  inventory: local.yml   # config/inventory/local.yml を読む
```

`stfw inventory list {group}` で解決結果を確認できる。

### 4.3 secret セットアップスクリプト

**パスワードはリポジトリに残さない**。ユーザーが実行時に環境変数で渡すスクリプト
`setup-secrets.sh` を生成する:

```bash
#!/bin/bash
set -euo pipefail
# 鍵が無ければ生成 (既存キーは維持: keygen は既存があるとエラーになる)
[ -f config/encrypt/key.txt ] || stfw secret keygen
# パスワードは環境変数から。値を argv・ファイルに書かない
printf '%s' "${APPDB_PASSWORD:?APPDB_PASSWORD を設定してください}" | stfw secret set -f postgres-host appuser
```

secret のキーは `{host}-{user}`。プラグインは inventory で解決したホストと config の
`user` からこのキーでパスワードを取り出す。**config への接続情報直書きは禁止契約。**

### 4.4 scenario tree — spec 経由で骨格を作り、葉を実装する

手でディレクトリを掘らず、**spec YAML → `stfw scenario scaffold`** の公式ルートを使う
(規約違反の混入を防げる。spec スキーマは AS-BUILT §12.3):

1. 設計サマリを spec YAML (`docs/{scenario}.yml` 相当) に落とす。各 process には
   `description` と、検証したい要求に紐づく `requirement_specifications` を書く
   (doc のトレーサビリティ表になる)。description にコロンや `#` を含めるときは
   YAML の文法エラーになるためクォートする (`description: "回帰運用: ..."`)。
2. `stfw scenario scaffold {spec}.yml` で骨格 (metadata.yml + config/config.yml) を生成。
3. **葉を実装する** (scaffold は生成しない。人＝このスキルが書く):
   - `data/{database}/{table}.csv` — 投入データ (**RDBMS 系の契約**)。ヘッダー付き、
     NULL は `\N`。DDL から列を起こす。export の出力形式と同一なので往復できる。
     **Redis 系は別契約**: 設定は `tables` ではなく `key_patterns`、入力は
     `data/{host}/{name}.csv` (ヘッダー `key,type,ttl,value`)。必ずプラグイン README で確認する。
   - clear 系 (clearPostgres / clearMysql) の挙動は**リリースで異なる**。
     **リリース v1.3.0 まで**: 配列順に 1 テーブルずつ TRUNCATE するため、他テーブルから
     FK 参照される「親」は初期化できない (制約の存在だけで失敗)。親を含む場合は
     scripts 等で子 → 親順の DELETE を代替として組む。
     **v1.3.0 より後のリリース**: 全テーブルを 1 トランザクションの DELETE で空にする。
     FK 制約は無効化されないため、**`tables` は FK の子 → 親の順に列挙**する
     (参照行が残る親の削除は DBMS の FK 違反エラーで失敗する)。
     **`-dev` 等のプレリリース版はバージョン文字列で挙動を判定しない**。実際に解決される
     プラグイン実体で判定する: プロジェクトに同名の上書き `plugins/process/clearPostgres/`
     があれば組込みより優先されるのでそちらの `bin/run/execute` を、無ければ展開済みの
     組込み実体 `.stfw/plugins/process/clearPostgres/bin/run/execute` を見る
     (未展開なら `stfw plugin install clearPostgres` で展開してから確認する)。
     `DELETE FROM` なら新挙動、`TRUNCATE TABLE` なら旧挙動。
   - `script.js` — invokeRest 用 k6 スクリプト。OpenAPI Spec からリクエストを起こす。
     接続先は `__ENV.stfw_target_host`。`thresholds: { checks: ['rate==1.0'] }` で
     失敗時に非 0 終了させる (これが無いと NG レスポンスでも Act が成功扱いになる)。
   - `expect/{収集プロセスのディレクトリ名}/{database}/{table}.csv` — 期待値。
     業務ロジックから計算して起こし、計算根拠をコメントか README に残す。
4. 設定は 3 層の上書きチェーン (プラグイン既定 → プロジェクト共通
   `config/plugins/process/{type}/config.yml` → プロセス `config/config.yml`)。
   **接続系・共通設定はプロジェクト共通に、プロセスには差分だけ**を書く。
   さらに DB の `host_group` / `database` / `user` は複数プラグイン
   (clear/import/export など) に重複しやすいので、**`stfw.yml` の `stfw.db.*` を単一
   ソース**にし、共通 config からは `${stfw_db_host_group}` 等で参照する
   (run 開始時に env へ export され `${...}` 展開で解決。AS-BUILT §8.2、
   daily-balance README「設定の共通化」参照)。

シナリオ完成後に `stfw scenario reverse {scenario}` を実行し、spec / doc
(`docs/{scenario}.{yml,md}`) を tree から再生成しておく (doc はレビュー資料になる)。

### 4.5 カスタムプラグイン (必要な場合のみ)

`plugins/process/{type}/` に次の構成で作る。実例は examples/daily-balance の
`updateBizdate` (組込みへの委譲) と `importMasterData` (共通データ投入):

```
plugins/process/{type}/
├── plugin.yml          # requires: [psql 等の依存コマンド]
├── config.yml          # 既定設定 (stfw.process.{type}.*) + 各キーの説明コメント
├── README.md           # 何をするか・設定キー
└── bin/
    ├── install/install       # 依存の用意。不要なら exit 0
    ├── install/is_installed   # stdout に `true` を出力して exit 0 (exit 0 だけでは未インストール扱い)
    └── run/pre_execute, execute, post_execute  # 不要フェーズは exit 0 の空実装
```

守る契約 (AS-BUILT §4):

- 入力は env のみ: config は `stfw_process_{type}_{key}` にフラット化されて渡る。
  実行コンテキスト (`stfw_bizdate`, `stfw_process_dir`, `STFW_PROJ_DIR_DATA` 等) も env。
- 終了コード: `0`=Success / `3`=Warn / `6`=Error (非 0 非 3 は Error 扱いで後続ノードは実行されない)。
- 実装言語は任意 (実行可能ファイルであればよい)。
- DB 操作などは再実装せず、**組込みプラグインの execute へ委譲する**のが定石
  (委譲先の env `stfw_process_{組込みtype}_*` を export して呼ぶ。updateBizdate 参照)。

### 4.6 比較レイアウト — compare-layout スキルへ委譲

行全体のテキスト比較で足りない場合 (連番 ID・更新時刻列の除外、キー列での行対応付け) は
比較レイアウト JSON が必要。**自分で書かず compare-layout スキルを呼ぶ**:

1. 収集エビデンスのサンプル (または DDL から起こした同型 CSV) を用意する。
2. compare-layout スキルに、そのファイルを渡してレイアウト生成を依頼する。
3. 生成物の置き場所: シナリオ横断で共通なら
   `config/plugins/process/compare/compare_layout/{table}.json`、
   プロセス固有の上書きは各プロセスの `config/compare_layout/` (後勝ち)。

**compare-layout が未導入のまま進める場合の暫定運用**: 可変列 (`now()` 既定の
タイムスタンプ、連番 ID など。DDL の DEFAULT から判別) を含むテーブルは、レイアウト
なしの行全体比較では必ず不一致になる。その場合は compare プロセスを分割し、
可変列を含むテーブル側だけ `on_mismatch: warn` にして完走可能にしておく
(固定値のみのテーブルは `error` のまま)。レイアウト導入後に可変列を Ignore にして
`error` へ戻すことを、引き渡し手順の TODO に明記する。

## 5. validate 通過まで

```bash
stfw validate {scenario}
```

エラー (exit 6) が出たら、メッセージの指す規約違反 (ディレクトリ名・プラグイン解決・
config 欠落) を修正して再実行する。**警告のみ (exit 0) になるまで繰り返す。**
validate は静的検証のみで、SUT への接続は行わない (secret 未設定でも通る)。

## 6. 引き渡し

最後に、ユーザーが実行するまでの手順を README かメッセージで提示する:

```bash
export APPDB_PASSWORD=...        # 1. secret 登録 (パスワードは env で渡す)
./setup-secrets.sh
stfw plugin install {type}       # 2. 外部バイナリ取得 (run は自動取得しない)
stfw ssh trust {group}           # 3. ssh 系プラグインを使う場合のみ
stfw run {scenario}              # 4. 実行 (exit 0=全 Success / 3=Warn あり / 6=Error)
stfw status                      # 5. 結果確認
stfw report                      # 6. HTML レポート
```

手順 2 は固定コマンドではなく、**生成した scenario tree で使っている type を列挙して**
外部バイナリが必要なものを個別に案内する (invokeRest / invokeWeb = k6、
compare = compare-files、collectLog = logfilter。カスタムプラグインも install を持つなら含める)。

失敗時のやり直しには部分実行がある: `stfw run {scenario} --from {bizdate_dir}/{process_dir}`
(そこから最後まで) / `--only` (そこだけ)。スキップしたノードの副作用は再現されない点を添える。

## 依存: compare-layout スキル

比較レイアウト生成 (§4.6) の前に、compare-layout スキルが使えるか確認する:

```bash
ls .claude/skills/compare-layout/SKILL.md ~/.claude/skills/compare-layout/SKILL.md 2>/dev/null
```

どちらにも無ければ、次を提示してインストールしてもらってから §4.6 に進む
(それまで比較レイアウトは保留にし、他の生成を先に進めてよい):

```
比較レイアウトの生成には compare-layout スキルが必要です。次のコマンドでインストールできます:

  npx skills add scenario-test-framework/compare-files --skill compare-layout -a claude-code

参照: https://github.com/scenario-test-framework/compare-files/tree/master/.claude/skills/compare-layout
```

## 焼き込み済みの最小規約 (オフライン時の拠り所)

- ディレクトリ規約: `scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/`。
  bizdate は 8 桁 `YYYYMMDD`、group に `_` 禁止、seq は数字列。
  実行順は**ディレクトリ名の辞書順 (バイト順)** で `_10` が `_2` より先になるため、
  同一階層の seq は桁数を揃えてゼロパディングして生成する (`_010` / `_020` ...)。
- プロセスは `setup → pre_execute → execute → post_execute → teardown` の 5 フェーズ。
  Error になると後続ノードは実行されない (Blocked と記録されるのは scripts 内の
  列挙済み後続ステップのみ。後続 process / bizdate は実行もジャーナル記録もされない)。
- compare のエビデンス規約: `expect/` (git 管理・期待値) / `actual/` (収集 evidence への
  自動 symlink) / `result/` (比較結果)。`expect/` 直下は**収集プロセスのディレクトリ名**。
- 組込みプラグイン: clear/import/export × Mysql/Postgres/Redis、invokeRest/invokeWeb (k6)、
  sshExec/scpPut/collectLog/collectFile (ssh)、compare、scripts (汎用)。
