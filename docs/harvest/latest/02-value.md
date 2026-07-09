# 02 システム価値

## 解析メタ情報

| 項目 | 値 |
|------|-----|
| 解析対象リポジトリ | /Users/suwa_sh/src/github.com/scenario-test-framework/stfw |
| コミットハッシュ | ed02ba61d48212a49c416e309925bbe0ac825759 |
| 解析日 | 2026-07-07 |
| フェーズ | Phase2（システム価値: アクター / 外部システム / 要求） |
| 前提 | analysis/01-overview.md |

前提: ロール定義・認証・認可コードはリポジトリに存在しない。CLI 実行者は単一種別であり、
アクターの区分は「作業の目的」による論理区分である（推測: src/bin/cmd/ 全コマンドが同一権限で実行され、
ログイン・ロール切替の仕組みが無い）。

## アクター一覧

| アクター群 | アクター | 説明 | 確度 | 根拠 |
|-----------|---------|------|------|------|
| テスト担当者 | テスト実行者 | stfw CLI でプロジェクト初期化（init）・シナリオ実行（run）・サーバ制御（server）を行う利用者。docs/design では単一の `actor user` として定義 | high | 事実: docs/design/architecture/app_arch.puml:5,38（actor user → stfw）, src/bin/cmd/init:25, src/bin/cmd/run:25 |
| テスト担当者 | シナリオ作成者 | scenario/bizdate/process の scaffold 生成とスクリプト配置でテストシナリオを記述する利用者。実行者と同一人物の可能性が高い（ロール分離なし） | medium | 事実: src/bin/cmd/scenario:34-38（--init/--generate-dig）, src/bin/cmd/bizdate:34-37, src/bin/cmd/process:34-38。推測: 実行者との区分は CLI サブコマンドの目的の違いからの論理区分。権限上の分離は無い |
| テスト担当者 | 環境管理者 | inventory（テスト対象ホストのグループ定義）の管理と、暗号化キー生成・パスワード暗号化登録を行う利用者 | medium | 事実: src/bin/cmd/inventory:25（read inventory settings）, src/bin/cmd/passwd:25,31（generate encrypted passwd file, `<host> <user> <password>`）, src/bin/cmd/gen-encrypt-key:25。推測: 専用ロールは未定義で、テスト実行者が兼務する想定と判断 |
| テスト担当者 | テスト結果確認者 | digdag server の Web UI / API または webhook 通知先で実行結果・進捗を確認する利用者 | medium | 事実: src/bin/cmd/server:35（"port number to listen for web interface and api clients"）。推測: UI 閲覧者の役割定義は無く、server コマンドの usage 記述から想定 |
| フレームワーク保守者 | stfw 開発者・コントリビューター | フレームワーク自体を開発・保守するコミュニティ（システムの利用者ではなくステークホルダー寄り。Phase3 以降の BUC には含めない想定） | high | 事実: .github/CONTRIBUTING.md, CODE_OF_CONDUCT.md, .travis.yml:6-9（gitter 通知） |

- 補足: バッチ・ワークフローエンジン（digdag）からの起動（digdag → stfw の逆方向呼び出し）はアクターではなく
  Phase4 のタイマー/外部システム連携として扱う（事実: docs/design/architecture/app_arch.puml:39 `digdag -up-> stfw`）。

## 外部システム一覧

