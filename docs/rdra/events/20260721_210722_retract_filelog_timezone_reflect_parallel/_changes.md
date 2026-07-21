# 変更サマリ

- event_id: 20260721_210722_retract_filelog_timezone_reflect_parallel
- 元USDM: 20260721_210722_retract_filelog_timezone_reflect_parallel
- 生成日時: 2026-07-21T21:07:22

## 追加

- 条件: parallel 子プロセスの並走実行（組込み parallel プロセスの子プロセス並走のセマンティクス: 子 1 件 = ジャーナル 1 ステップ、フックは親のみ、子同士に Blocked 無し、worse-wins 集約（Error > Warn > Success）、max_parallel の設定チェーン解決、部分実行は親まで）
- バリエーション: parallel 同時実行数設定（max_parallel: 0 = 上限なし（同梱デフォルト）/ 正整数 = 上限 / 負数・非整数 = 設定不正）
- BUC: シナリオ一括自動実行フロー → UC「プロセスを実行する」に条件「parallel 子プロセスの並走実行」の関連行を追加

## 変更

- 情報: 実行ログ → ファイルログ（.stfw/stfw.log・日次ローテーション・terminal カラー出力）の撤回を反映し、stderr への構造化ログ（slog + シークレットマスキング）へ改訂
- 情報: プロジェクト設定（stfw.yml） → timezone を「env 公開のみの任意キー（実装は参照しない。時刻はプロセスのローカルタイムゾーンが正）」に改訂、process.parallel.max_parallel（既定 0）を属性に追加、バリエーションに「parallel 同時実行数設定」を追加
- 情報: プラグイン → 組込みプラグイン群に並走系 parallel（子プロセス並走）を追加し、子 env の複製・上書きとエビデンス分離を追記
- 情報: プロセス → 属性に子プロセス（parallel タイプ配下。親と同形式・入れ子禁止・子 0 件は定義不正）を追加、子への validate 適用を追記
- 情報: Process / Plugin 設定（config.yml） → 並走系スキーマ（parallel: max_parallel）を追加、バリエーションに「parallel 同時実行数設定」を追加
- 情報: 実行ジャーナル（journal.jsonl） → steps_enumerated のステップ粒度（scripts はスクリプト名 / parallel は子ディレクトリ名 = 子 1 件で 1 ステップ）と parallel の記録規則を追記
- 条件: 逐次実行・エラー時 Blocked → 「実行モデルは逐次実行。例外として組込み parallel プロセス配下の子プロセスのみ並走する（子同士に Blocked は無い）」を追記
- バリエーション: プロセスタイプ → 値に parallel を追加し、並走系の説明を追記
- 状態: ステップ実行ステータス（「」→ Pending の列挙） → parallel では子プロセス 1 件を 1 ステップとして子ディレクトリ名昇順に Pending 列挙する旨を追記
- 状態: ステップ実行ステータス（Pending → Blocked） → parallel の子同士に Blocked は無い旨を追記
- BUC: プロジェクト初期化フロー → アクティビティ「プロジェクト設定（stfw.yml）を編集する」の説明を timezone 撤回（env 公開のみの任意キー・ローカルタイムゾーンが正）に合わせて改訂
- BUC: シナリオ一括自動実行フロー → UC「プロセスを実行する」の説明（情報: プロセスの関連行）に組込み parallel の並走を追記
- BUC: 実行結果監視・確認フロー → アクティビティ名を「レポート・ログファイルで失敗箇所を調査する」から「レポート・ログで失敗箇所を調査する」へ変更し、説明を stderr への構造化ログ閲覧（ファイルログは要求撤回により存在しない）に改訂

## 削除

- なし
