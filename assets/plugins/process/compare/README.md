# compare プラグイン

期待値 (`expect/`) と収集エビデンス (`actual/`) をディレクトリ突合する Assert フェーズの組込みプラグイン。比較は外部 OSS [compare-files](https://github.com/scenario-test-framework/compare-files) で行う。

ファイルの組み合わせは相対パスの完全一致で決まる。名前に出力時刻や連番を含み、実行ごとに名前が変わるファイルは
[compareRegex](../compareRegex/README.md) (比較対象を正規表現で指定する) を使う。

## 前提

- プロビジョニング: `stfw plugin install compare` (通常は `stfw init` が自動実行) が
  実行ホストの os_arch 版 `compare_files` をダウンロードし
  `.stfw/cache/plugins/compare/` へキャッシュする (install 時に `curl` / `tar` が必要)。
  `stfw:full` イメージは既定の版の compare_files を同梱しており、キャッシュが無ければ同梱版を
  使うため install 不要 (`compare_files_version` を上書きした場合は install が必要)
- 接続情報は不要 (ローカルのファイル比較)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `compare_files_version` | - | 取得する compare-files のリリースタグ (既定 v2.3.0)。**プロジェクト共通 `config/plugins/process/compare/config.yml` で指定する** (プロセスの `config/config.yml` に書いても使わない) |
| `on_mismatch` | - | 比較不一致の扱い。`error` (既定) = ステップ失敗として停止 + Blocked 伝播 (回帰テスト運用) / `warn` = Warn として記録して続行 (差分確認運用。run 終了コードは Warn あり・Error なしで 3) |

比較レイアウト (`compare_layout/*.json`) は次の後勝ちマージで解決される (AS-BUILT §4.11):
バイナリ同梱デフォルト → **プロジェクト共通** `{proj}/config/plugins/process/compare/compare_layout/`
(COMPAREFILES_CLASSPATH で注入) → **プロセスローカル** `config/compare_layout/`。
シナリオ横断の共通レイアウトはプロジェクト共通に置き、プロセス固有の上書きはプロセスローカルに置く。
compareRegex はこのディレクトリを読まない (compareRegex の共通ルールは `config/plugins/process/compareRegex/` に置く)。
起動設定 (`compare_files.{json,yaml,yml}`) はプロセスローカル `config/` にあれば `-config` で明示指定して最優先に使う。
プロジェクト共通に置いた起動設定は、compare-files の探索 (`COMPAREFILES_CLASSPATH` の各要素 → カレント → `./config`) で
最初に見つかった 1 件になる場合 (実行環境の既存の `COMPAREFILES_CLASSPATH` に起動設定が無い場合) に「プロジェクト既定」として効く:
プロセスローカルが無ければそのまま使われ、あればプロセスローカルで欠けているキーの補完元になる
(ローカルに書いた `false` と空配列 `[]` はそのまま使われ、共通側の値を無効化・解除できる。compare-files v2.3.0〜。v2.2.4 までは補完元の値を引き継いでいた。AS-BUILT §4.11)。
compare-files 同梱の起動設定を複製したプロセスローカルは、書かれた `sorted` / `writeDiffOnly` の `false` が共通側の `true` より優先されるので、共通側の値を使いたい項目は複製から消す。`sorted` などの設定項目のキーを大文字・小文字だけ変えて併記した起動設定 (`Sorted` と `sorted` など) は v2.3.0 から読込エラー (exit 6) になるので、一方にそろえる。真偽の項目に書いた空文字 `""` も `false` として効く (共通側の値を引き継ぐには、キーを書かないか `null` にする)。

## ディレクトリ規約 (エビデンスディレクトリ規約)

```
_{seq}_{group}_compare/
├── expect/                          # git 管理 (テスト作者が用意)
│   └── {収集系 process ディレクトリ名}/   # 同一 bizdate 内の収集系 process 名
│       └── ...                      # 当該 process の evidence/ 配下と同型
├── actual/                          # gitignore・自動生成 (evidence への symlink 群)
└── result/                          # gitignore (CompareSummary.csv 等の比較結果)
```

例: `expect/_20_web_collectLog/web01/var/log/app/app.log` は
同一 bizdate の `_20_web_collectLog/evidence/web01/var/log/app/app.log` と比較される。

## 動作

1. `actual/` と `result/` を毎回削除して再構築する (古い結果を残さない)
2. `expect/` 直下の各 process 名に対応する evidence ツリーを、実ディレクトリ +
   ファイル単位 symlink で `actual/` に再現する
3. `compare_files -od result expect actual` を実行する

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全ファイル一致 (OK / Ignore) |
| 3 | 差分あり (NG / LeftOnly / RightOnly) かつ `on_mismatch: warn` — Warn として記録され後続は継続する |
| 6 | 差分あり かつ `on_mismatch: error` (既定。ステップ失敗として後続を Blocked にする)、または比較エラー |

## 既知の制約

- evidence 配下のファイル名に改行を含むケースは非対応
- `compare_files_version` を変更したときは `.stfw/cache/plugins/compare/` を削除してから再 install する
