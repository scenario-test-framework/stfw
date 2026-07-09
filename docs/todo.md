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
- **推奨対応**: [x] 実装バックログ / [ ] 要求側を stderr のみに緩和 / [ ] 保留
- **ステータス**: open

### DIST-003: stfw.timezone が env 公開のみで実装未参照
- **発生元**: docs/AS-BUILT.md 作成時の実装照合
- **種別**: 実装ギャップ
- **提案内容**: stfw.yml の timezone は env（stfw_timezone）として公開されるのみで、Go コードは未参照。ジャーナル・レポートの時刻はプロセスのローカル TZ になる。要求どおり業務日付判定・レポート時刻表記に使うなら time.LoadLocation での適用実装が必要
- **影響範囲**: internal/（ジャーナル ts・レポート表示・OTel スパン時刻の TZ 統一）
- **推奨対応**: [x] 実装バックログ / [ ] 要求側を「env 公開のみ」に緩和 / [ ] 保留
- **ステータス**: open

