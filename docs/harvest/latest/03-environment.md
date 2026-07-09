# 03 システム外部環境

## 解析メタ情報

| 項目 | 値 |
|------|-----|
| 解析対象リポジトリ | /Users/suwa_sh/src/github.com/scenario-test-framework/stfw |
| コミットハッシュ | ed02ba61d48212a49c416e309925bbe0ac825759 |
| 解析日 | 2026-07-07 |
| フェーズ | Phase3（システム外部環境: 業務 / BUC / アクティビティ） |
| 前提 | analysis/01-overview.md, analysis/02-value.md |

前提・読み方:

- stfw は「シナリオテスト実行基盤」のフレームワークであり、テスト対象システム側の業務（銀行・EC 等）は
  リポジトリに現れない。本書の「業務」は **stfw を使ったシナリオテスト業務のライフサイクル** を指す
  （事実: docs/adoc/manual/manual_index.adoc:3-7 の章立てが「server制御 / シナリオ作成 / シナリオ実行」と
  利用ライフサイクル単位。推測: テスト対象業務は環境外として除外）。
- 業務のフロー全体（準備 → 作成 → 実行 → 確認）の順序は、IT スクリプトが同順で
  project init → encrypt passwd → inventory → create scenario → run scenario を実行していることを
  一次証拠とする（事実: build/product/integration_test.sh:67,89,108,131,188 の STEP 定義順）。
- アクター名は Phase2（02-value.md）の定義「テスト実行者 / シナリオ作成者 / 環境管理者 /
  テスト結果確認者」と一致させる。外部システム名は「digdag / webhook 受信先 / テスト対象ホスト群」を使う。
- UC 候補名は Phase4 で確定する仮置きである。

## 業務領域一覧

| 業務 | 概要 | 確度 | 根拠 |
|------|------|------|------|
| テスト環境準備 | stfw の導入（アーカイブ展開・install）、プロジェクト初期化、暗号化キー・パスワード・inventory 等の接続情報準備を行う | high | 事実: docs/adoc/install/install_index.adoc:1-65（Installation / Getting Started）, build/product/integration_test.sh:67-125（STEP: project init / encrypt passwd / inventory） |
| シナリオ作成 | scenario > bizdate > process の 3 階層 scaffold を生成し、テストスクリプトと設定を配置してワークフロー定義（dig）を生成する | high | 事実: docs/adoc/manual/manual_index.adoc:5（=== シナリオ作成）, build/product/integration_test.sh:131-168（STEP: create scenario）, src/plugins/process/scripts/README.adoc:11-27（ディレクトリ構成） |
| シナリオ実行 | ワークフローエンジン（digdag server）を起動し、シナリオを一括自動実行する。dry-run・setup/teardown・ログ追従を含む | high | 事実: docs/adoc/manual/manual_index.adoc:3,7（=== server制御 / === シナリオ実行）, build/product/integration_test.sh:188-212（STEP: run scenario）, src/bin/cmd/run:31-38 |
| テスト結果確認 | 実行結果を webhook 通知・digdag Web UI・ログで監視・確認する | medium | 事実: src/config/webhook/payload.yml:1-22, src/bin/lib/stfw/domain/service/run_service:135-139（attempt URL 生成）, docs/adoc/install/install_index.adoc:67-68（digdag URL の案内）。推測: 独立した「業務」としての括りは、通知/UI/ログという確認手段の実装が揃っていることからの再構成。manual に確認手順の章は無い |

- 補足: フレームワーク自体の開発・保守（CI・ドキュメント公開・リリース）はステークホルダー「stfw 開発者・
  コントリビューター」の活動であり、Phase2 の申し送りどおり本システムの業務（BUC）には含めない
  （事実: 02-value.md アクター一覧の注記, .travis.yml:36-47）。

## BUC 一覧

