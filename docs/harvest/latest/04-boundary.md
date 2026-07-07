# 04 システム境界

## 解析メタ情報

| 項目 | 値 |
|------|-----|
| 解析対象リポジトリ | /Users/suwa_sh/src/github.com/scenario-test-framework/stfw |
| コミットハッシュ | ed02ba61d48212a49c416e309925bbe0ac825759 |
| 解析日 | 2026-07-07 |
| フェーズ | Phase4（システム境界: UC / 画面 / イベント / タイマー） |
| 前提 | analysis/01-overview.md, analysis/02-value.md, analysis/03-environment.md |

前提・読み方:

- アクター名は Phase2 の定義「テスト実行者 / シナリオ作成者 / 環境管理者 / テスト結果確認者」と統一する。
  外部システム名は「digdag / webhook 受信先 / テスト対象ホスト群 / dl.bintray.com」を使う（02-value.md 申し送り）。
- UC 粒度は Phase3 申し送りに従い「ユーザーの目的単位」で再編した。CLI サブコマンド 1 本 = UC 1 個とせず、
  scaffold 生成 3 コマンド（scenario -i / bizdate -i / process -i）は 1 UC に統合、
  暗号化キー生成 + パスワード登録/参照（gen-encrypt-key / passwd）も 1 UC に統合した
  （事実: 各コマンドの usage。推測: 統合の括りは「テストシナリオの構造を組み立てる」「資格情報を管理する」
  という作業目的の同一性からの判断）。
- digdag からの呼び戻し（run/scenario/bizdate の setup・teardown、process の run/dry-run）は
  Phase3 申し送りどおり UC として独立させ、digdag との連携は「イベント」として整理した
  （事実: src/bin/lib/stfw/domain/repository/dig_repository:60-88, src/template/scenario/sample/scenario.dig:3-30,
  src/template/scenario/sample/_10_99990101/bizdate.dig:3-23）。
- 情報名・状態名は Phase5 確定前の仮置きである（phase4-boundary.md の規則）。
- 本システムに GUI は無い。アクターとのインターフェースは CLI（ターミナル）であり、
  RDRA の「画面」は CLI・digdag Web UI・ログ出力を境界面として整理した
  （事実: src/bin/stfw:65-87 の usage 出力。01-overview.md「フロントエンド: 専用 UI なし」）。

## UC 一覧

