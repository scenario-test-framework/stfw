# 変更サマリ

- event_id: 20260712_015850_warn_first_class_status
- 元USDM: 20260712_015850_warn_first_class_status
- 生成日時: 2026-07-12T02:10:37

## 追加

- 状態: 階層実行ステータス Started→Warn（REQ-023/SPEC-023-01/02/05。階層setup/teardown・プロセス実行の両遷移UC。配下に Warn あり・Error なしで Warn に確定。Warn 終了状態の行も追加し、teardown フック env stfw_run_status=Warn を記載）
- 状態: ステップ実行ステータス Pending→Warn（REQ-023/SPEC-023-01。リターンコード 3 で Warn として記録して続行。on_mismatch: warn の compare 比較不一致を含む。Warn 終了状態の行も追加）
- 条件: 実行ステータス集約（Error > Warn > Success）（REQ-023/SPEC-023-02。上位階層 process / bizdate / scenario / run は配下の実行ステータスを Error > Warn > Success の優先度で集約。状態モデル「階層実行ステータス、ステップ実行ステータス」参照）
- 条件: run 終了コードの集約（0/3/6）（REQ-023/SPEC-023-03。全 Success=0 / Warn あり・Error なし=3 / Error あり=6。CI で「差分あり」を終了コードで検知。バリエーション「終了コード」参照）
- 条件: Warn ステータスの後方互換（旧ジャーナル混在）（REQ-023/SPEC-023-04。旧バージョンの run のジャーナル（Warn なし）と混在しても report 再生成・status 表示・ハウスキープが壊れない）
- バリエーション: compare on_mismatch 設定（REQ-014/SPEC-014-02。値: error（既定）、warn。error は compare-files exit 3 を exit 6 に変換してステップ失敗（従来挙動）、warn は exit 3 のまま Warn として記録・続行）

## 変更

