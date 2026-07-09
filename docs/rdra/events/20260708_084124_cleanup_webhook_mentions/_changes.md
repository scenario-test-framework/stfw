# 変更サマリ

- event_id: 20260708_084124_cleanup_webhook_mentions
- 生成日時: 2026-07-08T08:42:20
- 元USDM: なし（要求変更を伴わない自由記述の掃除。20260708_073703_webhook_to_otel の残作業）

## 追加

- なし

## 変更

- 情報: プラグイン → 属性の「bin/（install・run・webhook）」を「bin/（install・run）」に修正（webhook 詳細契約は REQ-010 で廃止済み）
- BUC: プロジェクト初期化フロー → アクティビティ「プロジェクト設定（stfw.yml）を編集する」の説明を「webhook URL・…」から「OTelエンドポイント・…」に修正（stfw.webhooks.* は REQ-010 で廃止、stfw.otel.endpoint に置換済み）

## 削除

- なし