| UC | 目的 | 操作する情報（仮置き） | 遷移させる状態（仮置き） | 対応 BUC | 確度 | 根拠 |
|----|------|----------------------|------------------------|---------|------|------|
| UC01 stfw をインストールする | tar.gz 展開後、install 実行で依存モジュール（digdag jar 等）を取得し実行基盤を整える | stfw 本体、依存モジュール | - | stfw を導入する | high | 事実: src/bin/install:1-7（scenario test framework installer）,100-144（依存ダウンロード）, docs/adoc/install/install_index.adoc:49-52, build/product/integration_test.sh:46-61 |
| UC02 プロジェクトを初期化する | テンプレート一式（stfw.yml / config / sample シナリオ）を展開してプロジェクトを開始する | プロジェクト、プロジェクト設定 | - | プロジェクトを初期化する | high | 事実: src/bin/cmd/init:25,70, src/template/（stfw.yml, config/, scenario/sample/）, build/product/integration_test.sh:70-83 |
| UC03 資格情報を管理する | 暗号化キーペアを生成し、ホスト×ユーザー単位のパスワードを暗号化登録・復号参照する | 暗号化キー、パスワード | - | テスト対象ホストの接続情報を管理する | high | 事実: src/bin/cmd/gen-encrypt-key:25,80, src/bin/cmd/passwd:31-36,88-104（generate / --show）, src/bin/lib/commons/bash_utils:255-299, build/product/integration_test.sh:77-102 |
| UC04 テスト対象ホスト情報を参照する | inventory のホストグループ存在確認・ホスト一覧取得を行う（定義の編集はシステム外の手作業） | インベントリ | - | テスト対象ホストの接続情報を管理する | high | 事実: src/bin/cmd/inventory:34-35（--is-exist / --list）, src/template/config/inventory/staging.yml:1-8, build/product/integration_test.sh:111-125 |
| UC05 シナリオ構造を組み立てる | scenario > bizdate > process の 3 階層 scaffold を生成してテストシナリオの骨格を作る | シナリオ、業務日付、プロセス | - | テストシナリオを作成する | high | 事実: src/bin/cmd/scenario:34（-i <scenario-name>）, src/bin/cmd/bizdate:34（-i <seq> <bizdate>）, src/bin/cmd/process:36（-i <seq> <group> <process-type>）, build/product/integration_test.sh:135-162。統合粒度は 推測: 3 コマンドが同一目的（構造組み立て）の連続作業である judgement（03-environment.md 申し送り対応） |
| UC06 ワークフロー定義を生成する | ディレクトリ構造から dig 定義（scenario.dig / bizdate.dig）を自動生成する（cascade 生成含む） | ワークフロー定義（dig） | - | ワークフロー定義を生成・検証する | high | 事実: src/bin/cmd/scenario:35-36（-g / -G）, src/bin/cmd/bizdate:35（-g）, src/bin/lib/stfw/domain/repository/dig_repository:96-120（generate_scenario）, build/product/integration_test.sh:165-168 |
| UC07 プロセスプラグインを管理する | プロセスプラグインの一覧確認と依存モジュールのインストールを行う | プラグイン | - | プロセスプラグインを拡張する | high | 事実: src/bin/cmd/process:34-35（-l / -I）,141-144（デフォルトコマンド list）, src/plugins/process/{__common,scripts}, src/plugins/process/scripts/bin/install/install |
| UC08 実行基盤（server）を制御する | digdag server の起動・停止・再起動・稼働確認を行う（bind/port/状態 DB/スレッド数の指定可） | サーバプロセス（pid）、状態 DB | server 稼働状態: 停止中 ⇔ 起動中 | 実行基盤（server）を制御する | high | 事実: src/bin/cmd/server:31-38,117-146, src/bin/lib/stfw/domain/gateway/digdag_gateway:337-381（server.start: nohup 起動 + pid 保存 + 起動待機）,402-417（stop: SIGTERM）,439-459（is_running）, build/product/integration_test.sh:192-212 |
| UC09 シナリオを実行する | 指定シナリオ群を digdag ワークフローとして一括自動実行する（dry-run 検証を含む）。run_id 発行 → digdag プロジェクト push → 実行開始 → attempt_id 保存 | 実行（run_id / attempt_id）、ワークフロー定義、実行コンテキスト | 実行ステータス: Started → Success / Error | シナリオを一括自動実行する（dry-run はワークフロー定義を生成・検証する） | high | 事実: src/bin/cmd/run:31-34（<scenario-names...> / -d, --dry-run）,113-117, src/bin/lib/stfw/domain/service/run_service:18-60, src/bin/lib/stfw/domain/repository/run_repository:13-55, build/product/integration_test.sh:201 |
| UC10 実行ログを追従する | 実行中 attempt のログを終了までリアルタイム表示し、最終 state を判定する | 実行ログ、実行（attempt_id） | - | 実行結果を監視・確認する | high | 事実: src/bin/cmd/run:35（-f, --follow）, src/bin/lib/stfw/domain/service/run_service:63-81, src/bin/lib/stfw/domain/gateway/digdag_gateway:261-287（digdag log --follow）,308-316（get_state） |
| UC11 階層 setup / teardown を実行する（run・scenario・bizdate） | ワークフローの各階層の前処理・後処理（処理時間記録・webhook 通知起動・プロジェクト共通スクリプト実行）を行う。digdag からの呼び戻しが主経路（run 階層のみ stfw run -s / -t で手動実行も可） | 実行コンテキスト、処理時間、webhook payload | 各階層の実行ステータス: Started → Success / Error（teardown 時に stfw_run_status で確定） | シナリオを一括自動実行する | high | 事実: src/bin/lib/stfw/domain/repository/dig_repository:60-61,80-88（run.dig が sh>: stfw run --setup/--teardown を生成）, src/template/scenario/sample/scenario.dig:3-4,22-30（sh>: stfw scenario --setup/--teardown。_error 時 stfw_run_status=Error）, 同 _10_99990101/bizdate.dig:3-4,15-23, src/bin/cmd/run:36-37, src/bin/cmd/scenario:37-38, src/bin/cmd/bizdate:36-37, src/plugins/{run,scenario,bizdate}/__common/setup/10_webhook_start・teardown/90_webhook_end |
| UC12 プロセスを実行する | 1 プロセスを setup → pre_execute → execute → post_execute → teardown の順で実行する。scripts プラグインでは scripts/ 配下をファイル名昇順に逐次実行し、エラー時は後続を Blocked にして停止する。dry-run は setup / pre_execute / teardown のみ | プロセス、スクリプト、ステップ実行結果、webhook payload | プロセス・ステップの実行ステータス: Pending → Started → Success / Error / Blocked | シナリオを一括自動実行する | high | 事実: src/bin/cmd/process:37-38（-d / -r）, src/bin/lib/stfw/domain/service/process_service:77-122（dry_run）,124-186（run: setup→pre→execute→post→teardown）, src/template/scenario/sample/_10_99990101/bizdate.dig:13（sh>: stfw process ${run_mode}）, src/plugins/process/scripts/bin/run/execute:13（bulk_exec_scripts）, src/plugins/process/scripts/bin/lib/common:61,134-157（Pending/Blocked/Success/Error 判定）, src/plugins/process/scripts/README.adoc:202-213 |
| UC13 実行状況を通知する | run/scenario/bizdate/process の各階層の start / end（Success・Error）時に、payload（id・status・時刻・処理時間・digdag URL 等）を webhook 受信先へ非同期 HTTP POST する。on_start / on_success / on_error 設定で抑制可 | webhook payload、実行ステータス | -（ステータスを参照して payload に記録） | 実行結果を監視・確認する | high | 事実: src/plugins/{run,scenario,bizdate,process}/__common/setup/10_webhook_start・teardown/90_webhook_end, src/bin/lib/stfw/adapter/cli/webhook_controller:11-138, src/bin/lib/stfw/domain/repository/webhook_repository:11-89（status=Started / ${stfw_run_status}）, src/bin/lib/stfw/domain/gateway/webhook_gateway:30-44（非同期）,68-116（curl POST）, src/bin/lib/stfw/domain/service/spec/webhook_spec:59-77（on_start/on_success/on_error 判定）, src/config/webhook/payload.yml:1-21 |
| UC14 digdag を直接操作する | digdag CLI を stfw 環境設定（PATH_DIGDAG 等）込みでラップ実行する（トラブルシュート・詳細操作用の補助コマンド） | digdag 管理情報全般 | - | 実行基盤（server）を制御する（補助） | medium | 事実: src/bin/cmd/digdag:24（execute digdag command）,55（${PATH_DIGDAG} "$@"）。推測: 「トラブルシュート・詳細操作用」という位置づけは usage に説明が無く、wrapper 実装であることからの導出。BUC への帰属も仮置き |
| UC15 SSH サーバキーを登録する（候補・未確定） | テスト対象ホストのサーバキーを known_hosts へ登録する（ssh-keygen -R + ssh-keyscan） | SSH known_hosts | - | テスト対象ホストの接続情報を管理する（仮置き） | low | 推測: 関数 gen_ssh_server_key は実装済み（事実: src/bin/lib/commons/bash_utils:131-164）だが CLI コマンド・dig・プラグインからの呼び出し箇所がリポジトリに無い。ユーザースクリプトからの利用想定と判断（03-environment.md FIXME の引き継ぎ）。UC として採用するかは Phase3 ユーザー確認に委ねる |

