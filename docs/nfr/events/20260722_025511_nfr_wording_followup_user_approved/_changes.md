# 変更サマリ

- event_id: 20260722_025511_nfr_wording_followup_user_approved
- trigger_event: nfr:20260721_211823_nfr_update_for_retract_filelog_timezone_reflect_parallel
- 生成日時: 2026-07-22T02:55:11

グレード値（Lv）の変更は 1 件も無い。confidence も user のまま維持する。
前イベント 20260721_211823 で「ユーザー確定値（confidence: user）は差分更新で上書きしない」
保護ルールにより見送った 2 メトリクスの説明文言（grade_description / reason / source_model）のみを、
現行実装の正本 docs/AS-BUILT.md（§5.6 実行結果ハウスキープ / §4.14 組込み parallel プロセス）へ追従する。

## ユーザー承認の記録（保護ルールの例外適用）

- **ユーザー（プロダクトオーナー）が C.6.1.1 / B.1.1.4 の 2 項目の文言更新を明示承認した**
  （承認日: 2026-07-22。前イベント 20260721_211823 の確認推奨項目に対する回答）。
- 承認範囲は説明文言の追従のみ。グレード Lv の変更・confidence の変更は承認範囲外であり行わない。
- 本イベントに限り、上記承認を根拠として confidence: user 保護ルールの例外として
  latest へのマージ（文言上書き）を実施する。

## 追加

- なし

## 変更

- 性能・拡張性/業務処理量/B.1.1.4 バッチ処理件数: Lv1 のまま。grade_description の
  「逐次実行」を「逐次実行（組込み parallel 配下の子プロセスのみ単一実行内で並走）」へ改訂。
  規模判断（GB オーダー以内）は不変。source_model に条件「parallel 子プロセスの並走実行」を追加
- 運用・保守性/ログ管理/C.6.1.1 ログ保管期間: Lv1（1ヶ月）のまま。source_model に残存していた
  撤回済み記述「情報: 実行ログ（日次ローテーション）」を撤去し、現行の根拠
  （実行ジャーナル・HTML レポートの run 開始時ハウスキープ stfw.housekeep.retention_days で
  保管期間を担保。stderr ログの保管は呼び出し元の terminal / CI に委ねる）へ改訂

## 削除

- なし