- 状態: 階層実行ステータス → Started→Success の 2 行（階層setup/teardown・プロセス実行）を「配下に Warn・Error なし」の条件つきに更新し、プロセス実行の Started→Error を「リターンコード 0・3 以外」に更新（3 は Warn へ分離）
- 状態: ステップ実行ステータス → Pending→Success（compare 一致は on_mismatch 設定によらず Success）、Pending→Error（リターンコード 0・3 以外。on_mismatch: error（既定）の比較不一致は exit 3→6 変換で Error）、Pending→Blocked（先行の Error のみが対象で Warn は Blocked を発生させない）を更新
- 条件: 逐次実行・エラー時 Blocked → Error（終了コード 0・3 以外）は停止 + 後続 Blocked、Warn（終了コード 3）は記録して続行に改訂（REQ-023/SPEC-023-01、REQ-014/SPEC-014-02）
- 条件: 比較不一致はステップ失敗 → on_mismatch: error（既定）| warn の選択制に改訂（マージキー保護のため名称は維持）。error は従来挙動（exit 3→6 変換・停止・Blocked 伝播）、warn は exit 3 のまま Warn 記録・続行。バリエーション「compare on_mismatch 設定」を参照に追加（REQ-014/SPEC-014-02）
- 条件: スパンステータス・属性マップ → Warn は OTel スパンステータスに相当が無い（Ok / Error / Unset のみ）ため、スパンステータス Ok + stfw の status 属性で表現する旨を追記（REQ-024/SPEC-024-03）
- 条件: 実行結果ハウスキープ → 旧バージョンの run のジャーナル（Warn なし）が混在しても保存期間判定・削除が壊れない旨を追記（REQ-023/SPEC-023-04）
- バリエーション: 終了コード → 3（WARN）は記録して続行・0・3 以外は Error 停止の実行意味論と、stfw run 自体の終了コード集約（0/3/6）による CI 差分検知を説明に追記（REQ-023/SPEC-023-01/03）
- 情報: 実行ジャーナル（journal.jsonl） → status に Warn を追加（属性に明記）。旧ジャーナル混在でも report 再生成・status 表示・ハウスキープが壊れない旨を追記（REQ-023/SPEC-023-01/02/04）
- 情報: ステップ実行結果 → result に Warn を追加（Pending / Success / Warn / Error / Blocked）。終了コード基準を 0=Success / 3=Warn / 0・3 以外=Error に更新し、Warn はスパンステータス Ok + status 属性として投影される旨を追記（REQ-023/SPEC-023-01、REQ-024/SPEC-024-03）
- 情報: プラグイン → RC 契約（0/3/6）は不変のまま 3（Warn）に実行意味論を追加（記録して続行。「止まらなくなる」挙動変化はリリースノートで明示）。env 契約に stfw_run_status（Warn あり・Error なしの run で Warn）を追加し、compare の on_mismatch 設定を追記（REQ-023/SPEC-023-01/05、REQ-014/SPEC-014-02）
- 情報: 実行（run） → 属性に終了コード（全 Success=0 / Warn あり・Error なし=3 / Error あり=6）を追加し、Error > Warn > Success 集約と CI 差分検知を説明に追記（REQ-023/SPEC-023-02/03）
- 情報: HTML レポート → Warn の黄系表示と「比較 NG の鳥瞰」ビュー、旧ジャーナル混在でも再生成が壊れない旨を追記（REQ-024/SPEC-024-02、REQ-023/SPEC-023-04）
- 情報: OTel トレース（スパンツリー） → スパン属性に status 属性（Warn 表現）を追加し、Warn はスパンステータス Ok + status 属性・Error は従来どおりスパンステータス Error である旨を追記（REQ-024/SPEC-024-03）
- 情報: Process / Plugin 設定（config.yml） → 属性に検証系スキーマ（compare: on_mismatch = error（既定）| warn）を追加し、バリエーション「compare on_mismatch 設定」を参照に追加（REQ-014/SPEC-014-02）
- 情報: 比較結果 → 比較不一致の扱いを on_mismatch 選択制（error=Error 停止・Blocked 伝播 / warn=Warn 記録・続行して鳥瞰）に改訂し、バリエーション「compare on_mismatch 設定」を参照に追加（REQ-014/SPEC-014-01/02）
- 情報: スクリプト（ステップ） → 説明のステップ実行ステータス遷移を Pending→Success/Warn/Error/Blocked に更新（REQ-023/SPEC-023-01。外部レビュー指摘による追補）
- アクター: テスト実行者 → 回帰テストモードと機能変更の差分確認モードの 2 運用の選択と、CI での終了コード（0/3/6）による差分検知を役割に追記（REQ-023/SPEC-023-03）
- アクター: テスト結果確認者 → Warn の確認手段（status / HTML レポートの黄系表示、OTel の Ok + status 属性）と HTML レポートの「比較 NG の鳥瞰」ビュー利用を役割に追記（REQ-024/SPEC-024-01/02/03）
- 外部システム: compare-files（ファイル比較 OSS） → 比較不一致時は exit 3 を返し、compare プラグインの on_mismatch 設定で Error 変換（既定）/ Warn 続行として扱われる旨を役割に追記（REQ-014/SPEC-014-02）
- BUC: シナリオ一括自動実行フロー → UC「シナリオを実行する」に条件「run 終了コードの集約（0/3/6）」を追加。UC「階層setup/teardownを実行する」の説明を Started→Success/Warn/Error・Error > Warn > Success 集約・stfw_run_status=Warn に更新し、条件「実行ステータス集約（Error > Warn > Success）」を追加。UC「プロセスを実行する」の説明を RC 3=Warn 続行・Pending→Success/Warn/Error/Blocked・on_mismatch 選択制に更新し、条件「実行ステータス集約（Error > Warn > Success）」を追加。受益者アクティビティに 2 運用モードと CI 終了コード検知を追記（REQ-023/SPEC-023-01〜05、REQ-014/SPEC-014-02）
- BUC: 実行結果監視・確認フロー → UC「実行状況を通知する」の説明に Warn のスパンステータス Ok + status 属性表現を追記。UC「実行状況を確認する」「HTMLレポートを生成する」の説明に Warn の黄系表示・鳥瞰ビュー・旧ジャーナル混在の後方互換を追記し、両 UC に条件「Warn ステータスの後方互換（旧ジャーナル混在）」を追加（REQ-024/SPEC-024-01/02/03、REQ-023/SPEC-023-04）

## 削除

- なし
