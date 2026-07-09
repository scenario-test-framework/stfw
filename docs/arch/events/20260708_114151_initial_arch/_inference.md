# アーキテクチャ推論根拠サマリ

- event_id: 20260708_114151_initial_arch
- created_at: 2026-07-08T11:41:51
- trigger_event: rdra:20260708_103305_builtin_plugin_ecosystem, nfr:20260708_112906_nfr_user_confirm

## 推論の前提（as-is 制約）

stfw は v1.0 として実装済み。以下は白紙推論ではなく、承認済み計画（golang-docker-image-compose-yaml リアーキテクティング計画）と実装（internal/ 配下）を as-is 制約として反映した:

- Go 単一バイナリ CLI（サーバレス・実行エンジン内包・逐次実行）
- 5 層構成（presentation / usecase / domain / repository / gateway。domain 依存ゼロ・IF なし直接依存）
- domain の BC 分割: scenario・run（Core）/ notify・project（Supporting）
- データストアなし（実行ジャーナル JSONL・stfw.yml・age 暗号化ファイル）
- OTLP トレースエクスポート（ジャーナルイベントの投影）・静的 HTML レポート + nginx 配信
- 配布: GitHub Releases バイナリ / ghcr.io Docker image / compose.yaml、CI: GitHub Actions + goreleaser

未実装の組込みプラグイン群（collectLog / collectFile・export / import / clear・compare・invokeWeb / invokeRest）は to-be として設計対象に含めた（confidence: medium）。

## RDRA/NFR モデル分析結果

### 分析した RDRA 要素

| モデル | 要素数 | 主な特徴 |
|--------|--------|---------|
| BUC | 8 フロー / 4 業務 | テスト環境準備・シナリオ作成・シナリオ実行・テスト結果確認。バッチ/タイマー起動なし（全てオンデマンド CLI 操作） |
| アクター | 4 | 全員社内のテスト担当者（テスト実行者 / シナリオ作成者 / 環境管理者 / テスト結果確認者）。外部アクターなし |
| 外部システム | 8 | テスト対象ホスト群・テスト対象データストア・OTLP 受信先・配布元・開発 CI 基盤・テスト支援 OSS（logfilter / compare-files / grafana k6） |
| 情報 | 25 | 4 コンテキスト（プロジェクト環境管理 9 / シナリオ構造管理 7 / 実行管理 6 / 通知管理 3）。全てファイルベース |
| 状態 | 2 モデル | 階層実行ステータス（Started→Success/Error）・ステップ実行ステータス（Pending→Success/Error/Blocked）。いずれも終了状態から再遷移なし |
| 条件 | 29 | 命名規約・実行規約・秘匿情報保護・互換境界・OTel 送信制御が中心 |
| バリエーション | 13 | 実行モード・終了コード・プラグイン系（スコープ / フェーズ / 対応製品）・配布系（OS / イメージ構成） |

### 参照した NFR グレード

モデルシステム: model1（社内テスト担当者のみ・停止影響はテスト遅延に限定。ユーザー確認済み）

| カテゴリ | 傾向 | 主な影響 |
|---------|------|---------|
| A. 可用性 | ほぼ Lv1・災害対策 Lv0 | 冗長化・DR なしの単一構成。サーバレス・オンデマンド実行で十分（SP-001, SP-203） |
| B. 性能・拡張性 | ほぼ Lv1（レスポンス Lv2） | 単一利用者・逐次実行前提。増分レポートで長時間実行の進捗確認（CTP-007, SP-301） |
| C. 運用・保守性 | 監視範囲 Lv3・障害検知 Lv2 | OTLP トレース一本化 + ジャーナルリプレイ + 単一ログファイル（CTP-002, SP-005, SR-002） |
| D. 移行性 | Lv1 中心 | 互換境界 3 つの維持 + secret migrate（CTP-003, SR-201） |
| E. セキュリティ | Lv1 中心・WAF/診断 Lv0 | 認証認可の OS 委譲・age 暗号化・マスキング・known_hosts 検証（CTP-001, SP-202, CTP-004, CTR-001） |
| F. 環境 | 対応 OS Lv3 | マルチプラットフォーム単一バイナリ配布（SP-401） |

