# 変更サマリ

- event_id: 20261125_140812_reflect_compare_files_v230
- 元USDM: 20261125_140812_reflect_compare_files_v230
- 生成日時: 2026-11-25T14:08:12

既存 SPEC の affected_models には過去のモデル変更も含まれる。このイベントで適用する差分は以下と同梱の TSV に記載する。

## 追加

- 条件: compare / compareRegex の起動設定補完（compare-files v2.3.0 の型別補完、ローカル値の優先、キーの検証。SPEC-014-03）
- BUC: シナリオ一括自動実行フロー（既存 UC「プロセスを実行する」に条件「compare / compareRegex の起動設定補完」への関連行を追加。SPEC-014-03）

## 変更

- 条件: compareRegex の比較ルール解決 → false と空配列の保持を反映し、型別補完は共通条件を参照（SPEC-028-05）
- 外部システム: compare-files（ファイル比較 OSS） → 既定版 v2.3.0、同梱版の VERSION の共有、起動設定の補完規則を反映（SPEC-014-03・SPEC-028-05・SPEC-031-01）
- バリエーション: プロセスタイプ → compare と compareRegex のファイル組み合わせ方式の区別を維持し、compare 全体の変更有無に関する記述を整理（SPEC-028-01）
- バリエーション: Docker イメージ構成 → stfw:full に同梱する compare-files を v2.3.0 とし、両バイナリの VERSION の共有を明記（SPEC-031-01）
- 情報: Process / Plugin 設定（config.yml） → compare と compareRegex の compare_files_version の既定が v2.3.0 であることを明記（SPEC-031-01）

## 削除

- なし