| 外部システム群 | 外部システム | 連携内容 | 確度 | 根拠 |
|---------------|-------------|---------|------|------|
| ワークフローエンジン | digdag（0.9.24） | シナリオ実行の委譲先。同梱 jar を server モードで起動し、run.dig/scenario.dig/bizdate.dig を実行。バージョン・URL を gateway 経由で取得 | high | 事実: docs/design/architecture/app_arch.puml:7,80（gateway → digdag）, src/bin/lib/stfw/domain/gateway/digdag_gateway:31-42, src/bin/lib/setenv:89-91,103-104 |
| 通知受信システム | webhook 受信先（CI / チャット / 監視等の任意 HTTP エンドポイント） | プロセス実行の start/success/error 時に JSON payload を HTTP POST。URL は stfw.yml に複数指定可（環境変数参照も可） | high | 事実: src/bin/lib/stfw/domain/gateway/webhook_gateway:87-95（curl POST）, src/template/stfw.yml:11-22, src/config/webhook/payload.yml:1-21 |
| 通知受信システム | webhook.site（動作確認用テストサイト） | webhook 動作確認用の外部サービス（テンプレートにコメントアウトで案内） | high | 事実: src/template/stfw.yml:17-19 |
| テスト対象システム | テスト対象ホスト群（web / ap / db） | inventory でグループ管理されるテスト適用先ホスト。SSH known_hosts へのサーバキー登録関数と、ホスト×ユーザー単位の暗号化パスワード管理を持つ | high | 事実: src/template/config/inventory/staging.yml:1-8, src/bin/cmd/inventory:34-35, src/bin/lib/commons/bash_utils:131-164（gen_ssh_server_key: ssh-keygen -R / ssh-keyscan）, src/bin/cmd/passwd:31 |
| モジュール配布元 | dl.bintray.com（digdag jar 配布元） | インストール時の依存モジュールダウンロード元 | high | 事実: src/bin/lib/setenv:103（URL_DIGDAG）, src/bin/install:100-144（private.download / curl） |
| 開発・CI 基盤（実行時システムコンテキスト外） | Travis CI / gitter / GitHub（gh-pages） | フレームワーク自体の CI・ビルド通知・ドキュメント公開。stfw 利用時のシステム連携ではないため、RDRA システムコンテキストからは除外候補 | high | 事実: .travis.yml:6-9,36-47, build/docs/publish.sh, build/env.properties:8,12 |

- 補足: テスト対象ホストへの実際のリモート実行手段（ssh 実行コマンド等）は scripts plugin の
  ユーザースクリプト側に委ねられており、フレームワーク本体には ssh 実行コードが無い
  （事実: src/plugins/process/scripts/README.adoc:244 "実行ホストで利用できる全ての言語を実行できます"。
  推測: git log の PR #7 ブランチ名 feature/remote_process と gen_ssh_server_key / passwd 群から、
  リモートホスト操作はユーザースクリプトが担う設計と判断）。

## 機能要求

as-is 要求（現行システムがすでに実現している価値）として記述する。

| 分類 | アクター | 機能要求 | 説明 | 確度 | 根拠 |
|------|---------|---------|------|------|------|
| シナリオ自動実行 | テスト実行者 | ディレクトリ構造で記述したシナリオテストをワークフローとして一括自動実行できること | scenario/{scenario}/{bizdate}/{seq}_{group}_{type}/ の構造から dig 定義を生成し、digdag で実行。dry-run・setup/teardown・ログ追従（--follow）も可能。理由（なぜ）: 初回コミットの機能名 "run workflow according to directory structure" が唯一の一次記述 | high | 事実: git log 044a51e / CHANGELOG.md:7, src/bin/cmd/run:25,32-38, src/bin/cmd/scenario:34-38 |
| 業務日付管理 | シナリオ作成者 | 業務日付（bizdate）をまたぐ一連の業務処理を、シナリオ > 業務日付 > プロセスの階層で管理・逐次実行できること | bizdate 単位の scaffold 生成（YYYYMMDD）と、プロセス内スクリプトのファイル名昇順逐次実行・エラー時停止。理由: バッチ処理系業務システムのテストでは業務日付の進行が本質という推測（明文ドキュメントなし） | high（理由は low） | 事実: src/bin/cmd/bizdate:25,34（bizdate format: YYYYMMDD）, src/plugins/process/scripts/README.adoc:19,212-213。理由の根拠は 推測: ドメイン一般論（01-overview.md のドメイン特性より） |
| 結果通知 | テスト結果確認者 | プロセス実行結果（開始/成功/失敗、各ステップの結果・所要時間）を webhook で外部システムへ通知できること | on_start/on_success/on_error を設定で ON/OFF 可。payload に実行ホスト・ユーザー・digdag URL・ステップ別結果を含む。理由: コミットメッセージは「webhook機能を追加」のみで目的の明文なし。実行状況の外部監視・可視化のためと推測 | high（理由は low） | 事実: git log c01f746（PR #6 feature/add_webhook）, src/template/stfw.yml:20-22, src/config/webhook/payload.yml:1-21, src/plugins/process/scripts/README.adoc:31-129。理由の根拠は 推測: payload 内容（進捗・所要時間）からの導出 |
| マルチホスト・秘匿情報管理 | 環境管理者 | 環境別 inventory によるテスト対象ホストのグループ管理と、資格情報（パスワード）の暗号化保管ができること | inventory ファイル（staging.yml 等）の切替、グループ存在確認・ホスト一覧取得、RSA+S/MIME(AES256) によるパスワード暗号化・復号。理由: リモートホストへのプロセス適用（feature/remote_process）に伴い資格情報の平文保管を避けるためと推測 | high（理由は low） | 事実: git log 20a8812（PR #7 feature/remote_process）, 111d5af・15c881b（PR #5 feature/passwd）, src/bin/cmd/inventory:34-35, src/bin/cmd/passwd:31, src/bin/lib/commons/bash_utils:255-299。理由の根拠は 推測: コミット順序（passwd → remote_process）と機能の組み合わせからの導出 |
| プラグイン拡張 | シナリオ作成者 | プロセスタイプをプラグインとして追加・拡張できること | process plugin の一覧（--list）・依存インストール（--install）・scaffold 生成（--init <process-type>）を提供。現状の同梱プラグインは scripts のみ。理由: 「プロセスを追加できるように、既存のプロセスを整理」というコミットが直接の証拠 | high | 事実: src/bin/cmd/process:34-38, src/plugins/process/{__common,scripts}, git log e811c59（"プロセスを追加できるように、既存のプロセスを整理"） |

