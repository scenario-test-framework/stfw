# アーキテクチャ推論根拠サマリ

- event_id: 20260721_212545_arch_update_for_retract_filelog_timezone_reflect_parallel
- created_at: 2026-07-21T21:25:45

差分更新イベント。RDRA 差分（20260721_210722_retract_filelog_timezone_reflect_parallel）と
NFR 差分（20260721_211823_nfr_update_for_retract_filelog_timezone_reflect_parallel）に関連する
項目のみ再評価した。アーキテクチャの構造（ティア・レイヤー・BC・集約）は不変。

## RDRA/NFR 差分の分析結果

### 分析した RDRA 差分要素

| モデル | 差分 | 影響したアーキテクチャ項目 |
|--------|------|--------------------------|
| 条件: parallel 子プロセスの並走実行（追加） | 子並走のセマンティクス全体 | SR-007（新設）、CTP-007、technology_context.constraints、E-011 |
| バリエーション: parallel 同時実行数設定（追加） | max_parallel の設定チェーン | SR-007、E-003（process_parallel_max_parallel） |
| 条件: 逐次実行・エラー時 Blocked（変更） | 逐次実行 + parallel 例外の明記 | technology_context.constraints、tier-cli technology_candidates |
| 情報: 実行ログ（変更） | ファイルログ撤回 → stderr 構造化ログ | SR-002、CTR-003、CLP-003、E-017、storage_mapping/E-017 |
| 情報: プロジェクト設定（stfw.yml）（変更） | timezone は env 公開のみの任意キー | E-003 |
| 情報: プロセス（変更） | 子プロセス（parallel 配下）の追加 | E-011、data ER 図 |

### 参照した NFR 差分

| メトリクス | Lv | 影響 |
|-----------|----|------|
| B.3.1.1 CPU拡張性 | 1（変更なし） | CTP-007 の記述追従（逐次実行 + parallel 子並走はスケールアップ前提のまま） |
| C.6.1.2 ログ種別 | 1（変更なし） | SR-002 / CTR-003 / CLP-003 の記述追従（stderr 構造化ログ + ジャーナル） |
| C.6.1.1 ログ保管期間 | 1（変更なし・user 保護でNFR側は未改訂） | SR-002 に保管の担保手段（ハウスキープ）を明記して網羅を維持 |

## 設計判断サマリ

| 対象 | 判断 | confidence | 根拠 |
|------|------|-----------|------|
| SR-002 / CTR-003 / CLP-003 | ファイルログ（.stfw/stfw.log・日次ローテーション・カラー）を撤回し stderr 構造化ログへ一本化 | user | 変更要望 §1（プロダクトオーナー決定）。実装契約 AS-BUILT §2.1 |
| E-003.timezone | env 公開のみの任意キー（実装は参照しない。ローカルタイムゾーンが正） | user | 変更要望 §2。RDRA 情報: プロジェクト設定（stfw.yml） |
| SR-007（新設） | parallel 子プロセス並走のセマンティクスを tier-cli のルールとして明文化 | user | 変更要望 §3。実装契約 AS-BUILT §4.14 |
| CTP-007 | parallel はスケールアウトと区別し NFR B.x は据え置き | user | 変更要望 §3 の位置づけ（NFR グレード引き上げは意図しない） |
| storage_mapping/E-017 | file → cache（stderr ストリーム。永続化なし） | user | E-019（OTel トレース）と同じ「非永続の出力」分類に合わせた |

## ユーザー確認による変更

非対話実行（変更要望に明示された決定のみを反映。推論による新規判断は行っていない）。

| 対象 | 項目 | 旧値 | 新値 | 変更理由 |
|------|------|------|------|---------|
| SR-002 | 実行ログ運用 | .stfw/stfw.log 集約・日次ローテーション・カラー | stderr 構造化ログ + ハウスキープで保管担保 | 変更要望 §1（DIST-002 撤回） |
| CTR-003 | 名称・内容 | 構造化ログの単一ファイル集約 | 構造化ログの stderr 出力 | 変更要望 §1 |
| CLP-003 | ロギング方針 | ファイル出力（ローテーション・カラー） | stderr 出力 | 変更要望 §1 |
| E-003.timezone | 説明 | timezone | env 公開のみの任意キー（実装は参照しない） | 変更要望 §2（DIST-003 撤回） |
| constraints | 実行モデル | v1.0 は逐次実行のみ（将来 --parallel の余地） | 逐次実行 + parallel 子プロセス並走（実装済み） | 変更要望 §3 |
| CTP-007 | 並列実行の位置づけ | 将来の並列実行はランナーのシナリオ単位分離で余地を残す | 実装済み parallel とスケールアウトを区別 | 変更要望 §3 |

## confidence 内訳（本イベントで出力した要素のみ）

| セクション | high | medium | low | default | user | 合計 |
|-----------|:----:|:------:|:---:|:-------:|:----:|:----:|
| システムアーキテクチャ | 0 | 0 | 0 | 0 | 4 | 4 |
| アプリケーションアーキテクチャ | 0 | 0 | 0 | 0 | 1 | 1 |
| データアーキテクチャ（storage_mapping） | 0 | 0 | 0 | 0 | 1 | 1 |
| 合計 | 0 | 0 | 0 | 0 | 6 | 6 |
