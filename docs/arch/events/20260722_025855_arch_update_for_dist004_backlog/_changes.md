# 変更サマリ

- event_id: 20260722_025855_arch_update_for_dist004_backlog
- trigger_event: rdra:20260709_071603_remote_housekeep, rdra:20260711_041310_spec_roundtrip_invoke_config, rdra:20260712_015850_warn_first_class_status, rdra:20260713_130252_partial_run
- 背景: docs/todo.md DIST-004（arch 未追従の RDRA 要素 14 件）。実装はすべて実装・リリース済みで、実装契約の正本は docs/AS-BUILT.md。アーキテクチャの構造（ティア・レイヤー・BC・集約）は変更しない

## 追加

### system_architecture / tier-cli

- policies: SP-006 シナリオ構造の spec / doc 往復（単一ファイル運用）
  - カバー: 情報「シナリオ spec（scenario.yml）」「シナリオドキュメント（scenario.md）」
- rules: SR-008 実行ステータスと終了コードの集約
  - カバー: 条件「実行ステータス集約（Error > Warn > Success）」「run 終了コードの集約（0/3/6）」（AS-BUILT §4.6）
- rules: SR-009 Warn ステータスの後方互換
  - カバー: 条件「Warn ステータスの後方互換（旧ジャーナル混在）」（AS-BUILT §4.6）
- rules: SR-010 実行結果ハウスキープ
  - カバー: 条件「実行結果ハウスキープ」（AS-BUILT §5.6）
- rules: SR-011 部分実行の指定制約
  - カバー: 条件「部分実行の指定制約」（AS-BUILT §3.4）
- rules: SR-012 部分実行のスキップ意味論
  - カバー: 条件「部分実行のスキップ意味論」（AS-BUILT §3.4）
- rules: SR-013 scaffold の差分同期（--sync）
  - カバー: 条件「scaffold の差分同期（--sync）」（AS-BUILT §12）
- rules: SR-014 spec / doc 往復の可逆性
  - カバー: 条件「spec/doc 往復の可逆性」（AS-BUILT §12）
- rules: SR-015 config の ${...} 環境変数展開
  - カバー: 条件「config の ${...} 環境変数展開（stfw.yml 値参照）」（AS-BUILT §8.2）

### system_architecture / tier-plugin

- rules: SR-104 sshExec のリモートスクリプト一括実行
  - カバー: 条件「sshExec のリモートスクリプト一括実行」（AS-BUILT §4.13）
- rules: SR-105 scpPut の原子的配置
  - カバー: 条件「scpPut の原子的配置」（AS-BUILT §4.13）
- rules: SR-106 invoke エビデンス HTML レポート生成の優先順
  - カバー: 条件「invoke エビデンス HTML レポート生成の優先順」（AS-BUILT §4.12）

### data_architecture

- entities: E-026 シナリオ spec（scenario.yml）（storage: file）
- entities: E-027 シナリオドキュメント（scenario.md）（storage: file）
- storage_mapping: E-026 / E-027（file）

## 変更

- system_architecture/tiers/tier-cli: description に scaffold / scenario コマンドを追記
- system_architecture/tiers/tier-plugin: description の組込みプラグイン群にリモートアクセス系を追記
- domain_architecture/bounded_contexts/BC-001: owned_entity_ids に E-026 / E-027 を追記、ubiquitous_language に「spec / doc」を追加（RDRA 情報.tsv のコンテキスト割当「シナリオ構造管理」に従う。既存のユーザー確定内容（BC 境界・既存語彙・既存割当）は不変・追記のみ。confidence: user は維持）
- app_architecture/tier_layers/tier-cli:
  - L-cli-usecase: responsibility に scenariodoc（spec / doc 生成）を追記
  - L-cli-repository: responsibility に scenariospec / scenariodoc を追記
  - L-cli-gateway: responsibility に mdwriter を追記
  - diagram_mermaid: gateway ノードに mdwriter を追記
- data_architecture/diagram_mermaid: SCENARIO_SPEC / SCENARIO_DOC の関連を追加

## 削除

- なし

## 網羅率

- RDRA: 84%（73/87）→ 100%（87/87）
- NFR（重要メトリクス）: 100%（44/44）維持
- BC への entity 割当率: 100%（27/27）維持