## ドメインアーキテクチャ推論

| 要素 | 判断 | confidence | 根拠 |
|------|------|-----------|------|
| SD-001 シナリオ構造管理 | core | user | 規約ベース記述（ディレクトリ構造 = 実行定義）が競争優位の源泉（承認済み計画で確定） |
| SD-002 実行管理 | core | user | 逐次実行・Blocked 伝播・ジャーナルがエンジン本体（承認済み計画で確定） |
| SD-003 通知管理 | supporting | user | ジャーナルイベントの投影に徹する。OTLP 標準へ委譲 |
| SD-004 プロジェクト環境管理 | supporting | user | 暗号化・SSH・HTML 生成等の Generic 能力はライブラリ採用（age, net/http, html/template）で自作回避 |
| BC 分割 | RDRA 4 コンテキスト = 4 BC | user | 情報.tsv のコンテキスト列と実装 internal/domain/{scenario,run,notify,project} が一致 |
| BC : tier 対応 | モジュラモノリス（1 バイナリ内 BC = パッケージ） | user | 単一チーム・NFR A Lv1・CLI 特性から独立サービス分割は過剰 |
| CM-002 notify→run | Published Language（journal イベント: node_start / steps_enumerated / step_end / node_end） | user | ジャーナルが唯一のソース、通知・レポートは投影。BC 跨ぎ直接ファイルアクセス禁止 |
| AG-001 Run 集約 | root: 実行（run）、member: 実行コンテキスト・ジャーナル | user | 承認済み計画で確定（リプレイ経路でも生成時と同じ検証） |
| AG-002 / AG-003 | シナリオ / プロジェクト集約仮説 | low | 実装は ScenarioTree（ファーストクラスコレクション）・トランザクションスクリプトであり、集約とするかは未確定 |

## システムアーキテクチャ推論

| ティア | テクノロジー候補 | confidence | 根拠 |
|--------|----------------|-----------|------|
| tier-cli | CLI（単一バイナリ）・内蔵実行エンジン | user | 実装済み as-is。常駐サービスなし（NFR A.1.1.1 Lv1） |
| tier-plugin | 外部プロセス実行・ssh/scp・DB クライアント・k6 | medium | to-be（組込みプラグイン群 REQ-012〜017）。env 契約・終了コードは実装済み互換境界 |
| tier-file-datastore | ローカルファイル（YAML / JSONL / 静的 HTML / 暗号化ファイル） | user | 実装済み as-is。外部データストアなし |
| tier-report-delivery | 静的 HTML + 静的 Web サーバ + 共有 volume | user | 実装済み as-is（compose 構成・読み取り専用配信） |
| tier-distribution | バイナリホスティング・Container Registry・CI/CD | user | 実装済み as-is（NFR F.1.1.1 Lv3 対応 OS） |

推論ルール上の標準ティア（フロントエンド / API Gateway / IdP / 認可サービス / ワーカー / RDB）は全て「不要」判定:

- フロントエンド不要: 外部アクターなし・画面は CLI と静的 HTML のみ
- API Gateway / IdP / 認可サービス不要: 全アクター社内・OAuth2/OIDC 認証なし・認可は OS 実行ユーザー + ファイルパーミッションに委譲（CTP-001）
- ワーカーティア不要: バッチ/タイマー/MQ なし（BUC にタイマー系アクティビティなし。長時間実行は CLI プロセス内の逐次実行）
- RDB / KVS 不要: トランザクション整合性が必要なエンティティなし。追記専用 JSONL + ファイルで十分

## アプリケーションアーキテクチャ推論

### tier-cli（5 層）

