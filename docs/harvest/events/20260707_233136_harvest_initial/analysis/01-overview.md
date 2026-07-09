# 01 システム概要

## 解析メタ情報

| 項目 | 値 |
|------|-----|
| 解析対象リポジトリ | /Users/suwa_sh/src/github.com/scenario-test-framework/stfw |
| コミットハッシュ | ed02ba61d48212a49c416e309925bbe0ac825759 |
| ブランチ | feature/kcov |
| 解析日 | 2026-07-07 |
| フェーズ | Phase1（システム概要・技術スタック・ビジネスドメイン） |

## システム概要

- システム名: **stfw**（scenario test framework）
  - 確度: high / 根拠: 事実: README.md:1-2, src/bin/stfw:5（"scenario test framework cli"）, build/env.properties:8（GIT_URL）
  - 後段の USDM `system_name` と RDRA `システム概要.json` の `system_name` は `stfw` に統一する。
- バージョン: 0.2.0-SNAPSHOT
  - 確度: high / 根拠: 事実: src/VERSION:1
- 目的:
  - stfw は、シナリオテスト（業務日付をまたぐ一連の業務処理の検証）をディレクトリ構造として記述し、ワークフローエンジンで自動実行する CLI フレームワークである（確度: high / 根拠: 事実: CHANGELOG.md:6 "run workflow according to directory structure"、src/plugins/process/scripts/README.adoc:19 のディレクトリ構成 `scenario/{scenario}/{bizdate}/{seq}_{group}_scripts/`）。
  - テスト担当者は `stfw init` でプロジェクトを初期化し、シナリオ／業務日付／プロセスの階層にスクリプトを配置して `stfw run` で一括実行する（確度: high / 根拠: 事実: src/bin/cmd/init:25 "initialize stfw project", src/bin/cmd/run:25 "run a scenario", src/template/scenario/sample/ のサンプル構造）。
  - 実行はワークフローエンジン digdag に委譲され、各プロセスの実行結果（Pending/Started/Success/Error/Blocked）は webhook で外部システムへ通知される（確度: high / 根拠: 事実: src/bin/lib/setenv:88-91（run.dig/scenario.dig/bizdate.dig）, src/bin/lib/setenv:27-31（ステータス定数）, src/config/webhook/payload.yml:1-22）。
  - 想定ユーザー価値は「複数ホスト・複数日付にまたがるシナリオテストの再現可能な自動実行」である（確度: medium / 根拠: 推測: inventory によるホストグループ管理（src/template/config/inventory/staging.yml）と bizdate ディレクトリ構造から導出。価値を明文化したドキュメントは無い）。
- FIXME: 公式ドキュメントの Feature 説明が placeholder のまま（docs/adoc/overview/overview_index.adoc:8-10 が "AAA/BBB/CCC"）。システムの提供価値の一次記述が存在しないため、目的の一部は構造からの推測で補完した。

## 技術スタック