### Phase3 UC 候補との対応（粒度再編の記録）

| Phase3 の UC 候補（03-environment.md） | 本書の UC | 再編内容 |
|----------------------------------------|-----------|---------|
| インストールする | UC01 | そのまま |
| プロジェクトを初期化する | UC02 | そのまま |
| 暗号化キーを生成する / パスワードを暗号化登録する / パスワードを参照する | UC03 | 目的単位（資格情報の管理）で 3 → 1 に統合 |
| inventory を参照する | UC04 | そのまま |
| シナリオ scaffold / 業務日付 scaffold / プロセス scaffold を生成する | UC05 | 目的単位（構造組み立て）で 3 → 1 に統合 |
| ワークフロー定義を生成する / シナリオを dry-run する | UC06 / UC09 | dry-run は UC09 のオプション（run_mode=--dry-run）として吸収（事実: src/bin/cmd/run:50,77-81 で同一コマンドのモード切替） |
| プラグインを一覧する / プラグイン依存をインストールする | UC07 | 2 → 1 に統合 |
| server を起動する / 状態を確認する / 停止する | UC08 | 3 → 1 に統合 |
| シナリオを実行する | UC09 | そのまま |
| 実行ログを追従する | UC10 | そのまま |
| run setup / teardown、scenario/bizdate/process の setup・execute・teardown（Phase4 で UC 化） | UC11 / UC12 | digdag 呼び戻しを「階層 setup/teardown」（UC11）と「プロセス実行」（UC12）の 2 UC に整理 |
| 結果を通知する（Phase4 で UC 化） | UC13 | webhook 通知を独立 UC 化 |
| SSH サーバキーを登録する（呼出経路未確定） | UC15 | low のまま保留 |

