# 変更サマリ

- event_id: 20260721_212545_arch_update_for_retract_filelog_timezone_reflect_parallel
- trigger_event: rdra:20260721_210722_retract_filelog_timezone_reflect_parallel, nfr:20260721_211823_nfr_update_for_retract_filelog_timezone_reflect_parallel

要求撤回（ファイルログ DIST-002 / timezone DIST-003）と、v1.5.0 で実装・リリース済みの
組込み parallel プロセスタイプの反映。アーキテクチャの構造（ティア・レイヤー・BC）は不変で、
記述整合のみを行う。実装契約の正本は docs/AS-BUILT.md（§4.14 parallel / §2.1 ログ）。

## 追加

- system_architecture/tiers/tier-cli/rules: SR-007「parallel 子プロセスの並走実行」
  （子 1 件 = 1 ステップ / フックは親のみ / 子同士に Blocked 無し / worse-wins 集約 /
  max_parallel の設定チェーン解決 / 部分実行は親まで / 互換境界不変）
- data_architecture/entities/プロセス（E-011）: 属性 children（parallel タイプ配下の子プロセス）と
  自己参照リレーション（親 1:N 子）を追加
- data_architecture/entities/プロジェクト設定（stfw.yml）（E-003）: 属性
  process_parallel_max_parallel（同梱デフォルト 0 = 上限なし）を追加
- data_architecture/diagram_mermaid: `PROCESS ||--o{ PROCESS : parallel_children` を追加

## 変更

- technology_context/constraints: 「v1.0 は逐次実行のみ（将来 --parallel の余地をランナー分離で残す）」
  → 「実行モデルは逐次実行（1 実行 = 1 プロセス）。例外として組込み parallel プロセスタイプにより
  子プロセスの並走を選択できる（max_parallel で制御。互換境界は不変）」
- system_architecture/tiers/tier-cli/technology_candidates: 「内蔵実行エンジン（木構造の逐次実行）」
  → 「（木構造の逐次実行 + parallel 子プロセス並走）」、「構造化ログ（ローカルファイル出力）」
  → 「構造化ログ（stderr 出力）」
- system_architecture/tiers/tier-cli/rules/SR-002 実行ログ運用: .stfw/stfw.log への集約・
  日次ローテーション・terminal カラー出力を撤回し、stderr への構造化ログ（マスキング済み）へ改訂。
  ログ保管期間（NFR C.6.1.1・1 ヶ月）は実行ジャーナル・HTML レポートのハウスキープ
  （stfw.housekeep.retention）で担保する旨を明記
- system_architecture/cross_tier_policies/CTP-007: 「将来の並列実行はランナーのシナリオ単位分離で
  余地を残す」を削除し、実装済み parallel（単一実行内の子プロセス並走）とスケールアウト
  （NFR B.x の対象・引き続きスコープ外）を区別する記述へ改訂
- system_architecture/cross_tier_rules/CTR-003: 名称「構造化ログの単一ファイル集約」
  → 「構造化ログの stderr 出力」。.stfw/stfw.log への集約を撤回し stderr 出力へ改訂
- app_architecture/tier_layers/tier-cli/cross_layer_policies/CLP-003 ロギング方針:
  「.stfw/stfw.log へ出力（日次ローテーション・terminal 実行時カラー）」→ 「stderr へ出力
  （マスキング済み）。ファイルログは要求撤回により持たない」
- data_architecture/entities/プロジェクト設定（stfw.yml）（E-003）: timezone 属性を
  「env 公開のみの任意キー（実装は参照しない。時刻はプロセスのローカルタイムゾーンが正）」へ改訂
- data_architecture/entities/実行ログ（E-017）: 属性を stderr 構造化ログへ改訂
  （log_file（.stfw/stfw.log・日次ローテーション）・colored を削除し、log_stream を追加。
  run（E-015）へのリレーションを追加）
- data_architecture/storage_mapping/E-017: storage_type file（.stfw/stfw.log・日次ローテーション）
  → cache（stderr ストリーム出力のみ・ファイル永続化なし）
- decisions/arch-decision-006: consequences.negative の「v1.0 は逐次実行のみ（並列実行は将来の
  拡張余地として残す）」を、実装済み parallel の実態へ改訂（artifact_id 単位の upsert）

## 削除

- なし（要素の削除は無し。SR-002 / CTR-003 / CLP-003 / E-017 は同一 ID の内容改訂）

## diff 表記の注意

- arch-design-diff.yaml でスキーマ必須のため併記した項目は「変更なし」を意味する:
  tier-cli の `policies: []` / `name` / `description`、tier_layers/tier-cli の
  `cross_layer_rules: []` / `diagram_mermaid` / `layers`（L-cli-presentation を無変更のまま
  1 件のみ収載。他レイヤーも無変更）。latest へのマージでは既存要素を削除・置換しない
  （マージは各要素の id 単位。削除は本ファイルの「削除」節のみを正とする）。

## confidence: user の上書きについて

SR-002 / CTR-003 / CLP-003 / CTP-007 / SP-004 等の既存 user 確定値のうち、本イベントで
上書きしたのは SR-002 / CTR-003 / CLP-003 / CTP-007 のみ。上書きは自動再推論ではなく、
プロダクトオーナーの変更要望（.tmp/change-request-20260721.md）による明示的な決定に基づく
（新しい値も confidence: user として記録）。
