# v0.2 (Bash + digdag) から v1.0 (Go) への移行ガイド

v1.0 は全面再実装です。次の 3 つの互換境界は維持されているため、**シナリオ資産（ディレクトリ・スクリプト）はそのまま動きます**:

1. ディレクトリ規約: `scenario/{name}/_{seq}_{bizdate}/_{seq}_{group}_{type}/`
2. プロセスプラグイン実行契約: 環境変数 + リターンコード (0/3/6)
3. webhook payload JSON スキーマ（下記の非互換 2 点を除く）

## コマンド対応表

| v0.2 | v1.0 |
|---|---|
| `install` スクリプト + tar.gz 展開 | バイナリ配布 (GitHub Releases / Docker) |
| `stfw init` | `stfw init`（同じ） |
| `stfw scenario -i <name>` | `stfw new scenario <name>` |
| `stfw bizdate -i <seq> <bizdate>` | `stfw new bizdate <seq> <bizdate>` |
| `stfw process -i <seq> <group> <type>` | `stfw new process <seq> <group> <type>` |
| `stfw scenario -g/-G`（dig 生成 + 検証） | `stfw validate [scenario...]`（dig 生成は廃止） |
| `stfw server start` → `stfw run -f <scenario>` | `stfw run <scenario...>`（server 不要） |
| `stfw run -d/--dry-run` | `stfw run --dry-run`（意味を整理 — 下記） |
| digdag Web UI / ログ追従 | `stfw status [run_id]` + HTML レポート（`stfw report`） |
| `stfw gen-encrypt-key` | `stfw secret keygen` |
| `stfw passwd <host> <user>` | `stfw secret set <host> <user>` |
| `stfw passwd -s <host> <user>` | `stfw secret show <host> <user>` |
| （なし） | `stfw secret migrate`（旧形式からの一括変換） |
| `stfw inventory --list [group]` | `stfw inventory list [group]`（出力互換） |
| `stfw inventory --is-exist <group>` | `stfw inventory exists <group>`（出力互換） |
| `stfw process -l` / `-I <type>` | `stfw plugin list` / `stfw plugin install <type>` |
| （未配線の `gen_ssh_server_key`） | `stfw ssh trust <host\|group>` |
| `stfw server *` / `stfw digdag` | **廃止**（実行エンジン内包化により不要） |

## 移行手順

1. v1.0 バイナリを導入する（旧 `install` 資産・digdag jar・JVM は不要）
2. プロジェクトの `stfw.yml` から `stfw.server.*` セクションを削除する（残っていても無害。読み込み時に警告が出ます）
3. 資格情報を移行する: `stfw secret migrate`
   - 旧 openssl S/MIME 形式を旧 RSA 秘密鍵で復号し、age 形式で再暗号化します
   - 事前に `stfw secret keygen` で age キーペアを生成してください
   - 旧ファイルは `.bak` として退避されます
4. `stfw validate` を実行する。残存 `*.dig` ファイルは「v1.0 では不要」と警告されるので削除してよい
5. webhook 受信側がある場合は下記「非互換事項」を確認する

## 非互換事項

### webhook

- **webhook_id の形式が変わりました**: 旧実装には id 導出に余分な `}` が混入するバグがあり、v1.0 は正しい形式（`{run_id}+run+...`）に統一しています。受信側が旧（壊れた）id 形式に依存している場合は修正が必要です
- **通知抑制設定が機能するようになりました**: 旧実装はバグ（文字列の `-eq` 数値比較）により `on_success` / `on_error` = false でも end 通知が常に送信されていました。v1.0 では設定どおり抑制されます。`on_*` は `true` のみ有効（未設定は送信しない）
- payload の `stfw.home`・`digdag.url`・`digdag.version` はキーを維持したまま空文字になります（STFW_HOME / digdag の廃止のため）
- 独自プラグインの webhook 詳細契約（`bin/webhook/get_{start,end}_content`）は廃止しました。step 詳細の投影は組込み `scripts` タイプのみです（実行ジャーナルが唯一のソース）

### 実行

- **dry-run の意味を整理しました**: 旧 dry-run は「dig 生成 + setup/teardown 実行」の 2 役でした。v1.0 では静的検証を `stfw validate` に分離し、`run --dry-run` は「execute / post_execute をスキップして実行」の意味だけを持ちます
- `run -s/-t/-f` オプションと attempt_id は廃止（digdag 固有のため）。run_id 採番規則（`_{yyyymmddhhmmss}_{pid}`）は互換です
- 実行は逐次のみです（旧実装も実質逐次。`max_task_threads` 設定は廃止）
- run 開始前に validate 相当の静的検証が自動実行されます

### 環境変数（プラグイン env 契約）

- `STFW_HOME` は廃止されました（単一バイナリ化のため。配布ディレクトリという概念がなくなりました）
- `stfw_*`（設定のフラット化）・`STFW_PROJ_DIR` 系・実行コンテキスト（`run_id` / `stfw_scenario_*` / `stfw_bizdate_*` / `stfw_process_*` 等）は 1:1 で維持されています

### その他

- **旧 Bash 版のプロジェクトプラグイン**（`${STFW_HOME}/bin/lib/setenv` を source するもの）は、そのままでは動きません。setenv 依存を除去して自己完結なスクリプトにしてください（env 契約の変数はすべて注入済みのため、通常は source 行の削除だけで済みます）
- bizdate の検証が強化されました: 旧は「8 桁数字」のみでしたが、v1.0 は実在日付もチェックします（例: `20260231` はエラー。`99990101` は実在日付なので従来どおり使えます）
- 資格情報の保存形式が age (X25519) になりました。ログのシークレットマスキング（`PASSWORD` / `TOKEN` と復号値の `[secret]` 置換）は維持されています
