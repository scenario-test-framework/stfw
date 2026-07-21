# NFR 推論根拠サマリ（差分更新）

- event_id: 20260721_211823_nfr_update_for_retract_filelog_timezone_reflect_parallel
- created_at: 2026-07-21T21:18:23
- trigger_event: rdra:20260721_210722_retract_filelog_timezone_reflect_parallel
- model_system: model1（変更なし。単一利用者の CLI ツールという性質は不変）

## 再推論の入力（RDRA 差分）

| RDRA 変更 | 内容 | 影響した NFR メトリクス |
|---|---|---|
| 情報: 実行ログ | ファイルログ（.stfw/stfw.log・日次ローテーション・カラー出力）撤回 → stderr 構造化ログ（slog + マスキング） | C.6.1.2（記述追従）、C.6.1.1（user 保護でスキップ） |
| 情報: プロジェクト設定（stfw.yml） | timezone を env 公開のみの任意キーへ改訂（ローカルタイムゾーンが正） | なし（NFR に timezone 参照なし） |
| 条件: 逐次実行・エラー時 Blocked / parallel 子プロセスの並走実行 | 逐次実行の例外として組込み parallel 配下の子プロセス並走を追加 | B.3.1.1（記述追従）、B.1.1.4（user 保護でスキップ） |

## 再推論結果（変更メトリクス）

| ID | メトリクス | Lv | confidence | 根拠 |
|----|----------|-----|-----------|------|
| B.3.1.1 | CPU拡張性 | 1（不変） | medium | 条件「parallel 子プロセスの並走実行」は単一実行内の子プロセス並走であり、1 run = 単一ホスト 1 プロセスは不変。性能向上はスケールアップで対応（スケールアウト要求なし）のため Lv1 を維持し、reason の「逐次実行が仕様」を並走例外を含む表現へ改訂 |
| C.6.1.2 | ログ種別 | 1（不変） | high | 情報「実行ログ」が stderr への構造化ログ（slog）へ改訂されたため、grade_description のファイルログ記述（.stfw/stfw.log）を追従。ログ 2 種（実行ログ + ジャーナル）の構成は不変のため Lv1 を維持 |

## グレード判断の方針

- 変更要望に「NFR グレードの引き上げは意図しない」と明記されているため、全メトリクスで
  Lv 値は据え置き。parallel による同時アクセス・スループットのグレード変更は行わない
- 変更は RDRA 差分に追従する記述（grade_description / reason / source_model）のみ

## ユーザー確認による変更

- なし（非対話実行。confidence: user のメトリクスはイベントソーシングルールに従いスキップ）

## 確認推奨項目

- C.6.1.1 ログ保管期間（confidence: user）: source_model の「日次ローテーション」参照が
  撤回済み記述として残存。保管期間 Lv1（1ヶ月）の判断は不変のため、参照文言の更新要否のみ
  次回のユーザー対話で確認を推奨
- B.1.1.4 バッチ処理件数（confidence: user）: grade_description の「逐次実行」文言の更新要否を
  同様に確認を推奨