## 画面一覧

GUI は存在しないため、アクターとの境界面（CLI・外部 UI・ログ出力）を「画面」相当として記録する。

| 画面 | 説明 | アクター | 関連 UC | 確度 | 根拠 |
|------|------|---------|--------|------|------|
| stfw CLI（ターミナル） | `stfw <command>` 形式のサブコマンド体系。--help で全コマンドの usage を動的一覧表示。全操作の入口 | テスト実行者、シナリオ作成者、環境管理者 | UC01〜UC12, UC14 | high | 事実: src/bin/stfw:65-87（usage: commands 一覧を各コマンドの --description から動的生成）, src/bin/cmd/ 配下 10 コマンド。アクター割当は 02-value.md のアクター定義に従う |
| digdag Web UI | digdag server 同梱の Web インターフェース。stfw は attempt URL（http://{ip}:{port}/attempts/{attempt_id}）を案内し、実行状況・詳細の確認に使う | テスト結果確認者 | UC09 の結果閲覧（システム外操作）、UC08（port 設定） | medium | 事実: src/bin/lib/stfw/domain/service/run_service:135-139（attempt URL 生成）, src/bin/cmd/server:35（web interface and api clients）, docs/adoc/install/install_index.adoc:67-68。推測: UI 画面自体は digdag 側の資産で、stfw リポジトリに画面定義は無い。確認手順のドキュメントも無い |
| コンソール / ログファイル出力 | 実行ログの表示面。terminal 実行時はカラー出力、ログファイルはシークレットマスキング済み。--follow 時は attempt ログを追従表示 | テスト結果確認者（テスト実行者が兼務） | UC10、全 UC のログ確認 | medium | 事実: src/bin/lib/setenv:144-151（PATH_LOG / log.mask）, git log 935294e（terminal カラー出力）, src/bin/lib/stfw/domain/gateway/digdag_gateway:261-287。推測: 出力専用の境界面を「画面」として扱う分類は本 Phase の判断（入力機能は無い） |

## イベント一覧

外部システムとの連携（入出力）を整理する。方向は「stfw → 外部」「外部 → stfw」で記す。

