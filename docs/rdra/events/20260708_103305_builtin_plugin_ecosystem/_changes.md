# 変更サマリ

- event_id: 20260708_103305_builtin_plugin_ecosystem
- 元USDM: 20260708_103305_builtin_plugin_ecosystem
- 生成日時: 2026-07-08T10:47:07

対応要求: REQ-008（プラグイン 2 層構造の具体化）、REQ-012（collectLog / collectFile）、REQ-013（export / import / clear: MySQL / PostgreSQL / Redis）、REQ-014（compare による期待値比較）、REQ-015（invokeWeb / invokeRest: grafana k6）、REQ-016（横断規約: 接続情報のグループ名参照・エビデンスディレクトリ規約・ランタイム依存宣言）、REQ-017（カスタムプラグイン実装ガイド）

補足:

- 本イベントは「要求定義のみ・実装なし」の to-be 追加であり、既存要素の削除はない
- BUC.tsv / 状態.tsv は、変更のあった BUC・状態モデルの該当グループの完全行セットで収録している（グループ単位で latest を置換する。前イベントと同じマージ方式）
- 組み込みプラグインの実行は新 UC を追加せず、既存 UC「プロセスを実行する」の中でプロセスタイプとして実行される形で関連モデルを拡充した

## 追加

- 情報: エビデンス（属性: エビデンスディレクトリ（エビデンスディレクトリ規約に従う収集系プラグインの出力先）、収集ファイル（実行ログ・外部 IF ファイル・ヘッダー付き CSV）、収集元（ホストグループ / テーブル名リスト）。関連情報: プロセス、実行（run）、期待値（expect）、比較結果。コンテキスト: 実行管理）
- 情報: 期待値（expect）（属性: expect ディレクトリ（収集系プラグインの出力ディレクトリ構造と同型）、期待値ファイル。関連情報: シナリオ、エビデンス、比較結果。コンテキスト: シナリオ構造管理）
- 情報: 比較結果（属性: result ディレクトリ（gitignore 対象）、actual ディレクトリ（symlink または一時コピー、gitignore 対象）、比較成否。関連情報: エビデンス、期待値（expect）、ステップ実行結果。コンテキスト: 実行管理）
- 情報: カスタムプラグイン実装ガイド（属性: ガイドドキュメント、想定パターン（updateBizDate / invokeJob / importMaster / export / clear）、組み込みプラグインの組み合わせ例。関連情報: プラグイン。コンテキスト: プロジェクト環境管理）
- 外部システム: logfilter（ログ収集 OSS）（外部システム群: テスト支援ツール。collectLog が利用する scenario-test-framework/logfilter。scp 転送 → ssh 実行 → scp 収集 → バイナリ削除）
- 外部システム: compare-files（ファイル比較 OSS）（外部システム群: テスト支援ツール。compare が利用する scenario-test-framework/compare-files。起動設定・比較レイアウトはプロジェクトごとに保持）
- 外部システム: grafana k6（外部システム群: テスト支援ツール。invokeWeb（ブラウザモード）/ invokeRest が利用するテスト実行 OSS）
- 外部システム: テスト対象データストア（MySQL / PostgreSQL / Redis）（外部システム群: テスト対象システム。export / import / clear プラグインの接続先。MariaDB / Valkey は互換製品として同一プラグインでサポート。public cloud のマネージド製品は対象外）
- 条件: 収集先ホストとの時刻同期前提（collectLog のフィルタ基準時刻は実行ジャーナルの bizdate node_start イベント時刻（stfw_bizdate_start_ts 等）。ドキュメント明記の制約。コンテキスト: 実行管理）
- 条件: export/import ラウンドトリップ互換（export のヘッダー付き CSV を import がそのままインポート可能。コンテキスト: プロジェクト環境管理）
- 条件: エビデンスディレクトリ規約（収集系プラグインの出力構造 = compare の expect 構造。ディレクトリ規約・プラグイン env 契約に次ぐ第 3 の互換境界。具体的な命名は設計フェーズで確定。コンテキスト: シナリオ構造管理）
- 条件: 比較不一致はステップ失敗（差分検出時に該当ステップを Error とし、既存のエラー時停止・Blocked 伝播の対象とする。状態モデル: ステップ実行ステータス。コンテキスト: 実行管理）
- 条件: プラグイン接続情報のグループ名参照（プラグイン設定は inventory のホストグループ名参照のみ。資格情報は secret、SSH ホストキーは ssh trust の既存機構を利用。コンテキスト: プロジェクト環境管理）
- 条件: プラグインのランタイム依存宣言と存在チェック（前提コマンド（k6・mysql/psql クライアント・ssh/scp 等）を宣言し stfw validate / run 前静的検証が存在チェック。コンテキスト: プロジェクト環境管理）
- バリエーション: プラグインフェーズ（値: Arrange（準備）、Act（実行）、Collect（収集）、Assert（検証）。コンテキスト: プロジェクト環境管理）
- バリエーション: 対応データストア製品（値: MySQL（MariaDB 互換含む）、PostgreSQL、Redis（Valkey 互換含む）。コンテキスト: プロジェクト環境管理）
- バリエーション: Docker イメージ構成（値: 最小構成（既存）、依存全部入り（例: stfw:full）。コンテキスト: プロジェクト環境管理）

## 変更

