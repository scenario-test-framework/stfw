# CLAUDE.md

## このリポジトリの正体

**stfw**（scenario test framework）— 業務日付をまたぐシナリオテストをディレクトリ規約で記述し、
単一バイナリで自動実行する Go 製 CLI。実行エンジン（逐次実行・Blocked 伝播・ジャーナル・HTML レポート・
OTLP トレース）を内包する。Apache License 2.0 の OSS。

- 配布: マルチプラットフォームバイナリ（GitHub Releases）+ Docker イメージ（`ghcr.io/scenario-test-framework/stfw:latest` / `:full`）。
- v0.2 系（Bash + digdag）はタグ `v0.2.0` で凍結済み。v1.0 は全面 Go 再実装。

## 開発プロセス（必ずこの順で）

このリポジトリは **distillery（要件・仕様パイプライン）を上流の正本**とし、**実装契約は
`docs/AS-BUILT.md`（as-built）を正本**とする仕様駆動開発で運用する。コードを直接変更する前に、
まず仕様・契約側から入る。

### 1. 要件・設計は distillery / as-built で確認する

- 上流の要求・要件（イベントソーシング。`events/` は不変、`latest/` が最新スナップショット）:

  | パス | 内容 |
  |---|---|
  | `docs/usdm/latest/` | USDM 要求仕様 |
  | `docs/rdra/latest/` | RDRA モデル（アクター/情報/状態/条件/バリエーション/BUC） |
  | `docs/nfr/latest/` | IPA 非機能要求グレード |
  | `docs/arch/latest/` | アーキテクチャ設計（レイヤー・依存ルール・ADR） |
  | `docs/harvest/latest/` | 旧 Bash 版からの as-is 抽出 |

- **実装契約の正本は [`docs/AS-BUILT.md`](docs/AS-BUILT.md)**（コマンド契約・env 契約・状態モデル・
  プラグイン契約・ディレクトリ規約・往復セマンティクス）。シナリオ作成手順は [`docs/GUIDE.md`](docs/GUIDE.md)。
- 仕様・契約に無いものを実装で発明しない。実装中に不足・矛盾を見つけたら、コードでごまかさず
  `docs/AS-BUILT.md`（必要なら distillery の該当 `latest/`）へ戻して更新してから実装する。
- 機能を追加・変更したら、`docs/AS-BUILT.md` の該当節・`README.md` / `README.ja.md` の図表を必ず追従させる。

### 2. 互換境界を壊さない（最優先の制約）

v1.0 は次の **3 つの互換境界**を維持する。ここを変えるとプラグイン・既存プロジェクトが壊れる。

1. **ディレクトリ規約** `scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/`
   （bizdate = 8 桁 `YYYYMMDD` / group に `_` 禁止 / seq は数値、昇順逐次実行）
2. **プロセスプラグイン実行契約**: 入力 = env（`stfw_*` フラット化 + 実行コンテキスト）、
   出力 = リターンコード `0`=Success / `3`=Warn / `6`=Error、実装言語は任意
3. **エビデンスディレクトリ規約**（`data/` 入力・`evidence/` 出力・`expect/` 期待値の配置）

### 3. レイヤー構成と依存方向（domain は依存ゼロ）

```
presentation (cli / logger)
  → usecase (initialize / scaffold / validate / runscenario / status / report /
             inventory / secret / sshtrust / plugin / scenariodoc)
    → domain (scenario / run / notify / project)   ← 依存ゼロの純粋ロジック
    → repository (journal / scenariotree / config / secret / inventory / plugin /
                  scaffold / report / metadata / processconfig / scenariospec / scenariodoc)
      → gateway (scriptexec / webhookhttp / sshkeyscan / htmlwriter / mdwriter)
```

- 依存方向: presentation → usecase → domain / repository、repository → domain / gateway。
  **domain はどこにも依存しない**（値オブジェクトで関所を一本化。`NewX() (X, error)` で検証）。
- 詳細は `docs/arch/latest/` と `docs/design/architecture/app_arch.puml`。

### 4. テストは受け入れ → ユニットの順で実効性を担保する

- **受け入れテスト（testscript / `.txtar`）**: `test/acceptance/testdata/script/*.txtar`。
  コマンドの外部契約（stdout/stderr/exit code/生成物）を golden で固定する。挙動を変えたら
  該当 txtar を必ず更新する。
  - **ファイル名がサブテスト名になる**（`TestAcceptance/{filename}`）。ユニットと同じ書式
    **`対象_XXXの場合_YYYであること`** で命名する（例: `run_実行する場合_公開envが契約どおりであること.txtar`）。
    リネーム時は `docs/AS-BUILT.md` の「根拠」参照も追従させる。