| イベント | 説明 | 方向 | 外部システム | 関連 UC | 確度 | 根拠 |
|---------|------|------|-------------|--------|------|------|
| ワークフロー登録・実行指示 | run_id 単位の digdag プロジェクトを push し、run.dig を start する（endpoint: localhost:{port}） | stfw → 外部 | digdag | UC09 | high | 事実: src/bin/lib/stfw/domain/gateway/digdag_gateway:123-147（digdag push --endpoint）,170-183（digdag start --session now）, src/bin/lib/stfw/domain/repository/run_repository:13-55 |
| ワークフロー呼び戻し（sh> オペレータ） | digdag がタスク実行時に stfw CLI を呼び戻す: `stfw run --setup/--teardown`（run.dig）、`stfw scenario --setup/--teardown`（scenario.dig）、`stfw bizdate --setup/--teardown` と `stfw process ${run_mode}`（bizdate.dig）。エラー時は _error タスクで teardown（status=Error）を呼ぶ | 外部 → stfw | digdag | UC11, UC12 | high | 事実: src/bin/lib/stfw/domain/repository/dig_repository:60-88, src/template/scenario/sample/scenario.dig:3-30, 同 _10_99990101/bizdate.dig:3-23, docs/design/architecture/app_arch.puml:39（digdag -up-> stfw） |
| 実行状況照会 | attempt ログの追従取得（digdag log --follow）、タスク state 取得（digdag tasks）、起動ヘルスチェック（GET /api/projects） | stfw → 外部 | digdag | UC10, UC08 | high | 事実: src/bin/lib/stfw/domain/gateway/digdag_gateway:261-287,308-316,557-577 |
| server プロセス起動・停止 | digdag server プロセスの nohup 起動（bind/port/db_mode/max-task-threads/ログ出力先を指定）・SIGTERM 停止・pid 管理 | stfw → 外部 | digdag | UC08 | high | 事実: src/bin/lib/stfw/domain/gateway/digdag_gateway:337-381,402-417, src/template/stfw.yml:25-33 |
| webhook 通知（start / end） | run/scenario/bizdate/process の各階層の開始・終了時に、payload yml を JSON 変換して設定済み URL 群へ非同期 HTTP POST。end 時 status は Success / Error。on_start / on_success / on_error で ON/OFF 可 | stfw → 外部 | webhook 受信先 | UC13（発火元は UC11, UC12） | high | 事実: src/bin/lib/stfw/domain/gateway/webhook_gateway:30-44,68-116（curl --request POST --data-binary @payload.json）, src/bin/lib/stfw/domain/repository/webhook_repository:11-89, src/template/stfw.yml:11-22, src/config/webhook/payload.yml:1-21, src/config/webhook/{run,scenario,bizdate,process}.yml |
| 依存モジュールダウンロード | install 時に digdag jar 等の依存モジュールを配布元から curl 取得する | stfw → 外部 | dl.bintray.com（digdag 配布元） | UC01 | high | 事実: src/bin/lib/setenv:103（URL_DIGDAG）, src/bin/install:100-144（private.download / curl）。FIXME 参照（配布元 URL の現行有効性） |
| テスト対象ホストへの操作 | フレームワーク本体は SSH 実行コードを持たず、テスト対象ホストへの操作はユーザースクリプト（scripts/ 配下）に委ねられる。イベントとしての定型 I/F は無い | stfw → 外部（ユーザースクリプト経由） | テスト対象ホスト群 | UC12（スクリプト実行の中で） | medium | 事実: src/plugins/process/scripts/README.adoc:244（実行ホストで利用できる全ての言語を実行できます）。推測: 02-value.md の補足どおり、リモート操作の実装はユーザー側の責務で、stfw 側にプロトコル定義は無い |

## タイマー一覧

| タイマー | タイミング | 起動 UC | 確度 | 根拠 |
|---------|-----------|--------|------|------|
| （該当なし） | - | - | medium | 事実: 生成される run.dig / scenario.dig / bizdate.dig のいずれにも digdag の schedule 定義が無く（src/bin/lib/stfw/domain/repository/dig_repository:52-89, src/template/scenario/sample/scenario.dig:1-30）、cron・定期実行の設定もリポジトリに存在しない。推測: 「定期実行は想定外か未実装」という解釈は 03-environment.md 申し送りを踏襲。シナリオ実行の起動は常に人手（stfw run） |

