# NFR 推論根拠サマリ（差分更新・文言追従）

- event_id: 20260722_025511_nfr_wording_followup_user_approved
- created_at: 2026-07-22T02:55:11
- trigger_event: nfr:20260721_211823_nfr_update_for_retract_filelog_timezone_reflect_parallel
- model_system: model1（変更なし。単一利用者の CLI ツールという性質は不変）

## 本イベントの位置づけ

前イベント 20260721_211823（RDRA 差分への追従）で、confidence: user の 2 メトリクスは
保護ルールにより文言追従を見送り、確認推奨項目として返却した。
今回、**ユーザー（プロダクトオーナー）が当該 2 項目の文言更新を明示承認**したため、
承認を根拠に保護ルールの例外として文言のみ追従する。再推論は行わない
（Lv・confidence・規模判断はユーザー確定値のまま）。

## 文言追従の根拠（実装の正本 docs/AS-BUILT.md との突き合わせ）

| ID | メトリクス | Lv | confidence | 追従内容と根拠 |
|----|----------|-----|-----------|------|
| C.6.1.1 | ログ保管期間 | 1（不変） | user（不変） | source_model の「日次ローテーション」はファイルログ要求撤回（rdra:20260721_210722）により実装に存在しない。現行実装では保管期間は実行成果物（実行ジャーナル `.stfw/runs/{run_id}` + HTML レポート `.stfw/reports/runs/{run_id}.html`）の run 開始時ハウスキープ（stfw.yml の `stfw.housekeep.retention_days`。AS-BUILT §5.6）で担保し、stderr へ出力される実行ログはファイル保管しないため保管は呼び出し元の terminal / CI に委ねる。この実態へ reason・source_model を改訂 |
| B.1.1.4 | バッチ処理件数 | 1（不変） | user（不変） | 実行モデルは逐次実行だが、組込み parallel プロセス配下の子プロセスのみ単一実行内で並走する（AS-BUILT §4.14、条件「parallel 子プロセスの並走実行」）。grade_description の「逐次実行」をこの実態へ改訂。規模判断（GB オーダー以内）はユーザー確定値のまま不変 |

## グレード判断の方針

- 変更は説明文言（grade_description / reason / source_model）のみ。Lv 値は全メトリクスで据え置き
- confidence: user は維持する（ユーザー確定値であることに変わりはない）

## ユーザー確認による変更

- ユーザー（プロダクトオーナー）が C.6.1.1 / B.1.1.4 の文言更新を明示承認
  （前イベントの確認推奨項目への回答。詳細は _changes.md「ユーザー承認の記録」）

## 確認推奨項目

- なし（前イベントの確認推奨 2 件は本イベントで解消。confidence: low の項目は 0 件）