| 業務 | BUC | 価値（何を実現するか） | 主なアクター | 確度 | 根拠 |
|------|-----|---------------------|-------------|------|------|
| テスト環境準備 | stfw を導入する | tar.gz 展開 + install のみで各ホストにテスト実行基盤を用意できる | 環境管理者 | high（アクターは medium） | 事実: docs/adoc/install/install_index.adoc:17-53（Getting Started: download → tar → bin/install → ln -s）, build/product/integration_test.sh:46-61。推測: 導入担当のロール定義は無く、Phase2 の「環境管理者」（環境系作業の論理区分）に割当 |
| テスト環境準備 | プロジェクトを初期化する | テンプレート一式（stfw.yml / config / sample シナリオ）を持つプロジェクトを即座に開始できる | テスト実行者 | high | 事実: src/bin/cmd/init:25（initialize stfw project）, build/product/integration_test.sh:70-83, src/template/（stfw.yml, config/, scenario/sample/） |
| テスト環境準備 | テスト対象ホストの接続情報を管理する | 環境別 inventory によるホストグループ管理と、資格情報の暗号化保管・参照ができる | 環境管理者 | high | 事実: build/product/integration_test.sh:89-125（STEP: encrypt passwd / inventory）, src/bin/cmd/inventory:34-35, src/bin/cmd/passwd:25,31, src/bin/cmd/gen-encrypt-key:25 |
| シナリオ作成 | テストシナリオを作成する | 業務日付をまたぐ一連の業務処理を、ディレクトリ構造 + スクリプトとして記述できる | シナリオ作成者 | high | 事実: build/product/integration_test.sh:131-162（scenario -i → bizdate -i → process -i）, src/plugins/process/scripts/README.adoc:19-25,243-245, src/bin/cmd/{scenario:34, bizdate:34, process:36} |
| シナリオ作成 | ワークフロー定義を生成・検証する | ディレクトリ構造から dig 定義を自動生成し、dry-run で実タスクなしに検証できる | シナリオ作成者 | high | 事実: src/bin/cmd/scenario:35-36（--generate-dig / --generate-dig-cascade）, build/product/integration_test.sh:165-168, src/bin/cmd/run:34（--dry-run）, src/bin/cmd/process:37 |
| シナリオ作成 | プロセスプラグインを拡張する | テスト対象固有のプロセス種別（同梱は scripts のみ）を追加できる | シナリオ作成者 | medium（価値の理由は low） | 事実: src/bin/cmd/process:34-38（--list / --install / --init <process-type>）, src/plugins/process/{__common,scripts}, git log e811c59（"プロセスを追加できるように、既存のプロセスを整理"）。推測: 拡張が必要になる具体的ユースケースの明文は無い |
| シナリオ実行 | 実行基盤（server）を制御する | digdag server の起動・停止・再起動・状態確認と、bind/port/状態 DB の切替ができる | テスト実行者 | high | 事実: src/bin/cmd/server:31-38（start\|stop\|restart\|status とオプション）, build/product/integration_test.sh:192-212 |
| シナリオ実行 | シナリオを一括自動実行する | 指定シナリオ群を digdag ワークフローとして自動実行し、順序保証・エラー時停止・結果通知を得る | テスト実行者（実行主体は外部システム digdag） | high | 事実: src/bin/cmd/run:31-38, src/bin/lib/stfw/domain/service/run_service:31-56（push_digdag_proj → start → attempt_id 保存）, src/template/scenario/sample/scenario.dig:3-30, build/product/integration_test.sh:201 |
| テスト結果確認 | 実行結果を監視・確認する | 進捗・成否・所要時間を webhook 通知/Web UI/ログで把握し、失敗時の調査ができる | テスト結果確認者 | medium | 事実: src/config/webhook/payload.yml:1-22, src/bin/cmd/run:35（--follow）, src/bin/lib/stfw/domain/service/run_service:135-139（attempt URL）。推測: 確認手順の明文ドキュメントは無く、実装された確認手段からの再構成 |

## BUC 別アクティビティ