- BUC: シナリオ一括自動実行フロー → アクティビティ「内蔵ランナーがプロセス配下スクリプトを逐次実行する」を「内蔵ランナーがプロセスをプラグインで実行する」に改名し、UC「プロセスを実行する」の説明に組込みプラグイン群（収集系 collectLog / collectFile、データストア系 exportXxx / importXxx / clearXxx、検証系 compare、実行系 invokeWeb / invokeRest）による Arrange → Act → Collect → Assert パイプラインを追記。UC「プロセスを実行する」の関連に情報（プラグイン、エビデンス、期待値（expect）、比較結果、インベントリ、パスワード、SSH サーバキー（known_hosts））・条件（収集先ホストとの時刻同期前提、export/import ラウンドトリップ互換、エビデンスディレクトリ規約、比較不一致はステップ失敗、プラグイン接続情報のグループ名参照）・イベント（ログ・ファイル収集（scp/ssh）= テスト対象ホスト群、ログフィルタ転送・実行 = logfilter（ログ収集 OSS）、k6テスト実行 = grafana k6、テスト対象データストア接続 = テスト対象データストア（MySQL / PostgreSQL / Redis）、ファイル比較実行 = compare-files（ファイル比較 OSS））を追加。UC「階層setup/teardownを実行する」の説明に bizdate node_start イベント時刻の env 契約公開（stfw_bizdate_start_ts 等）を追記。受益者アクティビティの説明を Arrange 〜 Assert の全自動反復に更新
- BUC: 接続情報管理フロー → UC「テスト対象ホスト情報を参照する」に条件「プラグイン接続情報のグループ名参照」を追加。inventory 定義・資格情報登録・SSH サーバキー登録・受益者の各説明に組み込みプラグインからのグループ名参照利用を追記
- BUC: シナリオ静的検証フロー → UC「シナリオを検証する」に情報「プラグイン」と条件「プラグインのランタイム依存宣言と存在チェック」を追加し、説明にランタイム依存（前提コマンド）の存在チェックを追記
- BUC: プロセスプラグイン拡張フロー → アクティビティ「カスタムプラグイン実装ガイドで想定パターンを確認する」（システム外作業: ドキュメント参照、アクター: シナリオ作成者）を追加。プラグイン一覧の説明に組込みプラグイン群を列挙し、カスタムプラグイン作成・受益者の説明を 2 層構造（組み込みプラグインの組み合わせによる実装）に更新
- アクター: シナリオ作成者 → 組み込みプラグイン群によるデータ準備・実行・エビデンス収集・期待値比較の記述と、カスタムプラグイン実装ガイドを参照したカスタムプラグイン実装を役割に追加
- 情報: プラグイン → 属性にフェーズ（Arrange / Act / Collect / Assert）・ランタイム依存宣言・env 契約（stfw_bizdate_start_ts 系）を追加。関連情報にカスタムプラグイン実装ガイド、バリエーションにプラグインフェーズを追加。説明を組込みプラグイン群とカスタムプラグインの 2 層構造に更新
- 情報: 実行ジャーナル（journal.jsonl） → bizdate node_start イベント時刻が collectLog のフィルタ基準時刻としてプラグイン env 契約（stfw_bizdate_start_ts 等）に公開されることを追記。関連情報にプラグインを追加
- 情報: インベントリ → 組み込みプラグインの収集先・接続先がホストグループ名参照のみで指定されることを追記。関連情報に Process / Plugin 設定（config.yml）を追加
- 情報: パスワード → 組み込みプラグインの接続資格情報としても既存機構のまま利用されることを追記
- 情報: SSH サーバキー（known_hosts） → collectLog / collectFile の scp / ssh 接続でも既存機構のまま利用されることを追記
- 情報: Process / Plugin 設定（config.yml） → 属性に接続先のホストグループ名参照・収集ファイルパス正規表現リスト・テーブル名リスト等を追加。関連情報にインベントリを追加し、接続情報を直接記述しないことを明記
- 情報: stfw 本体 → 属性に Docker image タグ構成（最小構成 / 依存全部入り（例: stfw:full））を追加し、バリエーションに Docker イメージ構成を追加
- 外部システム: テスト対象ホスト群 → 組み込みプラグインによる scp/ssh 経由のログ・ファイル収集（collectLog / collectFile）と grafana k6 による取引入力・検証（invokeWeb / invokeRest）の対象であることを追記
- 外部システム: 配布元（GitHub Releases / ghcr.io） → Docker image の依存全部入りタグ（例: stfw:full）の配布を追記
- 条件: プラグイン解決順 → カスタマイズ・差し替えの対象を組込みプラグイン群（収集系・データストア系・検証系・実行系）に拡充（解決順は従来仕様を維持）
- 条件: 逐次実行・エラー時 Blocked → エラー発生の契機に compare の比較不一致によるステップ失敗を含めることを追記
- 条件: run 前静的検証 → 検証項目にプラグインが宣言したランタイム依存（前提コマンド）の存在チェックを追加
- 状態: ステップ実行ステータス → Error 遷移の契機に compare の比較不一致を追加。Success 遷移に比較一致時の後続継続、Blocked 遷移に比較不一致による伝播を追記
- バリエーション: プロセスタイプ → 値に組込みプラグイン群（collectLog、collectFile、exportMysql / importMysql / clearMysql、exportPostgres / importPostgres / clearPostgres、exportRedis / importRedis / clearRedis、compare、invokeWeb、invokeRest）を追加
- バリエーション: プラグインスコープ → 組込みプラグイン群とカスタムプラグインの 2 層構造を成立させる区分であることを追記（値は変更なし）

## 削除

- なし
