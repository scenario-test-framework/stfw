# 変更サマリ

- event_id: 20261125_044114_compare_regex_process_type
- 元USDM: 20261125_044114_compare_regex_process_type
- 生成日時: 2026-11-25T04:41:14

## 追加

- 情報: 比較対象設定（compare_target.csv）（REQ-028/SPEC-028-02。コンテキスト: シナリオ構造管理。compareRegex のプロセスディレクトリ配下 config/compare_target.csv・git 管理・必須。書式は compare-files の比較対象設定ファイルそのまま（CSV・UTF-8・# 始まりはコメント。1 行 = 左ディレクトリ（expect/{収集系 process 名}/…）, 右ディレクトリ（actual/{収集系 process 名}/…）, ファイル名正規表現）。探索は指定ディレクトリの直下のみ・名前順で最初に一致した 1 件を比較・CSV に書かなかったファイルは比較しない。関連情報: プロセス、期待値（expect）、エビデンス、比較結果）
- 条件: compareRegex 比較の素通り防止（REQ-028/SPEC-028-04。コンテキスト: シナリオ構造管理。比較せずに合格するのを防ぐため、比較対象設定が無い・データ行 0 件・BOM 付き UTF-8・左右のディレクトリ指定の不正（右に expect/・.. で戻る・process 名が無い）・1 行 = 1 レコードにならない書き方（行末でクォートが閉じない、行の途中の CR）・3 列でない行・on_mismatch の明示的な空文字・result/ が出ていない・古い actual/・result/ を消せない、を Error（exit 6）にする。バリエーション「プロセスタイプ」・状態モデル「ステップ実行ステータス」参照）
- BUC: シナリオ一括自動実行フロー → UC「プロセスを実行する」に情報「比較対象設定（compare_target.csv）」と条件「compareRegex 比較の素通り防止」の関連行を追加（REQ-028/SPEC-028-02/04）

## 変更

- バリエーション: プロセスタイプ → 値に compareRegex を追加し、検証系が compare（相対パスの完全一致）と compareRegex（ファイル名正規表現）を含むこと・比較方法をディレクトリ名で読めるよう別タイプとする（compare は変更しない）旨を説明に追記（REQ-028/SPEC-028-01）
- バリエーション: compare on_mismatch 設定 → compareRegex も同じ設定キー・同じ意味で共有する旨と、compareRegex では明示的な空文字も設定不正（exit 6）とする旨を説明に追記（マージキー保護のため名称は維持）（REQ-028/SPEC-028-03）
- 条件: エビデンスディレクトリ規約（プロジェクト環境管理）→ compareRegex は compare と同じ expect/・actual/・result/ に加えて比較対象設定 config/compare_target.csv（git 管理・必須）を持つ旨を追記（REQ-028/SPEC-028-01/02）
- 条件: エビデンスディレクトリ規約（シナリオ構造管理）→ 上と同じ追記（同名の 2 行が食い違わないよう両方を更新）（REQ-028/SPEC-028-01/02）
- 条件: 比較不一致はステップ失敗 → 対象を compare / compareRegex とし、compareRegex（compare_regex）も終了コードの対応付け（0/3/6）と on_mismatch の扱いが compare と同じ旨を追記（REQ-028/SPEC-028-03）
- 情報: プラグイン → 組込みプラグイン群の検証系に compareRegex を追加し、正規表現指定比較・比較対象設定による組み合わせ・compare との共有事項（on_mismatch・終了コード・プロジェクト共通の比較レイアウト / 起動設定・compare_files_version）・stfw plugin install compareRegex によるプロビジョニング（stfw:full は同梱版）を追記（REQ-028/SPEC-028-01/05/06）
- 情報: Process / Plugin 設定（config.yml） → 検証系スキーマを compare / compareRegex に拡張し compare_files_version（両者で一致させる。compareRegex の上書きはプロジェクト共通 config.yml）を属性に追加。比較レイアウト・起動設定のプロジェクト共通を compare と共有し config/plugins/process/compareRegex/ は設定チェーン専用である旨を追記（REQ-028/SPEC-028-03/05）
- 情報: 期待値（expect） → 関連情報に「比較対象設定（compare_target.csv）」を追加し、compareRegex の比較基準でもあること（左ディレクトリ直下のファイルを正規表現で選ぶ）を追記（REQ-028/SPEC-028-01/02）
- 情報: 比較結果 → compare / compareRegex（compare_files / compare_regex）の比較結果とし、compareRegex の古い actual/・result/ の作り直しと result/ が空なら Error の旨を追記（REQ-028/SPEC-028-01/04）
- 外部システム: compare-files（ファイル比較 OSS） → 利用元を compare / compareRegex とし、compare_files（ディレクトリ比較）と compare_regex（正規表現指定比較）の使い分け、版の共有、stfw:full での両バイナリ同梱と版の食い違い時のビルド失敗を役割に追記（REQ-028/SPEC-028-01/05）
- BUC: プロセスプラグイン拡張フロー → UC「プロセスプラグインを管理する」（情報: プラグインの関連行）の説明の検証系に compareRegex を追加し、stfw plugin install compareRegex（stfw:full は同梱版）を追記（REQ-028/SPEC-028-06）
- BUC: シナリオ一括自動実行フロー → UC「プロセスを実行する」（情報: プロセスの関連行）の説明の検証系に compareRegex を追加し、on_mismatch の共有と比較対象設定による組み合わせ・素通りの Error 化を追記（REQ-028/SPEC-028-01/02/03/04）

## 削除

- なし
