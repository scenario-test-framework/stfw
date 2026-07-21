# 変更サマリ

- event_id: 20260721_211823_nfr_update_for_retract_filelog_timezone_reflect_parallel
- trigger_event: rdra:20260721_210722_retract_filelog_timezone_reflect_parallel
- 生成日時: 2026-07-21T21:18:23

グレード値（Lv）の変更は 1 件も無い。ファイルログ・timezone 要求の撤回と組込み parallel
プロセスタイプの反映（RDRA 差分）に対する、記述（grade_description / reason / source_model）の
追従のみ。NFR グレードの引き上げは変更要望どおり行わない。

## 追加

- なし

## 変更

- 性能・拡張性/リソース拡張性/B.3.1.1 CPU拡張性: Lv1 のまま。reason・source_model を
  「逐次実行が仕様」から「逐次実行 + 組込み parallel 配下の子プロセスのみ単一実行内で並走」へ改訂
  （条件「parallel 子プロセスの並走実行」を根拠に追加）。スケールアウトは引き続き要求しない
- 運用・保守性/ログ管理/C.6.1.2 ログ種別: Lv1 のまま。grade_description・reason・source_model を
  ファイルログ（.stfw/stfw.log・日次ローテーション）撤回に合わせ「stderr への構造化ログ（slog）+
  実行ジャーナル」へ改訂

## 削除

- なし

## 変更なしと判断した関連項目

- timezone 撤回: NFR グレードに timezone / タイムゾーンを参照するメトリクスは存在せず影響なし
  （A.1.1.1 運用時間等は時間帯の要求であり、タイムゾーン解決方式に依存しない）
- C.3.3.1 障害復旧方式: 記述「レポート・ログで失敗箇所を調査」は改訂後の BUC 文言と既に一致
- B.1.x / B.2.x（業務処理量・性能目標値）: parallel は単一実行内の子プロセス並走であり、
  同時アクセス・スループットのスケールは対象外のため据え置き
- F.2.2.1 アーキテクチャ拡張性: 「プラグイン機構によるプロセスタイプ拡張」の記述は
  parallel（組込みプラグイン追加）と整合しており据え置き

## ユーザー確定値（confidence: user）保護によりスキップした項目

- C.6.1.1 ログ保管期間（Lv1・1ヶ月）: source_model「情報: 実行ログ（日次ローテーション）」が
  撤回済み記述として残存するが、ユーザー確定値保護ルールにより差分更新では上書きしない。
  保管期間 1 ヶ月の判断自体は実行成果物（ジャーナル・レポート）の housekeep.retention と整合
- B.1.1.4 バッチ処理件数（Lv1）: grade_description に「逐次実行」の文言が残るが、
  規模判断（GB オーダー以内）は不変のためユーザー確定値保護によりスキップ
