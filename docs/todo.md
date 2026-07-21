# TODO / 追加提案

本ファイルは後続スキルからの追加提案を集約する。
RDRA に存在しない要素を追加する前に、ここで合意を得てから requirements スキルで反映する。

## 2026-07-08 dist-architecture からの追加提案

### DIST-001: システム概要.json の旧アーキテクチャ記述（digdag / webhook / ログ追従）の更新
- **発生元**: dist-architecture (20260708_114151_initial_arch)
- **種別**: RDRA修正
- **提案内容**: docs/rdra/latest/システム概要.json の system_overview に旧アーキテクチャの記述が残存: (1)「ワークフローエンジンdigdagで複数シナリオを一括自動実行」(2)「ディレクトリ構造からワークフロー定義を自動生成」(3)「webhookで外部システムへ通知」(4)「ログ追従表示やdigdag Web UIで実行状況を確認」。情報.tsv・条件.tsv・状態.tsv は v1.0（内蔵ランナー・OTelトレース・digdag/webhook廃止）に更新済みで、システム概要のみ不整合。RDRA 変更イベントとして system_overview を v1.0 の記述（Go単一バイナリ・内蔵ランナー・OTLPトレース・stfw status / report / 静的HTMLレポート）へ更新することを推奨。latest の直接書き換えは行っていない。
- **根拠**: 情報.tsv・条件.tsv・状態.tsv は v1.0 に更新済みで、システム概要.json のみ旧アーキテクチャ記述が残存していたため
- **影響範囲**: docs/rdra/latest/システム概要.json（system_overview のみ。他 RDRA モデル・アーキテクチャ設計への影響なし）
- **推奨対応**: [x] requirements スキル再実行で反映 / [ ] 却下 / [ ] 保留
- **ステータス**: 解決済み（event: 20260708_120928_update_system_overview で対応。arch:20260708_121250_arch_user_confirm でユーザー確認済み）

## 2026-07-08 AS-BUILT 作成時に検出した実装ギャップ

### DIST-002: ファイルログ（.stfw/stfw.log・日次ローテーション・terminal カラー）が未実装
- **発生元**: docs/AS-BUILT.md 作成時の実装照合
- **種別**: 実装ギャップ（要求は正・実装が未達）
- **提案内容**: USDM REQ-006（ログ仕様は v0.2 から維持）と arch SR-002/CTR-003/CLP-003 は「.stfw/stfw.log への集約・日次ローテーション・terminal 実行時カラー出力」を定めるが、Go 実装は slog を stderr へ出力するのみ（internal/presentation/cli/root.go）。M1 で console 出力のみとした残債。logger/masker の経路は整備済みのため、ファイル出力ハンドラの追加で対応可能
- **影響範囲**: internal/presentation/logger/（実装のみ。要求・モデルの変更は不要）
- **推奨対応**: [ ] 実装バックログ / [x] 要求側を stderr のみに緩和 / [ ] 保留
- **ステータス**: 解決済み（要求撤回。プロダクトオーナー判断で「実装しない」を決定。stderr 構造化ログ + HTML レポート + OTLP トレースで運用要求を満たす。event: usdm/rdra 20260721_210722_retract_filelog_timezone_reflect_parallel で REQ-006 を改訂）

### DIST-003: stfw.timezone が env 公開のみで実装未参照
- **発生元**: docs/AS-BUILT.md 作成時の実装照合
- **種別**: 実装ギャップ
- **提案内容**: stfw.yml の timezone は env（stfw_timezone）として公開されるのみで、Go コードは未参照。ジャーナル・レポートの時刻はプロセスのローカル TZ になる。要求どおり業務日付判定・レポート時刻表記に使うなら time.LoadLocation での適用実装が必要
- **影響範囲**: internal/（ジャーナル ts・レポート表示・OTel スパン時刻の TZ 統一）
- **推奨対応**: [ ] 実装バックログ / [x] 要求側を「env 公開のみ」に緩和 / [ ] 保留
- **ステータス**: 解決済み（要求撤回。プロダクトオーナー判断で「実装しない」を決定。時刻はプロセスのローカルタイムゾーンが正。event: usdm/rdra 20260721_210722_retract_filelog_timezone_reflect_parallel で REQ-026 を新設）

## 2026-07-21 dist-architecture からの追加提案

### DIST-004: arch 未追従の RDRA 要素 14 件（過去イベント由来）の反映
- **発生元**: dist-architecture (20260721_212545_arch_update_for_retract_filelog_timezone_reflect_parallel)
- **種別**: Arch追加
- **提案内容**: coverage-report.md の RDRA 網羅率が 84%（73/87）。未カバー 14 件は本イベント（parallel/ログ/timezone）とは無関係で、arch 未更新のまま RDRA だけ先行した過去イベント（20260709 remote_housekeep / 20260711 spec_roundtrip_invoke_config / 20260712 warn_first_class_status / 20260713 partial_run）由来: 情報 2 件（シナリオ spec（scenario.yml）・シナリオドキュメント（scenario.md））、条件 12 件（sshExec のリモートスクリプト一括実行 / scpPut の原子的配置 / 実行結果ハウスキープ / scaffold の差分同期（--sync） / spec/doc 往復の可逆性 / invoke エビデンス HTML レポート生成の優先順 / config の ${...} 環境変数展開 / 実行ステータス集約 / run 終了コードの集約 / Warn ステータスの後方互換 / 部分実行の指定制約 / 部分実行のスキップ意味論）。別イベントとして SP/SR の追加・改訂で反映すること（本イベントは変更要望のスコープ外のため見送り）
- **根拠**: 旧 coverage-report（2026-07-08 時点の RDRA 72 要素に対して 100%）が stale で、最新 RDRA での再生成により顕在化した既存ギャップ。今回の変更要望のスコープ（要求撤回・parallel 反映）外のため、無関係な箇所の書き換えを避けて見送った
- **影響範囲**: docs/arch/latest/arch-design.yaml（SP/SR/CTR の追加・改訂のみ。構造変更・実装変更は不要。実装は AS-BUILT に反映済み）
- **推奨対応**: [x] dist-architecture の別イベントで反映 / [ ] 却下 / [ ] 保留
- **ステータス**: 解決済み（event: arch:20260722_025855_arch_update_for_dist004_backlog で対応。構造不変のまま SP-006 / SR-008〜SR-015 / SR-104〜SR-106 / E-026〜E-027 を追加し、RDRA 網羅率 84%（73/87）→ 100%（87/87）、NFR 網羅率 100%（44/44）維持）