### stfw を導入する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | GitHub Releases から配布アーカイブをダウンロードする | 環境管理者 | （システム外: curl） | high | 事実: docs/adoc/install/install_index.adoc:23-42（download_url と curl 手順） |
| 2 | アーカイブを展開し install を実行する（依存モジュール取得を含む） | 環境管理者 | インストールする | high | 事実: docs/adoc/install/install_index.adoc:49-52, src/bin/install:100-144（依存ダウンロード）, build/product/integration_test.sh:47-57 |
| 3 | stfw コマンドを PATH に登録する | 環境管理者 | （システム外: ln -s / PATH 設定） | high | 事実: docs/adoc/install/install_index.adoc:52（ln -s）, build/product/integration_test.sh:60-61 |

### プロジェクトを初期化する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | プロジェクトディレクトリを作成する | テスト実行者 | （システム外: mkdir） | high | 事実: docs/adoc/install/install_index.adoc:60-61, build/product/integration_test.sh:70-72 |
| 2 | プロジェクトを初期化する（stfw init。テンプレート・sample シナリオ展開） | テスト実行者 | プロジェクトを初期化する | high | 事実: src/bin/cmd/init:25,70, build/product/integration_test.sh:73, src/template/scenario/sample/ |
| 3 | プロジェクト設定（stfw.yml: webhook URL / inventory / timezone / server 設定等）を編集する | テスト実行者 | （システム外: エディタ） | medium | 事実: src/template/stfw.yml:1-33（編集前提のコメント付きテンプレート）, build/product/integration_test.sh:174-183（CI が sed で webhook 設定を有効化）。推測: 編集手順の明文ドキュメントは無い（docs/adoc/config/config_index.adoc:1 が見出しのみ） |
| 4 | 暗号化キーペアを生成する（stfw gen-encrypt-key） | 環境管理者 | 暗号化キーを生成する | high | 事実: src/bin/cmd/gen-encrypt-key:25（generate encrypt key, --force）, build/product/integration_test.sh:77-83 |

### テスト対象ホストの接続情報を管理する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | inventory ファイルに環境・ホストグループ（web/ap/db 等）を定義する | 環境管理者 | （システム外: エディタ） | medium | 事実: src/template/config/inventory/staging.yml:1-8, src/template/stfw.yml:7（inventory: staging.yml）。推測: 編集は手作業前提（inventory を書き換える CLI は無い） |
| 2 | inventory 定義を確認する（グループ存在確認・ホスト一覧） | 環境管理者 | inventory を参照する | high | 事実: src/bin/cmd/inventory:34-35（--is-exist / --list）, build/product/integration_test.sh:111-125 |
| 3 | ホスト×ユーザー単位のパスワードを暗号化登録する（stfw passwd <host> <user> <password>） | 環境管理者 | パスワードを暗号化登録する | high | 事実: src/bin/cmd/passwd:25,31, build/product/integration_test.sh:92-98, src/bin/lib/commons/bash_utils:255-299（RSA+S/MIME 暗号化） |
| 4 | 登録済みパスワードを復号表示して確認する（stfw passwd --show） | 環境管理者 | パスワードを参照する | high | 事実: src/bin/cmd/passwd の --show オプション, build/product/integration_test.sh:100-102 |
| 5 | テスト対象ホストの SSH サーバキーを known_hosts へ登録する | 環境管理者 | （UC候補: SSH サーバキーを登録する。呼出経路は未確定） | low | 推測: 関数 gen_ssh_server_key は実装済み（事実: src/bin/lib/commons/bash_utils:131-164）だが、CLI コマンド・dig からの呼び出し箇所が無く、ユーザースクリプトから利用する想定・実施タイミングは不明。FIXME 参照 |

