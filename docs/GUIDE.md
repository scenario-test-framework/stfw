# stfw シナリオ作成ガイド

組込みプラグインを組み合わせて、実プロジェクトのシナリオテストを記述するためのガイドです。
[各プラグインの README](../assets/plugins/process/)（ツマミの詳細）に対して、本書は
**「なぜ・いつ・どう組むか」の通し解説**です。動く実例は
[`examples/daily-balance`](../examples/daily-balance/) を参照してください（本書はこの例を題材にします）。

## 1. 考え方: シナリオテストの 4 フェーズ

stfw のシナリオテストは、業務システムの一連の処理を次の 4 フェーズで記述します。

| フェーズ | 何をするか |
|---|---|
| **Arrange**（準備） | テスト前提のデータ・ファイルを外部システムへ配置し、状態を初期化する |
| **Act**（実行） | テスト対象システム（SUT）に取引を入力し、処理を起動する |
| **Collect**（収集） | 実行後の状態（DB・ファイル・ログ）をエビデンスとして取り出す |
| **Assert**（検証） | 収集したエビデンスを期待値と突合し、合否を判定する |

さらに stfw は**業務日付（bizdate）**を第一級で扱います。日次バッチのように「前日の結果を
翌日へ繰り越す」処理は、bizdate ディレクトリを昇順に実行することで自然に表現できます。

```
scenario/{シナリオ}/_{seq}_{bizdate}/_{seq}_{group}_{type}/
                     ~~~~~~~~~~~~~~~  業務日付ごとに区切り、昇順で実行
                                      ~~~~~~~~~~~~~~~~~~~~~~ 1 プロセス = 1 プラグイン
```

## 2. フェーズと組込みプラグイン

`_{seq}_{group}_{type}` の **type** がプラグイン種別です。フェーズごとに次のプラグインを使います。

| フェーズ | プラグイン | 役割 |
|---|---|---|
| Arrange | `importMysql` / `importPostgres` / `importRedis` | データストアへ期待データを投入 |
| | `clearMysql` / `clearPostgres` / `clearRedis` | データストアを初期化（truncate / flush） |
| | `scpPut` | ローカルのファイル群をリモートホストへ配置 |
| Act | `invokeRest` | API へ取引入力・レスポンス検証（grafana k6） |
| | `invokeWeb` | ブラウザ操作（k6 browser、headless Chromium） |
| | `sshExec` | リモートホストでスクリプト・バッチを実行 |
| Collect | `collectLog` | リモートのログを業務日付で絞り込み収集 |
| | `collectFile` | リモートのファイルを収集 |
| | `exportMysql` / `exportPostgres` / `exportRedis` | データストアの内容を CSV エビデンス化 |
| Assert | `compare` | 期待値（expect/）と収集結果（actual/）を突合 |
| （汎用） | `scripts` | 任意言語の実行可能ファイルを昇順実行 |

> プロセスは `setup → pre_execute → execute → post_execute → teardown` の順で実行され、
> Error（exit 6 等の非 0・非 3）で後続はブロックされシナリオは失敗します。
> exit 3 は Warn として記録され実行は続行します（exit 0=Success / 3=Warn / 6=Error。AS-BUILT §4.6）。

## 3. 通しの例: daily-balance

口座残高の日次バッチを 2 業務日で検証します（[`examples/daily-balance`](../examples/daily-balance/)）。

