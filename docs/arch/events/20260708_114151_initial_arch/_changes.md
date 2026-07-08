# 変更サマリ

- event_id: 20260708_114151_initial_arch
- trigger_event: rdra:20260708_103305_builtin_plugin_ecosystem, nfr:20260708_112906_nfr_user_confirm
- モード: 初期構築（全要素を「追加」として記録）

## 追加

### domain_architecture

- subdomains: SD-001 シナリオ構造管理（core）/ SD-002 実行管理（core）/ SD-003 通知管理（supporting）/ SD-004 プロジェクト環境管理（supporting）
- bounded_contexts: BC-001 scenario / BC-002 run / BC-003 notify / BC-004 project（RDRA 4 コンテキスト = 4 BC、実装 internal/domain/ と一致）
- context_map: CM-001（run→scenario, Customer-Supplier）/ CM-002（notify→run, Published Language: journal イベント）/ CM-003（run→project, Customer-Supplier）/ CM-004（scenario→project, Customer-Supplier）
- aggregate_hypotheses: AG-001 Run 集約（user 確定）/ AG-002 シナリオ集約仮説（low）/ AG-003 プロジェクト集約仮説（low）

### system_architecture

- tiers: tier-cli（CLI 本体）/ tier-plugin（プラグイン実行）/ tier-file-datastore（ファイルデータストア）/ tier-report-delivery（レポート配信）/ tier-distribution（配布・CI）
- cross_tier_policies: CTP-001〜008（認証認可 OS 委譲・OTLP トレース一本化・互換境界維持・マスキング・新 run_id 再実行・i18n 非対応・性能前提・セキュリティ責務分担）
- cross_tier_rules: CTR-001〜003（SSH known_hosts 検証・OTel 送信抑制/非致命・構造化ログ集約）
- tier 個別 policies/rules: SP-001〜005, SR-001〜006（CLI）/ SP-101〜107, SR-101〜103（プラグイン）/ SP-201〜203, SR-201〜202（データストア）/ SP-301〜302, SR-301（レポート）/ SP-401〜403, SR-401〜403（配布・CI）

### app_architecture

- tier_layers: tier-cli の 5 層（presentation / usecase / domain / repository / gateway、domain 依存ゼロ・IF なし直接依存）
- tier_layers: tier-plugin の 2 層（プラグイン契約層 / 外部ツール実行層）
- cross_layer_policies/rules: CLP-001〜003, CLR-001〜003（CLI）/ CLP-101, CLR-101〜102（プラグイン）

### data_architecture

- entities: E-001〜E-025（情報.tsv の全 25 情報と 1:1）
- storage_mapping: 全 25 件（file 23 件 / cache 2 件。外部データストアなしの as-is 制約）

## 変更

- なし（初期構築）

## 削除

- なし（初期構築）
