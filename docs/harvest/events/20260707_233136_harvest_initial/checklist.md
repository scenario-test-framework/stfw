# dist-harvest 進捗チェックリスト

event_id: `20260707_233136_harvest_initial`
対象: /Users/suwa_sh/src/github.com/scenario-test-framework/stfw

## Phase0: 入力確認

- [x] 対象リポジトリパスの確認
- [x] 既存 RDRA チェック（docs/rdra/latest/ なし → 初期構築モード）
- [x] event_id 採番
- [x] sources.md 記録
- [x] checklist.md 初期化

## Phase1: リポジトリ解析

- [x] phase1-overview → analysis/01-overview.md（low 1 件, FIXME 2 件）
- [x] phase2-value → analysis/02-value.md（low 3 件: 要求理由の推測）
- [x] phase3-environment → analysis/03-environment.md（low 3 件, FIXME 3 件）
- [x] phase4-boundary → analysis/04-boundary.md（low 1 件, FIXME 3 件）
- [x] phase5-internal → analysis/05-internal.md（low 4 件, as-is バグ疑い 2 件）
- [x] 整合性チェック（8 観点、矛盾なし。未参照要素 2 件を FIXME 記録）
- [x] analysis/ → docs/harvest/latest/ コピー

## Phase2: USDM 逆生成

- [x] docs/usdm/events/{event_id}/requirements.yaml 生成（要求 8 / 仕様 21 / affected_models 98）
- [x] docs/usdm/events/{event_id}/source.txt 生成
- [x] validateRequirements.js PASS（初回）
- [x] docs/usdm/latest/requirements.yaml + requirements.md 生成

## Phase3: ユーザー確認

- [x] confidence: low 項目の提示・確認（10 項目提示、全項目 Option A（推奨案）で承認、2026-07-08）

## Phase4: RDRA フルビルド

- [x] RDRA Phase1〜5 + 統合（Phase1: 11 タスク / Phase2: 3 / Phase3: 3 / Phase4: 4 / Phase5: 1）
- [x] 関連データ.txt / ZeroOne.txt 生成
- [x] generateRdraMd.js 1_RDRA --lint エラー 0 件（初回 4 件 → 状態.tsv 遷移UC 分割・情報.tsv 関連情報名の修正で解消）
- [x] docs/rdra/latest/ + docs/rdra/events/{event_id}/ 配置（10 ファイル）
- [x] views/ 生成（不整合チェック 0 件）
- [x] 一時ディレクトリ削除

## 完了報告

- [x] サマリ・confidence: low 一覧の報告