## 非機能要求

読めた範囲の as-is 記録（後段 quality-attributes の入力）。

| 分類 | 非機能要求 | 説明 | 検証方法 | 確度 | 根拠 |
|------|-----------|------|---------|------|------|
| 信頼性（再現性・順序性） | スクリプトはファイル名昇順で逐次実行し、途中エラーで後続を実行せずエラー終了すること | テストの再現性・順序保証。ステータス Blocked で未実行を表現 | scripts plugin の仕様記述 + IT（build/product/integration_test.sh） | high | 事実: src/plugins/process/scripts/README.adoc:212-213,43, src/bin/lib/setenv:27-31 |
| 安全性（事前検証） | dry-run モードで実タスクを実行せずにワークフローを検証できること | run / process の両方に dry-run オプションあり | CLI オプションの実行確認（UT: test/ut/） | high | 事実: src/bin/cmd/run:34（-d, --dry-run "doesn't execute tasks"）, src/bin/cmd/process:37 |
| セキュリティ | パスワードは暗号化保管（RSA 2048 + S/MIME AES256）し、ログ出力時はシークレットをマスキングすること | 平文パスワードのファイル・ログ露出防止 | UT（test/ut/, git log cc33930 "gen-encrypt-key: 引数なしの動作確認"）+ 配布物からの鍵除去（git log b0b5bca） | high | 事実: src/bin/lib/commons/bash_utils:255-299, src/bin/lib/setenv:146-151（log.mask で PASSWORD/TOKEN を [secret] 置換） |
| セキュリティ（ホスト認証） | テスト対象ホストの SSH サーバキーを known_hosts へ自動登録できること | ssh-keygen -R による旧キー削除 + ssh-keyscan による再登録 | 関数の UT / 実機確認 | high | 事実: src/bin/lib/commons/bash_utils:131-164 |
| 運用性（監視・ログ） | ログレベル（trace〜error）を設定で変更でき、ログはファイルへ集約、terminal 実行時はカラー出力すること | 障害調査・実行状況の把握 | 設定変更しての目視確認（UT: test/ut/） | high | 事実: src/config/stfw.yml:2-3, src/bin/lib/setenv:144（PATH_LOG）, git log 935294e（"terminalで実行されている場合、ログをカラーで出力"） |
| 性能 | digdag のタスク実行スレッド数を上限設定（デフォルト 64）できること | 実行並列度の制御。64 という値の選定理由は読み取れない | 設定値の確認（server start --max-task-threads） | medium | 事実: src/template/stfw.yml:32-33, src/bin/cmd/server:38。推測: 値の根拠・性能目標(SLA)の記述はどこにも無い |
| 可用性（状態保持） | digdag の状態 DB をメモリ / ファイル永続化から選択でき、永続化時は実行状態を再起動後も保持できること | プロジェクトテンプレートのデフォルトはファイル永続化（.stfw/db） | server 再起動後の状態確認 | medium | 事実: src/config/stfw.yml:11-12（--memory）, src/template/stfw.yml:30-31（--database）。推測: 「再起動後も状態保持」という目的は digdag の一般仕様からの導出で、stfw 側に明文なし |
| 可搬性 | Linux / macOS 上で、tar.gz 展開のみで動作するスタンドアロン CLI であること | cygwin 対応はコメントアウト（未対応） | IT（Travis CI 上の integration_test）+ macOS 分岐の目視確認 | high | 事実: src/bin/lib/setenv:162-177（mac 分岐, cygwin コメントアウト）, docker-compose-kcov.yml:9（stfw-with-depends-*.tar.gz） |
| 国際化（タイムゾーン） | タイムゾーンを設定でき、デフォルトは Asia/Tokyo であること | 業務日付（bizdate）判定の正確性に影響。git log にタイムゾーン起因の不具合修正あり | 設定値確認 + webhook payload の時刻表記確認 | high | 事実: src/config/stfw.yml:15-16, git log c520754（"fix(webhook): タイムゾーン、ファイル名調整"） |
| 保守性（品質保証） | フレームワーク自体が UT（shunit2）・IT（serverspec）・カバレッジ計測（kcov）・静的解析（shellcheck）で継続的に検証されていること | 利用者向け要求ではなく開発プロセス上の要求 | CI（.travis.yml のイベント別スクリプト）で自動実行 | high | 事実: test/ut/ut_all.sh, git log 7992aeb（serverspec）, d268216（kcov）, 6fffa30（"clean up shellcheck report"）, .travis.yml:36-47 |
| 拡張性（設定） | プロジェクト設定（stfw.yml）はデフォルト → プロジェクトの順に読込・上書きされ、環境変数として全スクリプトに公開されること | webhook URL の環境変数参照（${URL_WEBHOOK}）等の柔軟な設定を実現 | UT + 設定上書きの動作確認 | high | 事実: src/bin/stfw:42-49（01-overview.md より）, src/template/stfw.yml:16, src/plugins/process/scripts/README.adoc:228-231 |