### テストシナリオを作成する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | テストシナリオ（テスト対象の業務フロー・日付進行）を設計する | シナリオ作成者 | （システム外: 人手の設計作業） | medium | 推測: scenario > bizdate > process の階層（事実: src/plugins/process/scripts/README.adoc:19）を埋めるには事前設計が必要という導出。設計手順のドキュメントは無い |
| 2 | シナリオ scaffold を生成する（stfw scenario -i <scenario-name>） | シナリオ作成者 | シナリオ scaffold を生成する | high | 事実: src/bin/cmd/scenario:34, build/product/integration_test.sh:135-138 |
| 3 | 業務日付 scaffold を生成する（stfw bizdate -i <seq> <YYYYMMDD>。日付ごとに繰り返し） | シナリオ作成者 | 業務日付 scaffold を生成する | high | 事実: src/bin/cmd/bizdate:34, build/product/integration_test.sh:141-156（day1/day2 の 2 回実行） |
| 4 | プロセス scaffold を生成する（stfw process -i <seq> <group> <process-type>） | シナリオ作成者 | プロセス scaffold を生成する | high | 事実: src/bin/cmd/process:36, build/product/integration_test.sh:147-162 |
| 5 | テストスクリプトを scripts/ に配置する（任意言語。ファイル名昇順 = 実行順） | シナリオ作成者 | （システム外: スクリプト作成） | high | 事実: src/plugins/process/scripts/README.adoc:22-25,212,243-245, src/template/scenario/sample/_10_99990101/_10_pre_scripts/scripts/{100_1st_step,200_2nd_step} |
| 6 | Plugin/Process 設定（config.yml）に共通環境変数を定義する | シナリオ作成者 | （システム外: エディタ） | high | 事実: src/plugins/process/scripts/README.adoc:131-178, src/template/scenario/sample/_10_99990101/_10_pre_scripts/config/config.yml |
| 7 | run/scenario/bizdate/process 各階層の setup/teardown スクリプトを必要に応じて配置する | シナリオ作成者 | （システム外: スクリプト作成） | medium | 事実: src/template/plugins/{run,scenario,bizdate,process}/_common/{setup,teardown}/ のテンプレート。推測: 用途説明のドキュメントが無く、配置が任意か必須かは読み取れない |

### ワークフロー定義を生成・検証する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | ディレクトリ構造からワークフロー定義（dig）を生成する（stfw scenario -g / -G, stfw bizdate -g） | シナリオ作成者 | ワークフロー定義を生成する | high | 事実: src/bin/cmd/scenario:35-36, src/bin/cmd/bizdate:35, build/product/integration_test.sh:165-168（scenario -G） |
| 2 | dry-run で実タスクを実行せずワークフローを検証する（stfw run -d / stfw process -d） | シナリオ作成者 | シナリオを dry-run する | high | 事実: src/bin/cmd/run:34（doesn't execute tasks）, src/bin/cmd/process:37（run setup, pre_execute, teardown） |
| 3 | 生成された dig（scenario.dig / bizdate.dig）を確認・必要に応じ調整する | シナリオ作成者 | （システム外: エディタ） | medium | 事実: src/template/scenario/sample/scenario.dig:1-30, _10_99990101/bizdate.dig:1-23（生成物のサンプル）。推測: 手動調整の可否・手順は明文化されていない |

### プロセスプラグインを拡張する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | 利用可能なプロセスプラグインを一覧確認する（stfw process -l） | シナリオ作成者 | プラグインを一覧する | high | 事実: src/bin/cmd/process:34, src/bin/cmd/process:49（デフォルトコマンド list） |
| 2 | プラグインの依存モジュールをインストールする（stfw process -I <process-type>） | シナリオ作成者 | プラグイン依存をインストールする | high | 事実: src/bin/cmd/process:35 |
| 3 | 新規プロセスタイプのプラグインを作成する（__common の構造に従い setup/execute/teardown 等を実装） | シナリオ作成者 | （システム外: プラグイン開発） | medium | 事実: src/plugins/process/{__common,scripts} の構造, git log e811c59（"プロセスを追加できるように、既存のプロセスを整理"）。推測: プラグイン作成手順のドキュメントは無い |

