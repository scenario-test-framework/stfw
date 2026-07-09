# 変更サマリ

- event_id: 20260708_084933_v1_rearchitecting
- 元USDM: 20260708_084933_v1_rearchitecting
- 生成日時: 2026-07-08T09:06:41

対応要求: REQ-001（配布・導入の刷新）、REQ-002（3階層記述と静的検証）、REQ-003（内蔵ランナーによる一括自動実行）、REQ-004（廃止: 実行基盤制御）、REQ-006（status / HTML レポートによる結果確認）、REQ-007（inventory / secret / ssh trust）、REQ-008（プラグイン再編）、REQ-011（Go テスト・CI/CD 刷新）

補足:

- BUC.tsv / 状態.tsv は、変更のあった BUC・状態モデルの該当グループの完全行セットで収録している（グループ単位で latest を置換する）
- BUC「ワークフロー定義生成・検証フロー」は静的検証（stfw validate）のフロー「シナリオ静的検証フロー」に再編した（旧名の行は削除し、新名で追加）
- 各 BUC のアクティビティ・UC 名と説明は v1.0 CLI 体系（new / validate / run / status / report / secret / ssh trust / plugin / inventory）に更新した

## 追加

- BUC: シナリオ静的検証フロー（業務: シナリオ作成業務。「ワークフロー定義生成・検証フロー」の再編。アクティビティ: シナリオ構造を静的検証する（UC: シナリオを検証する = stfw validate）、dry-runで実行経路を検証する（UC: シナリオをdry-runする = stfw run --dry-run）、検証済みシナリオで本実行に臨める）
- 情報: HTML レポート（属性: 出力先（.stfw/reports/、index + run 詳細）、静的 HTML、増分再生成（process 終了ごと）、nginx 配信（Docker Compose 構成、http://localhost:8080）。関連情報: 実行ジャーナル（journal.jsonl）、実行（run）。コンテキスト: 通知管理）
- 外部システム: 配布元（GitHub Releases / ghcr.io）（外部システム群: 配布基盤。マルチプラットフォームバイナリ・Docker image・compose.yaml の配布元）
- 外部システム: 開発 CI 基盤（GitHub Actions）（外部システム群: 開発基盤。PR で lint + test、master マージで snapshot ビルド、tag 付与で goreleaser リリースと ghcr.io 配布。SPEC-011-01 の追加と SPEC-011-02 の変更を統合して収録）
- 条件: server 設定の廃止警告（stfw.server.* を含む stfw.yml の読み込み時に廃止警告を表示し、設定値は実行に影響させない。コンテキスト: プロジェクト環境管理）
- 条件: run 前静的検証（run 開始前に validate 相当の静的検証を自動実行し、エラー時は実行を開始しない。残存 *.dig への不要警告を含む。コンテキスト: 実行管理）
- 条件: run_id の採番・保持（run_id は _{YYYYMMDDHHMMSS}_{PID} 形式。attempt_id は存在しない。「run_id / attempt_id の採番・保持」の置換。コンテキスト: 実行管理）
- 条件: 資格情報の旧形式移行（旧 S/MIME 形式は読み込み専用サポート、stfw secret migrate で age 形式へ一括変換・.bak 退避。コンテキスト: プロジェクト環境管理）
- バリエーション: 資格情報暗号化方式（値: age (X25519)（現行）、S/MIME (RSA+AES256)（旧形式・読み込み専用）。コンテキスト: プロジェクト環境管理）

## 変更