| レイヤー | 責務 | confidence | 根拠 |
|---------|------|-----------|------|
| presentation | cobra コマンド・引数パース・ロガー + マスキング | user | 実装済み（internal/presentation） |
| usecase | 10 ユースケース。runscenario が実行オーケストレーション | user | 実装済み（internal/usecase） |
| domain | BC 4 パッケージ・依存ゼロ・VO / 状態遷移型 | user | 実装済み（internal/domain）。状態遷移 2 モデル以上 + 条件 29 件 → 5 層選定ルールにも合致 |
| repository | aggregate root 単位のファイルアクセス・リプレイ復元 | user | 実装済み（internal/repository） |
| gateway | scriptexec / otlptrace / sshkeyscan / htmlwriter | user | 実装済み（internal/gateway） |

### tier-plugin（2 層）

| レイヤー | 責務 | confidence | 根拠 |
|---------|------|-----------|------|
| プラグイン契約層 | env 契約受領・終了コード返却 | medium | 単純なスクリプト実行のため 2 層で十分（ビジネスルールは CLI 本体側） |
| 外部ツール実行層 | ssh/scp・DB クライアント・k6・logfilter・compare-files | medium | to-be の組込みプラグイン群 |

## データアーキテクチャ推論

| 分類 | エンティティ | ストレージ | confidence | 根拠 |
|------|-------------|----------|-----------|------|
| event（追記のみ） | E-017 実行ログ, E-018 ステップ実行結果, E-019 OTel トレース, E-020 実行ジャーナル, E-024 エビデンス, E-025 比較結果 | file（E-019 のみ cache） | high（E-019 は low） | 発生日時を持つ一度きりの記録。journal.jsonl は UPDATE/DELETE なし |
| event_snapshot | E-015 実行（run） | file | high | 状態モデル（階層実行ステータス）を持つ。イベント = ジャーナル、スナップショットはリプレイで都度導出（永続スナップショットなし） |
| resource_mutable | 上記以外の 18 エンティティ | file（E-016 のみ cache） | high | 設定・定義・マスタ的ファイル。プロセス / スクリプトの実行状態はジャーナル側で管理 |

## 実装との齟齬（報告事項）

1. **システム概要.json が旧アーキテクチャの記述のまま**: 「ワークフローエンジン digdag で複数シナリオを一括自動実行」「ディレクトリ構造からワークフロー定義を自動生成」「webhook で外部システムへ通知」「ログ追従表示や digdag Web UI」と記載されている。情報.tsv・条件.tsv 等は v1.0（内蔵ランナー・OTel トレース・digdag 廃止）に更新済みであり、システム概要のみ不整合。docs/todo.md に修正提案を登録済み（RDRA 変更イベントの作成をユーザーに確認推奨）。latest は書き換えていない。
2. **推論ルールとの意図的な乖離（プロセス / スクリプトのデータモデル分類）**: 状態モデルを持つエンティティは event_snapshot が標準ルールだが、E-011 プロセス / E-012 スクリプトは resource_mutable とした。実装では「定義（静的ディレクトリ）」と「実行状態（ジャーナルイベント）」が分離されており、状態は E-020 実行ジャーナル側で記録されるため。
3. **AG-002 / AG-003 は実装と抽象度が異なる**: 実装は ScenarioTree（ファーストクラスコレクション）と Supporting のトランザクションスクリプトであり、DDD 集約としては未確定（confidence: low の仮説として記録）。

## ユーザー確認による変更

初期構築のため対話による変更なし（AskUserQuestion 不可の実行環境。確認推奨項目は実行結果として返却）。

## confidence 内訳

| セクション | high | medium | low | default | user | 合計 |
|-----------|:----:|:------:|:---:|:-------:|:----:|:----:|
| ドメインアーキテクチャ | 0 | 3 | 2 | 0 | 10 | 15 |
| システムアーキテクチャ | 1 | 14 | 0 | 1 | 20 | 36 |
| アプリケーションアーキテクチャ | 0 | 6 | 0 | 7 | 15 | 28 |
| データアーキテクチャ（storage_mapping） | 23 | 1 | 1 | 0 | 0 | 25 |
| 合計 | 24 | 24 | 3 | 8 | 45 | 104 |