### 実行基盤（server）を制御する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | server を起動する（bind/port/状態 DB/スレッド数のオプション指定可） | テスト実行者 | server を起動する | high | 事実: src/bin/cmd/server:31-38, build/product/integration_test.sh:192, docs/adoc/install/install_index.adoc:63（stfw server start） |
| 2 | server の稼働状態を確認する（stfw server status） | テスト実行者 | server 状態を確認する | high | 事実: src/bin/cmd/server:31,45-54（private.status）, build/product/integration_test.sh:196-212 |
| 3 | server を停止・再起動する（stop / restart） | テスト実行者 | server を停止する | high | 事実: src/bin/cmd/server:31, build/product/integration_test.sh:206 |

### シナリオを一括自動実行する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | 実行前の run 共通 setup を実行する（stfw run -s。省略時は run 内で自動実行） | テスト実行者 | run setup を実行する | medium | 事実: src/bin/cmd/run:36（-s, --setup）, src/bin/lib/stfw/domain/service/run_service:84-101。推測: 手動 setup と自動フローの使い分け（いつ -s を使うか）は明文化されていない |
| 2 | シナリオを実行する（stfw run <scenario-names...>。run_id 発行 → digdag プロジェクト push → 実行開始） | テスト実行者 | シナリオを実行する | high | 事実: src/bin/cmd/run:31,113-117, src/bin/lib/stfw/domain/service/run_service:31-56, build/product/integration_test.sh:201（stfw run -f test） |
| 3 | （システム動作）digdag がワークフローを解釈し、scenario/bizdate 単位の setup → process 実行 → teardown を stfw に呼び戻して逐次実行する | （外部システム: digdag） | scenario/bizdate/process の setup・execute・teardown（Phase4 で UC 化） | high | 事実: src/template/scenario/sample/scenario.dig:3-30（sh>: stfw scenario --setup/--teardown）, _10_99990101/bizdate.dig:3-23（sh>: stfw process ${run_mode}）, docs/design/architecture/app_arch.puml:39（digdag -up-> stfw） |
| 4 | （システム動作）スクリプトをファイル名昇順に実行し、エラー発生時は後続を Blocked としてエラー終了する | （外部システム: digdag → stfw scripts plugin） | プロセスを実行する（Phase4 で UC 化） | high | 事実: src/plugins/process/scripts/README.adoc:202-213, src/bin/lib/setenv:27-31（Pending/Started/Success/Error/Blocked） |
| 5 | （システム動作）start/success/error の各時点で webhook 受信先へ JSON payload を通知する | （外部システム連携: webhook 受信先） | 結果を通知する（Phase4 で UC 化） | high | 事実: src/config/webhook/{payload.yml,run.yml,scenario.yml,bizdate.yml,process.yml}, src/bin/lib/stfw/domain/gateway/webhook_gateway:87-95, src/template/stfw.yml:20-22（notify_on_start/success/error） |
| 6 | 実行後の run 共通 teardown を実行する（stfw run -t。省略時はワークフロー内で自動実行） | テスト実行者 | run teardown を実行する | medium | 事実: src/bin/cmd/run:37, src/bin/lib/stfw/domain/service/run_service:104-121。推測: 手動 teardown の運用タイミングは明文化されていない |

### 実行結果を監視・確認する

| 順 | アクティビティ | アクター | システム利用（UC候補） | 確度 | 根拠 |
|----|--------------|---------|----------------------|------|------|
| 1 | 実行ログをリアルタイムに追従する（stfw run -f） | テスト結果確認者 | 実行ログを追従する | high | 事実: src/bin/cmd/run:35（show new logs until attempt or task finishes）, src/bin/lib/stfw/domain/service/run_service:63-81 |
| 2 | webhook 通知（進捗・成否・所要時間・digdag URL）を受信して監視する | テスト結果確認者（受信システムは外部システム: webhook 受信先） | （システム外: 通知の受信・閲覧） | medium | 事実: src/config/webhook/payload.yml:1-22（attempt URL・ステップ別結果を含む）。推測: 受信先での監視運用（チャット・CI 等）は設定例が webhook.site のみで実運用形態は不明 |
| 3 | digdag Web UI / API で attempt の実行状況・詳細を確認する | テスト結果確認者 | （外部システム UI: digdag） | medium | 事実: src/bin/lib/stfw/domain/service/run_service:135-139（http://{ip}:{port}/attempts/{attempt_id}）, src/bin/cmd/server:35（web interface and api clients）, docs/adoc/install/install_index.adoc:67-68。推測: UI での確認手順のドキュメントは無い |
| 4 | ログファイル（シークレットマスキング済み）で障害調査を行う | テスト結果確認者 | （システム外: ログ閲覧） | medium | 事実: src/bin/lib/setenv:144（PATH_LOG）,146-151（log.mask）。推測: 調査手順の明文は無い |
| 5 | エラー時にシナリオ・スクリプトを修正し、シナリオを再実行する | テスト結果確認者 → シナリオ作成者 / テスト実行者 | シナリオを実行する（再実行） | low | 推測: テスト業務の一般論による補完。エラー時の修正 → 再実行フロー（リラン・途中再開等）の実装・ドキュメントはリポジトリに存在しない |