```
scenario/daily-balance/
├── _010_20240101/                    # データ準備 (取引は流さない)
│   ├── _10_arrange_clearPostgres/    # Arrange: truncate
│   ├── _15_arrange_importMasterData/ # Arrange: 共通マスタ投入 (カスタム)
│   └── _20_arrange_importPostgres/   # Arrange: 初期残高を投入
│       └── data/appdb/accounts.csv
├── _020_20240101/                    # Day1
│   ├── _10_arrange_updateBizdate/    # Arrange: SUT の業務日付を 20240101 へ (カスタム)
│   ├── _30_act_invokeRest/           # Act: 取引 POST
│   │   └── script.js
│   ├── _40_collect_exportPostgres/   # Collect: 残高・取引履歴を収集
│   │   └── evidence/appdb/accounts.csv   (自動生成)
│   └── _50_assert_compare/           # Assert: 期待残高と突合
│       └── expect/_40_collect_exportPostgres/appdb/accounts.csv
└── _030_20240102/                    # Day2 (reset/seed なし = 前日を繰越)
    ├── _10_arrange_updateBizdate/    # Arrange: SUT の業務日付を 20240102 へ (カスタム)
    ├── _30_act_invokeRest/
    ├── _40_collect_exportPostgres/
    └── _50_assert_compare/
```

プロセスのグループ名（`_{seq}_{group}_{type}` の中央）をフェーズ名に揃えると、
ディレクトリ名だけで A→A→C→A の流れが読めます。同じ業務日付でも「データ準備」と
「実行」のように bizdate ディレクトリを分けられます（seq が実行順・bizdate が業務日付）。
`updateBizdate` は stfw が注入する `stfw_bizdate` から SUT の業務日付テーブルを更新する
カスタムプラグインの実装例です（[`examples/daily-balance`](../examples/daily-balance/) 参照）。

### Arrange — データを整える

プラグインの設定は `stfw.process.{type}` 配下に書き、3 層の上書きチェーンで解決されます
（プラグイン既定 → プロジェクト共通 `config/plugins/process/{type}/config.yml` →
プロセス `config/config.yml`。全プラグイン共通の仕組み。AS-BUILT §8.1）。

**シナリオを跨いで同じ設定（接続系・バージョン指定など）はプロジェクト共通に置き、
プロセスには差分だけを書く**のが規約です。**接続情報（ホスト・パスワード）は config に
直書きせず**、inventory と secret から解決します（§4）。

```yaml
# config/plugins/process/importPostgres/config.yml — プロジェクト共通 (全シナリオに効く)
stfw:
  process:
    importPostgres:
      host_group: db        # inventory グループ → 接続先ホスト
      database: appdb
      user: appuser          # パスワードは secret {host}-{user} で解決
```

```yaml
# _20_arrange_importPostgres/config/config.yml — プロセスは差分のみ
stfw:
  process:
    importPostgres:
      tables: [accounts]     # data/appdb/accounts.csv を投入
```

投入 CSV（`data/{database}/{table}.csv`）はヘッダー付き・NULL は `\N`。
これは `exportPostgres` の出力形式と同じで、収集結果をそのまま次回の投入データにできます。

### Act — システムを叩く

`invokeRest` は grafana k6 でスクリプトを実行します。`host_group` の先頭ホストが
`__ENV.stfw_target_host` として k6 に渡ります。

```yaml
# _30_act_invokeRest/config/config.yml
stfw:
  process:
    invokeRest:
      host_group: api
      script: script.js
```

```js
// script.js — 閾値を満たさない (=非 201) 応答があれば k6 が非 0 終了し Act 失敗
const host = __ENV.stfw_target_host;
export const options = { vus: 1, iterations: 1, thresholds: { checks: ['rate==1.0'] } };
export default function () {
  const res = http.post(`http://${host}:8080/transactions`, JSON.stringify(tx), ...);
  check(res, { 'status is 201': (r) => r.status === 201 });
}
```

### Collect — 結果を取り出す

`exportPostgres` は対象テーブルを `evidence/{database}/{table}.csv` へ書き出します。

```yaml
# _40_collect_exportPostgres/config/config.yml
stfw:
  process:
    exportPostgres:
      host_group: db
      database: appdb
      user: appuser
      tables: [accounts]