- **ユニットテスト**: ドメインルール・repository の境界。I/O 境界は**実体でテストする**
  （`t.TempDir()` の実ファイル・実バイナリ。モックで誤魔化さない）。
  - **構造は AAA パターン**（Arrange / Act / Assert）。各ケースを `// Arrange` `// Act`
    `// Assert` のコメントで 3 ブロックに区切る。準備・実行・検証を混在させない。
  - **ケース名は `t.Run` のサブテスト名に付ける**。書式は
    **`テスト対象_XXXの場合_YYYであること`**（例: `NewBizdate_7桁の場合_エラーであること`）。
    テスト関数は `TestXxx` のまま（Go 識別子に日本語を使わない）。1 ケース = 1 `t.Run`。
- **バグ修正は再現テストを先に書く**（RED を確認してから修正）。

### 5. 品質ゲート（commit 前に必ず緑）

```bash
go build ./...
go vet ./...
gofmt -l internal/ cmd/          # 出力が空であること
go test ./...                    # 単体 + testscript 受け入れ
golangci-lint run                # CI と同じ v2 設定（.golangci.yml）
```

- lint は **golangci-lint v2**（CI は `golangci-lint-action@v8` + `version: v2.12.2`）。
  設定は `.golangci.yml`（v2 フォーマット。`errcheck` の `fmt.Fprint*` 除外など）。
- go.mod は `go 1.25.0` + `toolchain go1.26.4`（依存 otel/x-crypto/grpc が 1.25 を要求）。
  golangci-lint はビルド Go が go.mod ターゲット未満だと拒否するので、両者のバージョン整合を保つ。

### 6. 作業単位ごとのレビュー（サブエージェント → Codex の二段）

機能単位の Definition of Done として、実装が一区切りしたら次を実施する。

- **(a) サブエージェントレビュー**: 生成した本人とは別のサブエージェントに、契約
  （`docs/AS-BUILT.md`）と突き合わせたレビューをさせる（仕様トレーサビリティ / クラッシュ耐性・
  冪等性 / テストの実効性）。
- **(b) Codex レビュー（codex-refute）**: `codex-refute` スキルで外部レビューを回し、指摘ごとに
  実体（コード・テスト・契約）と照合して**一度反証**する。反証しきれない指摘のみ修正する。
- **(c) 取り込み**: 修正は回帰テスト追加 → 再テスト → 品質ゲート確認まで。反証内訳（指摘数 /
  不採用と根拠 / 対応数）をコミットメッセージまたは PR に残す。

## 実装規約

- **コメントは日本語**で書く（仕様の制約・設計判断を示す最小限のもの）。
  **コード・識別子・エラーメッセージ・ログは英語**。
- doc コメントも日本語で、対象名から始める（`// FuncName は〜する。`）。
- パスワード等の秘匿値は **env 経由**で渡す（`SSHPASS` / `REDISCLI_AUTH` / `stfw_target_password`）。
  **argv に渡さない**。接続先ホスト・パスワードを config に直書きしない（inventory + secret で解決）。

## 横断的な注意点

- **CI**（`.github/workflows/ci.yml`）: `lint`（golangci-lint v2）+ `test`（go test）。
  GitHub Actions はアクションをバージョン参照し、`permissions` は最小に保つ。
- **リリース**（`.github/workflows/release.yml`）: タグ `v*` push で goreleaser がバイナリ +
  checksums を Releases へ、Docker イメージ（`stfw` / `stfw:full`）を ghcr.io へ push する。
  - **リリース手順**: ① master の CI 緑を確認 → ② 版数フォールバック
    `internal/presentation/cli/root.go` の `Version = "X.Y.Z-dev"` を**リリースする版**へバンプし、
    `docs/AS-BUILT.md` の「未注入時 `X.Y.Z-dev`」の記載も追従 → ③ `git tag vX.Y.Z && git push origin vX.Y.Z`。
    リリースノートは goreleaser がコミットから自動生成（`feat:`/`fix:` でグルーピング、`docs:` は除外）。
    CHANGELOG ファイルは持たない。
  - **挙動変化（後方互換に影響する変更）を含むリリース**: リリースノートに載るのは
    **コミットタイトルのみ**のため、①挙動変化はコミットタイトル自体に含める、
    ② goreleaser のリリース作成後に `gh release edit vX.Y.Z --notes-file ...` で
    「挙動変化」節（変化の内容・影響対象・確認事項）を先頭に追記する。
- **examples/daily-balance を変更したら**: ① `./run.sh` で end-to-end 確認（`sut/schema.sql` を
  変えた場合は先に `./run.sh --down` で DB を作り直す。schema は postgres コンテナ初期化時のみ実行）、
  ② `stfw scenario reverse daily-balance` で `stfw/docs/daily-balance.{yml,md}` を再生成
  （spec/doc はツリーからの生成物。手で編集しない）、③ example README と `docs/GUIDE.md` §3 の
  ツリー・表を追従。
- **README は bilingual**: `README.md`=英語 / `README.ja.md`=日本語（バッジ直下に相互リンク）。
  内部ドキュメント（`docs/`）とコメントは日本語。
- プラグインバイナリ（k6 / compare-files）はイメージに同梱せず、`stfw plugin install {type}` で
  取得する（`stfw run` は自動インストールしない）。