## FIXME / 申し送り

- FIXME: docs/adoc/manual/manual_index.adoc は「server制御 / シナリオ作成 / シナリオ実行」の見出しのみで
  本文が空。業務・BUC の括りはこの見出しと integration_test.sh の STEP 構成から再構成した
  （事実: docs/adoc/manual/manual_index.adoc:1-7, docs/adoc/config/config_index.adoc:1 も見出しのみ）。
  Phase3 ユーザー確認で業務の括り・命名の妥当性を確認すること。
- FIXME: `gen_ssh_server_key`（src/bin/lib/commons/bash_utils:131-164）は定義されているが、
  リポジトリ内に呼び出し箇所が無い（定義済み・未参照）。「接続情報を管理する」BUC のアクティビティ 5 として
  仮置きしたが、ユーザースクリプトからの利用想定か廃止予定かを確認すること。
- FIXME: エラー時のリカバリ業務（失敗シナリオの再実行・途中再開）の仕組み・手順が読み取れない。
  digdag 自体は resume 機能を持つが、stfw としての公式手順は無い。アクティビティは `low` で仮置きした。
- 申し送り（Phase4 へ）: 「シナリオを一括自動実行する」のアクティビティ 3〜5（digdag からの呼び戻し:
  scenario/bizdate --setup/--teardown, process --run/--dry-run, webhook 通知）は、
  digdag を起点とするシステム動作として UC・イベント整理の中心になる
  （事実: src/template/scenario/sample/scenario.dig, bizdate.dig, app_arch.puml:39）。
- 申し送り（Phase4 へ）: UC 候補の粒度は本書ではコマンド機能単位で仮置きした。Phase4 では
  「ユーザーの目的単位」（analysis-targets.md 原則 6）で統合・分割を再検討すること
  （例: scaffold 生成 3 コマンドを「シナリオ構造を組み立てる」1 UC に統合する等）。
- 申し送り: タイマー・スケジュール実行（cron / digdag schedule）の証拠は無し。BUC「シナリオを
  一括自動実行する」の起動は常に人手（stfw run）である（事実: src/template/scenario/sample/scenario.dig に
  schedule 定義なし。推測: 定期実行運用は想定外か未実装）。

## confidence: low 項目一覧（ユーザー確認対象）

| # | 項目 | 内容 | 手がかり |
|---|------|------|---------|
| 1 | BUC「プロセスプラグインを拡張する」の価値の理由 | 「テスト対象固有のプロセス種別への対応のため」という理由付け | 推測: git log e811c59 は「追加できるように整理」のみで、拡張が必要になる具体的ユースケースの明文なし |
| 2 | アクティビティ「SSH サーバキーを known_hosts へ登録する」 | BUC「接続情報を管理する」への帰属と実施タイミング | 推測: 関数実装（bash_utils:131-164）はあるが呼び出し箇所ゼロ。ユーザースクリプトからの利用想定と判断 |
| 3 | アクティビティ「エラー時にシナリオを修正し再実行する」 | 結果確認 BUC の締めとしての修正 → 再実行フロー | 推測: テスト業務の一般論による補完。リラン・途中再開の実装・手順はリポジトリに無い |