```

### Assert — 期待値と突合

`compare` は `expect/` と `actual/` を突合します。詳細は §5。期待残高を CSV で置くだけです。

```
# _50_assert_compare/expect/_40_collect_exportPostgres/appdb/accounts.csv
id,balance
acc-001,1500     # 1000 + 500
acc-002,2300     # 2000 + 300
```

Day2（`_030`）は reset / seed を持たず、updateBizdate で業務日付だけを進め、Day1 の残高
（1500 / 2300）に取引（-200 / +100）を反映して **1300 / 2400** を検証します。
これが「業務日付をまたぐ繰越」の表現です。

## 4. 接続情報（inventory / secret / ssh trust）

プラグインは接続情報を 3 つの仕組みから解決します。config への直書きは禁止契約です。

| 情報 | 仕組み | コマンド |
|---|---|---|
| ホスト | inventory グループ | `stfw inventory list {group}` |
| パスワード | secret（age 暗号） | `stfw secret set {host} {user} [pass]` / `show` |
| SSH ホストキー | known_hosts 登録 | `stfw ssh trust {host\|group}` |

```yaml
# config/inventory/local.yml
stfw_inventory:
  - db:  [postgres]     # DB プラグインが解決する接続先
  - api: [api]          # invokeRest が解決する接続先