## FIXME / 申し送り

- FIXME: 要求の「理由（なぜ）」の一次記述がほぼ存在しない。docs/adoc 配下の overview/concept が
  placeholder（"AAA/BBB/CCC"）のままであり（事実: docs/adoc/overview/overview_index.adoc:8-10、01-overview.md 参照）、
  コミットメッセージも「〜を追加」形式で目的を記述していない。機能要求 3 件の理由を `low`（推測）とした。
  Phase3 のユーザー確認で理由の妥当性を必ず確認すること。
- FIXME: 非機能の目標値（性能 SLA・可用性目標）は一切記述が無い。max_task_threads=64 等の設定値のみが
  手がかりであり、後段 quality-attributes では「目標値未定義」を前提に補完が必要。
- 申し送り（Phase3 以降で使う名前の統一）: アクター名は「テスト実行者 / シナリオ作成者 / 環境管理者 /
  テスト結果確認者」、外部システム名は「digdag / webhook 受信先 / テスト対象ホスト群」を正とする。
- 申し送り: digdag → stfw の逆方向呼び出し（app_arch.puml:39）は Phase4（システム動作: イベント・タイマー）で
  ワークフロー起点として整理すること。

## confidence: low 項目一覧（Phase3 確認対象）

| # | 項目 | 内容 | 手がかり |
|---|------|------|---------|
| 1 | 機能要求「業務日付管理」の理由 | バッチ処理系業務システムのテストでは業務日付の進行が本質、という理由付け | 推測: ドメイン一般論。bizdate 構造の存在は high だが「なぜ」の明文なし |
| 2 | 機能要求「結果通知」の理由 | 実行状況の外部監視・可視化のため、という理由付け | 推測: payload 内容（進捗・所要時間・digdag URL）からの導出。コミットは「webhook機能を追加」のみ |
| 3 | 機能要求「マルチホスト・秘匿情報管理」の理由 | リモートホスト適用に伴う資格情報の平文保管回避、という理由付け | 推測: PR #5（passwd）→ PR #7（remote_process）のコミット順序と機能の組み合わせ |