- 補足: ログローテーション（日次判定）は stfw 起動時の事後処理として実行されるもので、
  定時起動のタイマーではない（事実: src/bin/stfw:167-168 `log.rotatelog_by_day_first`。
  推測: ビジネス上の時間的区切りではなく内部運用処理のため、RDRA タイマーには含めない）。
- 補足: bizdate（業務日付）は「テストデータとしての日付ラベル」であり、実時刻に連動した起動条件ではない
  （事実: src/template/scenario/sample/ の bizdate=99990101 というダミー日付, src/bin/cmd/bizdate:34。
  推測: タイマーではなく Phase5 の情報・バリエーションとして扱うべきと判断）。

## FIXME / 申し送り

- FIXME: `gen_ssh_server_key`（src/bin/lib/commons/bash_utils:131-164）は定義済み・未参照のまま。
  UC15 として low で仮置めした。採用可否（ユーザースクリプトからの利用想定か、廃止か）を
  Phase3 ユーザー確認で決めること（03-environment.md からの継続 FIXME）。
- FIXME: 依存モジュール配布元 dl.bintray.com（src/bin/lib/setenv:103）は Bintray サービス終了により
  現在は到達不能の可能性が高い（推測: 一般知識。リポジトリ内に代替 URL・ミラーの記述なし）。
  UC01 のイベント「依存モジュールダウンロード」は as-is 記録であり、現行動作の保証はない。
- FIXME: エラー時の再実行・途中再開（resume）の専用 UC は存在しない。digdag 自体は resume 機能を持つが、
  stfw の CLI にリラン用 I/F が無く、再実行は UC09（シナリオを実行する）のやり直しになる
  （事実: src/bin/cmd/run:31-38 に resume 系オプションなし。03-environment.md low 項目 3 の境界レイヤーでの帰結）。
- 申し送り（Phase5 へ）: 状態モデルの一次証拠は次の 2 系統。
  (1) プロセス・ステップの実行ステータス Pending / Started / Success / Error / Blocked
  （事実: src/bin/lib/setenv:26-31 の定数, src/plugins/process/scripts/bin/lib/common:134-157 の遷移判定）。
  (2) 階層（run/scenario/bizdate）の実行ステータス: webhook payload の status として
  Started →（stfw_run_status 経由で）Success / Error が確定する
  （事実: src/bin/lib/stfw/domain/repository/webhook_repository:17,38, dig の _error / teardown ブロック）。
- 申し送り（Phase5 へ）: 情報の仮置き名「プロジェクト / シナリオ / 業務日付 / プロセス / スクリプト /
  ワークフロー定義（dig） / 実行（run_id・attempt_id） / 実行コンテキスト / インベントリ / 暗号化キー /
  パスワード / webhook payload / 実行ログ / プラグイン」を本書の UC 表で使用した。Phase5 で整合を取ること。
- 申し送り: webhook payload は run > scenario > bizdate > process の親子関係（id / parent_id / type）を持つ
  （事実: src/config/webhook/payload.yml:2-4, src/config/webhook/process.yml:1-8 のコメントアウトされた階層構造）。
  Phase5 の情報モデル（実行結果ツリー）の手がかりになる。

## confidence: low 項目一覧（ユーザー確認対象）

| # | 項目 | 内容 | 手がかり |
|---|------|------|---------|
| 1 | UC15「SSH サーバキーを登録する」 | UC としての採用可否と、BUC「テスト対象ホストの接続情報を管理する」への帰属 | 推測: 関数実装（src/bin/lib/commons/bash_utils:131-164）はあるが呼び出し箇所ゼロ。ユーザースクリプトからの利用想定と判断（03-environment.md low 項目 2 の引き継ぎ） |
