# 変更サマリ

- event_id: 20261125_052806_reflect_ci_vuln_scan
- trigger_event: rdra:20261125_052806_reflect_ci_vuln_scan, nfr:20261125_052806_reflect_ci_vuln_scan

## 追加

- `system_architecture/tiers/tier-distribution/rules` の SR-404 に PR・master の脆弱性検査とイメージの部品確認を記録し、SR-401 を参照。
- 同ティアの SR-405 に週次・手動の定期チェック、保守担当者の取り込み、失敗通知の条件と自動停止からの再開手順を記録し、SR-401 を参照。
- `decisions/arch-decision-015.yaml` にトリガーごとの検証分担を記録。

## 変更

- `system_architecture/tiers/tier-distribution` の責務に定期チェックとイメージの公開前検証を追記。
- SR-402 に両 arch のスモークテスト、digest push 後のレイヤー照合、配布用タグの付与条件を追加。
- `system_architecture/cross_tier_policies` の CTP-008 に配布物の脆弱性管理と利用先環境の責務を明記。

## 削除

- なし

## 網羅率

- RDRA 87/94 件（93%）。同じ RDRA を使った更新前と同数。本イベントの外部システム 2 件はカバー済み。
- 重要 NFR 44/44 件（100%）。RDRA の未カバー 7 件は既存の先行差分に由来する。
