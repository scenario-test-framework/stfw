# 変更サマリ

- event_id: 20260708_134805_plugin_contracts
- 元USDM: 20260708_134805_plugin_contracts
- 生成日時: 2026-07-08T13:55:58

対応要求: REQ-012（collectLog / collectFile）、REQ-013（export / import / clear: MySQL / PostgreSQL / Redis）、REQ-014（compare による期待値比較）、REQ-016（横断規約: 接続情報のグループ名参照・エビデンスディレクトリ規約・ランタイム依存宣言）

補足:

- 本イベントは、名前だけ定義されていた契約 3 点（エビデンスディレクトリ規約・プラグイン設定スキーマ・データ形式）とランタイム依存宣言（plugin.yml の requires）の具体規約確定であり、既存の条件・情報の説明/属性の詳細化（変更）のみで構成される
- 新規要素の追加・既存要素の削除はない。BUC・アクター・状態・バリエーション・外部システムに内容変更はないため差分 TSV は収録しない

## 追加

- なし

## 変更

- 情報: パスワード → データストア系プラグインが {host}-{user} を自動参照し、config.yml への直接記述は禁止であることを明記
- 情報: プラグイン → 属性にメタデータファイル plugin.yml（requires = 前提コマンドのリスト。例: mysql, scp, k6）を追加し、env 契約の stfw_bizdate_start_ts を RFC3339 形式として明記。requires は stfw validate と run 前静的検証がシナリオで使用するプロセスタイプ単位でコマンド存在チェックすることを追記
- 情報: Process / Plugin 設定（config.yml） → 設定スキーマを具体化（収集系: targets リスト（group = inventory グループ名参照、paths = 収集ファイルパス正規表現リスト）、データストア系: host_group・port・database・user・tables 等。パスワードは secret の {host}-{user} を自動参照）。関連情報にパスワードを追加
- 情報: 実行ジャーナル（journal.jsonl） → bizdate node_start イベント時刻の env 公開名を stfw_bizdate_start_ts（RFC3339 形式）として確定
- 情報: 期待値（expect） → expect/ は git 管理で、直下に同一 bizdate 内の収集系 process ディレクトリ名を置き、その配下は当該 process の evidence/ 配下と同型であることを具体化
- 情報: エビデンス → 出力ルートを自プロセスディレクトリ配下の evidence/（gitignore 対象）とし、配置規約（collectFile / collectLog: {host}/{収集元の絶対パス}、exportMysql / exportPostgres: {database}/{table}.csv、exportRedis: {host}/{keyパターン名}.csv）と CSV データ形式（RFC 4180・ヘッダー行・NULL は \N・Redis は key,type,ttl,value + 正規化 JSON）を具体化
- 情報: 比較結果 → actual/ は expect と同じ構造で各収集系 process の evidence/ への symlink（自動生成）、result/ は compare-files の比較結果出力であることを具体化（「symlink または一時コピー」の選択肢を symlink に確定）
- 条件: run 前静的検証 → ランタイム依存の存在チェックを「シナリオで使用するプロセスタイプが plugin.yml の requires に宣言したコマンド」として具体化
- 条件: エビデンスディレクトリ規約 → 具体的なディレクトリ命名規約を確定（「具体的な命名は設計フェーズで確定」を解消）
- 条件: プラグイン接続情報のグループ名参照 → 設定スキーマ（収集系 targets / データストア系 host_group 等）と secret {host}-{user} の自動参照を具体化
- 条件: プラグインのランタイム依存宣言と存在チェック → 宣言先をメタデータファイル plugin.yml の requires として確定し、検証対象を「シナリオで使用するプロセスタイプの requires」に具体化
- 条件: export/import ラウンドトリップ互換 → データ形式を具体化（CSV: RFC 4180 準拠・LF・UTF-8・必要時 quote・1 行目ヘッダー、NULL: \N、Redis: ヘッダー付き CSV（key,type,ttl,value）・string は生値・コレクション型はキー順ソートの正規化 JSON）
- 条件: 収集先ホストとの時刻同期前提 → フィルタ基準時刻の env 公開名を stfw_bizdate_start_ts（RFC3339 形式）として確定

## 削除

- なし