- BUC: stfw導入フロー → install スクリプト・依存モジュールダウンロード（dl.bintray.com）の廃止に伴い、配布物（バイナリ / Docker image / compose.yaml）の取得と配置・コンテナ起動のシステム外作業に再構成。UC「stfwをインストールする」・画面「インストールCLI」・イベント「依存モジュールダウンロード」を削除
- BUC: プロジェクト初期化フロー → stfw.yml 編集の説明から server 設定を除去（廃止警告に言及）。UC「暗号化キーを生成する」を stfw secret keygen（age (X25519)）の記述に更新
- BUC: 接続情報管理フロー → inventory 確認を stfw inventory exists / list に更新。UC「パスワードを暗号化登録・参照する」を「資格情報を暗号化登録・参照する」（stfw secret set / show、age (X25519)）に改名、画面「パスワード管理CLI」を「資格情報管理CLI」に改名。アクティビティ「旧形式の資格情報をage形式へ移行する」（UC: 資格情報を旧形式から移行する = stfw secret migrate）を追加。UC「SSHサーバキーを登録する」を stfw ssh trust <host|group> の正式コマンド（グループ一括登録）として更新
- BUC: テストシナリオ作成フロー → scaffold 生成を stfw new scenario / new bizdate / new process（旧 scenario -i / bizdate -i / process -i）に更新
- BUC: シナリオ一括自動実行フロー → stfw run <scenario-names...> の 1 コマンド統合（push・server 前提の廃止、run 前静的検証の自動実行）に更新。UC「階層setup/teardownを実行する」「プロセスを実行する」を digdag 呼び戻しではなく内蔵ランナーの内部実行として記述を更新（プロセスは setup → pre_execute → execute → post_execute → teardown の 5 段階）。関連の「ワークフロー定義（dig）」「run_id / attempt_id の採番・保持」・イベント「ワークフロー実行依頼」「階層タスク呼び戻し」「プロセスタスク呼び戻し」（digdag）を除去し、「run 前静的検証」「run_id の採番・保持」を関連付け。画面「階層前後処理CLI」「プロセス実行CLI」は内部実行化に伴い「シナリオ実行CLI」に集約
- BUC: 実行結果監視・確認フロー → アクティビティ「実行ログを追従表示して監視する」（UC: 実行ログを追従する、digdag）を「statusで実行状況を確認する」（UC: 実行状況を確認する = stfw status、実行ジャーナルのリプレイ）に置換。「HTMLレポートで実行結果を閲覧する」（UC: HTMLレポートを生成する = stfw report、nginx 配信）を追加。「Web UI・ログファイルで失敗箇所を調査する」を「レポート・ログファイルで失敗箇所を調査する」に置換。「修正後のシナリオを再実行する」から「ワークフロー定義（dig）」「run_id / attempt_id の採番・保持」・イベント「ワークフロー実行依頼」（digdag）への参照を除去
- BUC: プロセスプラグイン拡張フロー → stfw plugin list / install（旧 process -l / -I、解決順互換）に更新。scaffold 生成は stfw new process、env 契約（stfw_* / STFW_PROJ_DIR 系）維持を明記
- アクター: テスト実行者 → 実行基盤（digdag server）制御の責務を除去し、stfw run 1 コマンド実行に更新。主担当業務から「実行基盤制御フロー」を除去し「ワークフロー定義生成・検証フロー」を「シナリオ静的検証フロー」に置換（削除 BUC への参照整理）
- アクター: シナリオ作成者 → ワークフロー定義の生成を除去し、stfw new による scaffold 生成と静的検証（stfw validate）・dry-run 検証に更新（参照整理）
- アクター: 環境管理者 → 導入作業を配布物の取得・配置に更新し、stfw secret（旧形式移行含む）・stfw ssh trust を追加（参照整理）
- アクター: テスト結果確認者 → 確認手段を stfw status（ジャーナルリプレイ）・stfw report（静的 HTML レポート）・OTel トレース・ログファイルに変更（digdag Web UI・ログ追従を除去）
- 外部システム: テスト対象ホスト群 → 資格情報の暗号化方式（age (X25519)）と stfw ssh trust による一括登録を反映
- 情報: stfw 本体 → 属性を配布形態（マルチプラットフォームバイナリ / Docker image / compose.yaml）・内蔵ランナー・組込みプラグイン・内蔵デフォルト設定に変更。STFW_HOME・依存モジュール参照を除去し、Go 単一バイナリ（ランタイム不要）として記述
- 情報: プロジェクト → .stfw/ 内部データを runs / reports に更新。関連情報から「サーバプロセス」を除去（参照整理）
- 情報: プロジェクト設定（stfw.yml） → server.* 属性を除去し廃止警告を明記。関連情報から「状態 DB（digdag DB）」、バリエーションから「DB モード」を除去。上書き順をデフォルト（内蔵）→ プロジェクトに更新（STFW_HOME/config 廃止）
- 情報: インベントリ → stfw inventory exists / list（旧 --is-exist / --list と出力互換）に更新
- 情報: 暗号化キー → RSA 2048 キーペアから age (X25519) キーペアに変更。生成コマンドを stfw secret keygen（--force）に更新
- 情報: パスワード → 暗号化方式を age (X25519) に変更（バリエーション: 資格情報暗号化方式）。旧 S/MIME の読み込み専用サポートと stfw secret migrate による一括移行を追加
- 情報: プラグイン → 管理コマンドを stfw plugin list / install に更新。組込み配置を配布物側に変更（STFW_HOME 廃止）。env 契約（stfw_* / STFW_PROJ_DIR 系）の維持を属性に明記
- 情報: SSH サーバキー（known_hosts） → stfw ssh trust <host|group> の正式コマンド化（旧キー削除＋新キー登録、グループ一括登録）に更新し「未配線」注記を除去
- 情報: シナリオ → 属性・関連情報から scenario.dig・「ワークフロー定義（dig）」を除去（参照整理）。ディレクトリ構造そのものが実行定義となることを明記
- 情報: 業務日付 → 属性・関連情報から bizdate.dig・「ワークフロー定義（dig）」を除去（参照整理）
- 情報: プロセス → 実行順を setup → pre_execute → execute → post_execute → teardown の 5 段階に更新（SPEC-002-04 / SPEC-003-03 に伴う記述整合）
- 情報: Process / Plugin 設定（config.yml） → stfw validate / run 前静的検証の検証対象であることを追記
- 情報: メタ情報（metadata.yml） → 属性の「scaffold / dig 生成時」を「scaffold 生成時」に更新（参照整理）
- 情報: 実行（run） → attempt_id・digdag プロジェクト・digdag_start.info・params を除去し、run_id（従来形式維持）と実行ディレクトリ（.stfw/runs/{run_id}）に整理。関連情報から「ワークフロー定義（dig）」「digdag 管理情報」を除去し「HTML レポート」を追加
- 情報: 実行コンテキスト → attempt_id を除去し、内蔵ランナーの階層・プロセス実行間の引き継ぎとして記述を更新
- 情報: 実行ログ → terminal 実行時のカラー出力を属性に追加。ログ追従（run -f）による確認の廃止を反映し、障害調査用ログとして記述を更新
- 情報: OTel トレース（スパンツリー） → 投影元の実行ジャーナルをパス付き（.stfw/runs/{run_id}/journal.jsonl）で明記
- 情報: 実行ジャーナル（journal.jsonl） → 属性にパス（.stfw/runs/{run_id}/）・追記専用 JSONL・イベント種別（node_start / steps_enumerated / step_end / node_end）を明記。関連情報に「HTML レポート」を追加し、実行結果の唯一のソース・status / report のリプレイ元として記述を更新
- 状態: 階層実行ステータス → 遷移契機を digdag 呼び戻しから内蔵ランナーの直接実行に更新（_error 経路の記述を内蔵ランナーのエラー経路に変更）。確定結果の投影先に stfw status / report を追加
- 状態: ステップ実行ステータス → 列挙・逐次実行の主体を内蔵ランナーに更新（steps_enumerated イベントを明記）。終了状態の投影先に実行ジャーナルを明記
- 条件: プロジェクト再初期化禁止 → stfw init のコマンド名で明記（従来仕様維持）
- 条件: 設定の上書き順 → デフォルトを stfw 本体の内蔵デフォルトに変更（STFW_HOME/config 廃止）
- 条件: プラグイン解決順 → 組込みの配置を配布物同梱に変更（STFW_HOME 廃止）。env 契約の維持を明記
- 条件: 階層ディレクトリ判定 → 対象コマンドを scaffold 生成（stfw new）・静的検証（stfw validate）・プロセス実行に更新
- 条件: 実行対象ディレクトリ規則 → dig 自動生成時の規則から、静的検証（stfw validate）と内蔵ランナーの実行対象列挙の規則に更新
- 条件: dry-run の実行範囲 → 「execute / post_execute をスキップし setup → pre_execute → teardown は実行する」（stfw run --dry-run）に整理
- 条件: run 実行の前提条件 → digdag server 起動中の前提を廃止し、対象シナリオの存在と run 前静的検証の通過に変更（状態モデル「server 稼働状態」への参照を除去）
- 条件: 逐次実行・エラー時 Blocked → 提供主体を内蔵ランナーに更新（従来と同一仕様）
- 条件: シークレットマスキング → v1.0 でも従来仕様を維持することを明記
- 条件: パスワード重複登録禁止 → stfw secret set のコマンド名で明記
- 条件: 暗号化キー再生成の抑止 → stfw secret keygen のコマンド名で明記
- バリエーション: 実行モード（run_mode） → dry-run の意味を「execute / post_execute のスキップ」に整理
- バリエーション: 対応 OS 種別 → 値を linux、darwin（mac）、windows に変更（配布バイナリのプラットフォーム区分。JAVA_HOME 依存分岐は廃止）
- バリエーション: プラグインスコープ → 組込みの配置を配布物同梱に変更（STFW_HOME 廃止）

## 削除

- BUC: 実行基盤制御フロー
- BUC: ワークフロー定義生成・検証フロー（「シナリオ静的検証フロー」へ再編・名称変更）
- 情報: 依存モジュール
- 情報: ワークフロー定義（dig）
- 情報: サーバプロセス
- 情報: 状態 DB（digdag DB）
- 情報: digdag 管理情報
- 状態: server 稼働状態
- 条件: server 多重起動禁止
- 条件: run_id / attempt_id の採番・保持（「run_id の採番・保持」へ置換）
- バリエーション: dig 生成モード
- バリエーション: DB モード
- 外部システム: digdag
- 外部システム: dl.bintray.com