| 層 | 技術 | 確度 | 根拠 |
|----|------|------|------|
| CLI 本体 | Bash（サブコマンド分割型 CLI。`stfw <command>` 形式） | high | 事実: src/bin/stfw:1, src/bin/cmd/（init, run, scenario, bizdate, process, server, digdag, inventory, passwd, gen-encrypt-key） |
| アプリ構成 | Clean Architecture 簡易適用（cmd → adapter/controller → usecase → domain の層分割。interface/DI なしで直接参照） | high | 事実: docs/design/architecture/app_arch.puml:40-77, src/bin/lib/stfw/{adapter,usecase,domain} |
| ワークフローエンジン | digdag 0.9.24（Java, 同梱 jar / server モード起動） | high | 事実: src/archives/digdag-0.9.24.jar, src/bin/lib/setenv:103, src/bin/cmd/server:30-31（usage: start/stop/restart/status） |
| ランタイム依存 | Java（JDK 8）, Python 2（pyaml, docopt）, Ruby（serverspec） | high | 事実: dockerfiles/kcov/Dockerfile:44-48, .travis.yml:24,32-33, src/bin/lib/setenv:167（JAVA_HOME） |
| 補助ツール群 | Tukubai / Parsrs（シェル芸系テキスト処理コマンド）, yaml2json/json2yaml（Python） | high | 事実: src/modules/{Tukubai,Parsrs,bin,yaml2json}, src/modules/yaml2json/yaml2json:1 |
| 設定 | YAML（stfw.yml をデフォルト→プロジェクトの順に読込・環境変数へ export） | high | 事実: src/config/stfw.yml, src/bin/stfw:42-49, src/template/stfw.yml |
| データストア | RDBMS 等は不使用。digdag の状態 DB（--memory または file database）とファイルベースの内部データ（.stfw/ ディレクトリ、ログ、pid） | high | 事実: src/config/stfw.yml:11-12, src/template/stfw.yml:29-30, src/bin/lib/setenv:119-128 |
| 外部連携 | webhook（HTTP POST。start/success/error 時に JSON payload 送信） | high | 事実: src/config/webhook/payload.yml, src/template/stfw.yml:10-22 |
| セキュリティ | openssl による RSA 鍵生成 + S/MIME(AES256) でのパスワードファイル暗号化 | high | 事実: src/bin/lib/commons/bash_utils:260-299, src/bin/cmd/gen-encrypt-key:25, src/bin/cmd/passwd:25 |
| テスト（UT） | shunit2 | high | 事実: test/ut/shunit2, test/ut/ut_all.sh |
| テスト（IT） | serverspec（Ruby） + build/product/integration_test.sh | high | 事実: src/modules/serverspec/, .travis.yml:24, build/product/integration_test.sh |
| カバレッジ | kcov（Docker コンテナでビルド・実行） | high | 事実: docker-compose-kcov.yml, dockerfiles/kcov/Dockerfile, build/product/coverage.sh |
| CI | Travis CI（イベント種別 pr_created / master_pushed / tag_pushed で build/ci_event/*.sh に分岐） | high | 事実: .travis.yml:36-47, build/ci_event/{pr_created,master_pushed,tag_pushed,other}.sh |
| ドキュメント | AsciiDoc（asciidoctor でビルド, redpen で校正, gh-pages へ公開） | high | 事実: docs/adoc/index.adoc, .travis.yml:26-30, build/docs/{build.sh,publish.sh,redpen-conf.xml}, build/env.properties:12 |
| デプロイ形態 | tar.gz 配布物（stfw-with-depends-*.tar.gz）をホストに展開して利用するスタンドアロン CLI。対応 OS は Linux / macOS（cygwin はコメントアウト） | high | 事実: docker-compose-kcov.yml:9（dist/stfw-with-depends-0.2.0-SNAPSHOT.tar.gz）, src/bin/install, src/bin/lib/setenv:163-177 |
| フロントエンド | 専用 UI なし（CLI + digdag server の Web UI/API を利用） | medium | 推測: リポジトリに UI コードが無く、src/bin/cmd/server:34 の usage に "web interface and api clients" とあることから digdag 同梱 UI に依存と判断 |

## ビジネスドメイン

- ドメイン: **システム開発におけるシナリオテスト（結合・総合テスト）の自動実行基盤**
  - 確度: high / 根拠: 事実: README.md:2, src/plugins/process/scripts/README.adoc（scenario/bizdate/process の 3 階層で業務シナリオを表現）
- ドメインの中核概念（後段 Phase の入力）:
  - シナリオ（scenario）> 業務日付（bizdate）> プロセス（process）> スクリプト（step）の階層構造（確度: high / 根拠: 事実: src/template/scenario/sample/ のディレクトリ構造, src/bin/lib/setenv:88-94）
  - プロセス実行ステータス: Pending / Started / Success / Error / Blocked（確度: high / 根拠: 事実: src/bin/lib/setenv:27-31, src/plugins/process/scripts/README.adoc:43）
  - インベントリ: テスト対象ホストのグループ管理（web / ap / db 等）（確度: high / 根拠: 事実: src/template/config/inventory/staging.yml, src/bin/cmd/inventory:24-36）
- ステークホルダー:

| ステークホルダー | 直接/間接 | 説明 | 確度 | 根拠 |
|------------------|-----------|------|------|------|
| テスト実行者（テストエンジニア/開発者） | 直接 | CLI でプロジェクト初期化・シナリオ作成・実行を行う利用者 | high | 事実: docs/design/architecture/app_arch.puml:5,38（actor user → stfw）, src/bin/cmd/ の各コマンド |
| テスト対象システムの運用・管理者 | 間接 | inventory に定義された web/ap/db ホストへスクリプトを適用される側 | medium | 推測: src/template/config/inventory/staging.yml のホストグループ定義から想定。役割定義の明文化は無い |
| 結果通知の受信者（CI/チャット等の外部システム） | 間接 | webhook payload を受信して結果を監視する外部システム | high | 事実: src/config/webhook/payload.yml, src/template/stfw.yml:10-22, docs/design/architecture/app_arch.puml:8,81（other_system） |
| stfw 開発者・コントリビューター | 間接 | フレームワーク自体を保守する開発者コミュニティ | high | 事実: .github/CONTRIBUTING.md, CODE_OF_CONDUCT.md, .travis.yml:7-9（gitter 通知） |
| アクター種別（ロール・権限による区分） | - | ロール/権限定義は存在しない。CLI 実行者は単一種別 | medium | 推測: 認証・認可コードが無く、全コマンドが同一権限で実行されることから単一アクターと判断 |

- ドメイン特性:
  - **業務日付（bizdate）駆動**: シナリオを業務日付単位で区切って進行させる、バッチ処理系業務システムのテストに特化した特性を持つ（確度: high / 根拠: 事実: src/bin/cmd/bizdate:24 "scenario-bizdate control", bizdate.dig, src/config/webhook/bizdate.yml）。
  - **逐次実行・失敗時停止**: スクリプトはファイル名昇順で実行し、途中エラーで後続を実行せずエラー終了する。テストの再現性・順序性が重視される（確度: high / 根拠: 事実: src/plugins/process/scripts/README.adoc:212-213）。
  - **秘匿情報の保護**: パスワードの暗号化保管（openssl）とログのシークレットマスキングを備え、テスト環境の資格情報を扱う前提がある（確度: high / 根拠: 事実: src/bin/lib/commons/bash_utils:260-299, src/bin/lib/setenv:146-151（log.mask））。
  - **マルチホスト・環境別実行**: inventory による環境（staging 等）・ホストグループの切り替えを想定（確度: high / 根拠: 事実: src/template/stfw.yml:7（inventory: staging.yml）, src/template/config/inventory/staging.yml）。
  - **タイムゾーン依存**: Asia/Tokyo をデフォルトとし、日本の業務システムを主対象とする（確度: medium / 根拠: 事実: src/config/stfw.yml:16, dockerfiles/kcov/Dockerfile:4-5。ただし「日本市場向け」という明文は無く設定値からの推測）。
  - **リアルタイム性・規制要件**: 該当する証拠なし。テスト実行基盤であり、公共性・法規制の直接の制約は読み取れない（確度: low / 根拠: 推測: コード・ドキュメント・コミット履歴のいずれにも規制・SLA への言及が無いため「特段の規制要件なし」とドメイン一般論で判断）。

## 後段フェーズへの申し送り

- FIXME: docs/adoc 配下（overview/concept/manual）が placeholder のままで、意図・背景（なぜ）の一次情報がほぼ存在しない。要求の reason は Phase 以降もコミット履歴（git log）を主な証拠源とする必要がある。
- 外部システム候補: digdag（内包だが独立プロセス）、webhook 受信先、テスト対象ホスト群（inventory）。Phase2 以降でシステムコンテキストとして整理する（事実: docs/design/architecture/app_arch.puml:7-8,80-81）。
- 状態モデル候補: プロセス実行ステータス（Pending/Started/Success/Error/Blocked）は setenv の定数として一元定義されており、状態モデル逆生成の直接証拠になる（事実: src/bin/lib/setenv:27-31）。
