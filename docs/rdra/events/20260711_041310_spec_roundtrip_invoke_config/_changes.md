# 変更サマリ

- event_id: 20260711_041310_spec_roundtrip_invoke_config
- 元USDM: 20260711_041310_spec_roundtrip_invoke_config
- 生成日時: 2026-07-11T04:20:35

## 追加

- 情報: シナリオ spec（scenario.yml）（REQ-020/SPEC-020-01/02/03。シナリオツリーを機械可読化した単一ファイル。reverse で生成、scaffold でツリー再生成・差分同期。骨格が往復可逆）
- 情報: シナリオドキュメント（scenario.md）（REQ-020/SPEC-020-01。reverse が spec とセット生成する人間可読ドキュメント。group / type / description・要求トレーサビリティ・config サブツリーを表形式出力）
- 条件: scaffold の差分同期（--sync）（REQ-020/SPEC-020-02。stfw scaffold <spec.yml> [--sync] で spec からツリー骨格生成。既存ツリーは --sync で差分同期、--sync 無しはエラー）
- 条件: spec/doc 往復の可逆性（REQ-020/SPEC-020-03。reverse → scaffold → reverse で骨格が完全一致。data CSV・スクリプト・expect 等の葉ファイルは対象外）
- 条件: invoke エビデンス HTML レポート生成の優先順（REQ-021/SPEC-021-02。k6 web dashboard が十分なデータでレポート生成なら採用、不足時は summary.json から自己完結 HTML をフォールバック生成。k6 失敗でも report.html は常に残る。バリエーション「プロセスタイプ」参照）
- 条件: config の ${...} 環境変数展開（stfw.yml 値参照）（REQ-022/SPEC-022-01。config チェーンの ${VAR} は環境変数参照。run 開始時に stfw.yml のフラット化を export（v0.2 の export_yaml 互換）するため ${stfw_...} で stfw.yml 値を参照でき、共通 identity を単一ソース化できる）

## 変更

- BUC: テストシナリオ作成フロー → UC「シナリオを spec・ドキュメントに変換する」（reverse。情報「シナリオ spec（scenario.yml）」「シナリオドキュメント（scenario.md）」「メタ情報（metadata.yml）」「シナリオ」、条件「spec/doc 往復の可逆性」、画面「シナリオ往復生成CLI」）と UC「spec からシナリオツリーを生成する」（scaffold。情報「シナリオ spec（scenario.yml）」「シナリオ」「業務日付」「プロセス」「メタ情報（metadata.yml）」「Process / Plugin 設定（config.yml）」、条件「scaffold の差分同期（--sync）」「spec/doc 往復の可逆性」、画面「spec scaffold生成CLI」）を追加。単一ソースでの版管理・共有・再生成の受益者アクティビティを追加
- BUC: シナリオ一括自動実行フロー → UC「シナリオを実行する」に情報「プロジェクト設定（stfw.yml）」（run 開始時の stfw.yml フラット化 export）を追加。UC「プロセスを実行する」に条件「invoke エビデンス HTML レポート生成の優先順」「config の ${...} 環境変数展開（stfw.yml 値参照）」を追加し、プロセス実行の説明に invoke 系のエビデンス出力（summary.json / report.html）と config の ${...} 環境変数展開を追記
- アクター: シナリオ作成者 → stfw scenario reverse / stfw scaffold（--sync）で spec・ドキュメントを往復生成し、シナリオ構造を単一ファイルで版管理・共有・再生成できる旨を役割に追記
- アクター: テスト結果確認者 → invoke 系（invokeRest / invokeWeb）の Act 結果をエビデンス（k6 サマリ summary.json・HTML レポート report.html）で確認できる旨を役割に追記
- 情報: シナリオ → 関連情報に「シナリオ spec（scenario.yml）」「シナリオドキュメント（scenario.md）」を追加。reverse / scaffold による往復可逆生成を説明に追記
- 情報: メタ情報（metadata.yml） → requirement_specifications が reverse のドキュメント要求トレーサビリティ（どの要求をどの process が検証するか）として表形式出力される旨を追記
- 情報: プラグイン → 実行系（invokeWeb / invokeRest）の Act 後エビデンス出力（summary.json / report.html）と、config チェーンの ${VAR} 環境変数展開・${stfw_...} による stfw.yml 値参照を説明に追記
- 情報: プロジェクト設定（stfw.yml） → 属性に共通設定値（db.database / db.user 等の共通 identity）、run 開始時のフラット化 export（export_yaml 互換）と ${stfw_...} 参照・identity 単一ソース化を説明に追記
- 情報: Process / Plugin 設定（config.yml） → 属性・説明に値中の ${VAR} 環境変数展開（${stfw_...} で stfw.yml 値参照、例 ${stfw_db_database} / ${stfw_db_user}）を追加。集約対象は identity（database / user）のみで接続情報（ホスト・パスワード）は inventory + secret の禁止契約を維持する旨を追記。関連情報に「プロジェクト設定（stfw.yml）」を追加
- 情報: エビデンス → 属性に invoke 系（invokeWeb / invokeRest: evidence/summary.json・evidence/report.html）の配置を追加。実行系のエビデンス出力がエビデンスディレクトリ規約（第 3 の互換境界）の対象に含まれる旨を説明に追記
- 外部システム: grafana k6 → Act 後の end-of-test サマリ（summary.json）出力と web dashboard の report.html 生成（データ不足時フォールバック）、k6 失敗時も report.html が残る旨を役割に追記
- バリエーション: プロセスタイプ → 実行系（invokeWeb / invokeRest）が Act の結果を evidence/（summary.json / report.html）へエビデンス出力する旨を説明に追記
- 条件: エビデンスディレクトリ規約 → 実行系（invokeWeb / invokeRest）のエビデンス出力（summary.json / report.html）を対象に含める旨を追記
- 条件: プラグイン接続情報のグループ名参照 → 共通 identity（database / user）は stfw.yml 集約・${stfw_...} 参照可だが、接続情報（ホスト・パスワード）の config 直書き解禁ではなく inventory + secret の禁止契約を維持する旨を追記

## 削除

- なし
