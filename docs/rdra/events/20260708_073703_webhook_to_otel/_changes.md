# 変更サマリ

- event_id: 20260708_073703_webhook_to_otel
- 元USDM: 20260708_073703_webhook_to_otel
- 生成日時: 2026-07-08T07:46:34

対応要求: REQ-005（廃止・REQ-009 に置換）、REQ-009（OTLP トレースによる実行結果通知）、REQ-010（webhook 通知機能の完全廃止）

## 追加

- 情報: OTel トレース（スパンツリー）（属性: trace_id、スパンツリー（run=ルート、scenario / bizdate / process=子、step=末端）、スパン属性（run_id・階層タイプ・bizdate・seq・group・プロセスタイプ・終了コード等）、スパンステータス、開始・終了時刻。実行ジャーナルのイベントの投影。コンテキスト: 通知管理）
- 情報: 実行ジャーナル（journal.jsonl）（属性: イベント行（各階層・ステップの開始・終了イベント、イベント時刻、ステータス）。実行状況イベントの唯一のソース。コンテキスト: 実行管理）
- 外部システム: OTLP 受信先（OpenTelemetry Collector / 互換バックエンド）（外部システム群: オブザーバビリティ基盤。Jaeger / Grafana Tempo / Datadog 等でそのまま可視化・分析）
- バリエーション: スパン階層タイプ（値: run、scenario、bizdate、process、step。コンテキスト: 通知管理）
- バリエーション: OTel エクスポート設定（値: OTEL_EXPORTER_OTLP_ENDPOINT（環境変数）、stfw.otel.endpoint（stfw.yml）。コンテキスト: 通知管理）
- 条件: スパンステータス・属性マップ（実行コンテキストのスパン属性への引き継ぎ、実行ステータス Error → スパンステータス Error、Blocked のスパン属性表現。コンテキスト: 通知管理）
- 条件: OTel エクスポート先未設定時の送信抑制（環境変数・stfw.yml とも未設定の場合はトレースを送信しない。コンテキスト: 通知管理）
- 条件: OTel エクスポート失敗の非致命扱い（送信失敗は実行を失敗させず実行ログへの警告記録のみ。コンテキスト: 通知管理）

## 変更

- アクター: テスト結果確認者 → 実行状況の確認手段を webhook 通知から OTel トレース（既存オブザーバビリティ基盤での可視化・分析）に変更
- BUC: 実行結果監視・確認フロー → アクティビティ「webhook通知で進捗・成否を把握する」を「OTelトレースで進捗・成否を把握する」に置換。UC「実行状況を通知する」の関連を OTel トレース（スパンツリー）・実行ジャーナル（journal.jsonl）・スパンステータス・属性マップ・OTel エクスポート先未設定時の送信抑制・OTel エクスポート失敗の非致命扱いに変更し、イベント「実行進捗・成否通知」（webhook 受信先）を「実行トレース送信」（OTLP 受信先）に置換
- BUC: シナリオ一括自動実行フロー → 「各階層のsetup/teardownを実行させる」「プロセス配下スクリプトを逐次実行させる」の関連情報 webhook payload を実行ジャーナル（journal.jsonl）に変更し、削除された条件「webhook status 決定」への参照を除去（webhook payload 削除に伴う参照整理）
- 情報: プロジェクト設定（stfw.yml） → 属性から webhooks.urls / on_start / on_success / on_error を廃止し、otel.endpoint（バリエーション: OTel エクスポート設定）を追加
- 情報: ステップ実行結果 → 記録先を webhook payload から step スパンの属性（OTLP トレース）・実行ジャーナルに変更
- 情報: 実行ログ → OTel エクスポート失敗の警告記録を属性・説明に追加
- 情報: シナリオ → 関連情報の webhook payload を OTel トレース（スパンツリー）に変更（webhook payload 削除に伴う参照整理）
- 情報: 業務日付 → 関連情報の webhook payload を OTel トレース（スパンツリー）に変更（同上）
- 情報: プロセス → 関連情報の webhook payload を OTel トレース（スパンツリー）に変更（同上）
- 情報: 実行（run） → 関連情報の webhook payload を実行ジャーナル（journal.jsonl）・OTel トレース（スパンツリー）に変更（同上）
- 状態: 階層実行ステータス → 遷移契機の説明を webhook start / end 通知から実行ジャーナルへのイベント記録に変更し、永続化先を OTLP トレースのスパンステータス・属性に変更
- 状態: ステップ実行ステータス → 終了状態（Success / Error / Blocked）の記録先を process の webhook payload から step スパンの属性（OTLP トレース）に変更

## 削除

- 情報: webhook payload
- 外部システム: webhook 受信先
- バリエーション: 階層タイプ（webhook type）
- バリエーション: webhook イベント種別
- バリエーション: webhook 通知設定
- 条件: webhook 送信判定
- 条件: webhook status 決定