```

secret は `stfw secret keygen` で鍵を作り、`stfw secret set postgres appuser <pass>` で
`{host}-{user}` をキーに暗号化保存します。DB プラグインは inventory で得たホストと
`user` から `secret show` でパスワードを取り出します。

## 5. エビデンスディレクトリ規約（compare）

`compare` は 3 つのディレクトリで動きます。

| ディレクトリ | git | 内容 |
|---|---|---|
| `expect/` | 管理する | 期待値。直下に**収集プロセスのディレクトリ名**を置き、その下は当該プロセスの `evidence/` と同型 |
| `actual/` | 生成物 | 収集プロセスの `evidence/` 配下への file-level symlink（自動生成） |
| `result/` | 生成物 | compare-files の比較結果（`CompareSummary.csv` 等） |

つまり `expect/{収集プロセス名}/{database}/{table}.csv` に期待値を置けば、compare が
同じ bizdate 内の収集エビデンスと突合します。差分の扱いは設定キー `on_mismatch` で選べます
（AS-BUILT §4.11）:

- `error`（既定）: 差分でステップ失敗として停止し、後続は `Blocked`（回帰テスト運用）
- `warn`: 差分を `Warn` として記録して最後まで実行を続け、`stfw run` は exit 3 で完走する。
  status / HTML レポートの Warn（黄）表示がそのまま「比較 NG の鳥瞰」になる（差分確認運用）

既定は行全体のテキスト比較ですが、**比較レイアウト**を定義すると項目単位の比較にできます
（連番・更新時刻列の除外、キー列による行の対応付けなど）。シナリオを跨いで共通のレイアウトは
`config/plugins/process/compare/compare_layout/*.json`（プロジェクト共通）に、プロセス固有の
上書きは各プロセスの `config/compare_layout/` に置きます（後勝ち。AS-BUILT §4.11）。
daily-balance の `transactions.json`（連番 id を Ignore・`account_id`+`bizdate` をキーに突合）が
実例です。レイアウトの書き方は compare-files の比較レイアウトリファレンスを参照してください。

## 6. ホスト操作系プラグイン（ssh 経由）

`sshExec` / `scpPut` / `collectLog` / `collectFile` は ssh/scp でリモートホストを操作します。
daily-balance の例（compose 内で完結）では扱いませんが、実ホストに対しては次のように使います。

- 事前に `stfw ssh trust {group}` で known_hosts を登録
- Arrange で `scpPut`（設定ファイル配置）、Act で `sshExec`（バッチ起動）
- Collect で `collectLog`（業務日付で絞ったログ）/ `collectFile`

## 7. 実行と確認

```sh
stfw validate {scenario}          # 静的検証（規約・プラグイン解決・config）
stfw plugin install {type}        # 外部バイナリ（k6 / compare-files 等）を取得
stfw run {scenario}               # 実行（run 開始前に validate 相当を自動実行）
stfw status [run_id]              # 実行ジャーナルの状態表示
stfw report [run_id]              # HTML レポート再生成
```

`stfw run` はプラグインの外部バイナリを自動 install しません。k6（invokeRest/invokeWeb）や
compare-files（compare）を使うシナリオは、事前に `stfw plugin install {type}` が必要です
（`stfw init` は全プラグインの install をまとめて行います）。

## 8. シナリオを文書化・雛形生成する

ここまでの `daily-balance` はディレクトリを直接編集して作りました。既存シナリオを
レビュー用に文書化したい場合や、似た構造のシナリオを別名で量産したい場合は
`stfw scenario reverse/scaffold` を使います。方式は「**tree（ディレクトリ構造）が
真実の源**・spec（構造化 YAML）は tree と可逆・doc（Markdown）は tree からの
読み取り専用の投影」です（`stfw new scenario` の対話的な単一ノード生成とは別物）。

```sh
cd examples/daily-balance/stfw

# tree -> spec + doc: リバース生成 (spec .yml + doc .md をセット出力。既定出力先 docs/)
stfw scenario reverse daily-balance
#   -> docs/daily-balance.yml  (spec: ツリーと可逆な YAML)
#   -> docs/daily-balance.md   (doc:  要求トレーサビリティ表つきのレビュー資料)
# 出力先を変えるなら -o
stfw scenario reverse daily-balance -o /tmp/out

# spec -> tree: spec からシナリオ骨格 (metadata.yml + config/config.yml) を生成 (往復の入口)
# 別名で量産する例 (spec の scenario: の値を書き換えてから使う)
stfw scenario scaffold /tmp/daily-balance-2.yml
```

`reverse` は spec（`.yml`）と doc（`.md`）を常にセットで生成します。doc は要求トレーサビリティ表
（`requirement_specifications` を全 process から集約）と業務日付ごとの process 一覧・設定を並べた
読み取り専用のレビュー資料です。`reverse`（spec）⇄ `scaffold` は往復可能ですが、対象は**骨格のみ**です。

| 対象 | 往復可否 |
|---|---|
| シナリオ名・業務日付 (seq/bizdate)・プロセス (seq/group/type)・description・requirement_specifications・config/config.yml | ✅ |
| `data/**`（CSV 等）・`scripts/**`・`expect/**`・secret・階層フック `plugins/**` | ❌（人が書く葉。`scaffold` は生成しない） |

`scaffold` は既存のシナリオディレクトリがあると既定でエラーになります（誤上書き防止）。
spec を編集した後、tree をそれに揃えたい場合は `--sync` で差分同期します。

```sh
# 既存シナリオを spec に合わせて同期する (追加 / 維持 / 削除)
stfw scenario scaffold --sync docs/daily-balance.yml
# 削除したディレクトリは `removed: ...` 行で表示される
```

`--sync` の挙動（bizdate / process の各ディレクトリ単位）:

| spec | disk | 挙動 |
|---|:---:|---|
| あり | なし | **追加** |
| あり | あり | **維持**（`metadata.yml` / `config/config.yml` は spec で上書き、`data/`・`scripts/`・`expect/` 等の葉は温存） |
| なし | あり | **削除**（実装済みの葉ごと。破壊的） |

> `--sync` は spec から消えた bizdate/process を実装済みの葉（`data/`・`scripts/`・`expect/`）ごと
> 削除する破壊的操作です。規約に合致しないディレクトリ（`notes/` 等）には触れません。

詳細は [`docs/AS-BUILT.md`](AS-BUILT.md) §12（シナリオ reverse/scaffold 投影と往復）を参照してください。

## 参考

- 実行可能サンプル: [`examples/daily-balance`](../examples/daily-balance/)
- 各プラグインの詳細: [`assets/plugins/process/{type}/README.md`](../assets/plugins/process/)
- 実装契約の集約: [`docs/AS-BUILT.md`](AS-BUILT.md)
