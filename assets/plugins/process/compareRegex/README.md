# compareRegex プラグイン

期待値 (`expect/`) と収集エビデンス (`actual/`) を、**比較対象設定 `config/compare_target.csv` の正規表現で対応付けて**突合する Assert フェーズの組込みプラグイン。比較は外部 OSS [compare-files](https://github.com/scenario-test-framework/compare-files) の `compare_regex` (正規表現指定比較) で行う。

名前に出力時刻や連番を含むファイル (`report_20240101093015.csv` など) のように、実行ごとに名前が変わり、
[compare](../compare/README.md) のディレクトリ比較 (相対パスの完全一致) では組み合わせられないファイルに使う。

| 比べたいもの | 使うプロセスタイプ |
|---|---|
| 名前が毎回同じファイル (エクスポート CSV など) | `compare` (ディレクトリ比較。期待値に無いファイルも差分として検知する) |
| 名前が実行ごとに変わるファイル | `compareRegex` (このプラグイン。CSV に書いたファイルだけを比較する) |

## 前提

- プロビジョニング: `stfw plugin install compareRegex` (通常は `stfw init` が自動実行) が
  実行ホストの os_arch 版 compare-files のリリース物から `compare_regex` を取り出し
  `.stfw/cache/plugins/compareRegex/` へキャッシュする (install 時に `curl` / `tar` が必要)。
  `stfw:full` イメージは compare と同じ版の `compare_regex` を同梱しており、キャッシュが無ければ同梱版を
  使うため install 不要 (`compare_files_version` を上書きした場合は install が必要)
- 接続情報は不要 (ローカルのファイル比較)

## 設定

| キー | 必須 | 置き場所 | 説明 |
|---|---|---|---|
| `compare_files_version` | - | プロジェクト共通 `config/plugins/process/compareRegex/config.yml` | 取得する compare-files のリリースタグ (既定 v2.2.4。compare の既定と同じ)。install はプロセス非依存の設定を見るため、プロセスの `config/config.yml` に書いても取得する版は変わらない |
| `on_mismatch` | - | プロジェクト共通、またはプロセスの `config/config.yml` | 比較不一致の扱い。`error` (既定) = ステップ失敗として停止 + Blocked 伝播 / `warn` = Warn として記録して続行 (compare と同じ) |

## 比較対象設定 (config/compare_target.csv・必須)

compare-files の比較対象設定ファイルの書式そのまま (CSV・UTF-8・`#` 始まりはコメント)。1 行に
「左ディレクトリ (期待値), 右ディレクトリ (実績値), ファイル名正規表現」を書く。左右のディレクトリは
**プロセスディレクトリからの相対パス**で `expect/{収集系 process 名}/…` と `actual/{収集系 process 名}/…` を指す。

```csv
# 左ディレクトリ,右ディレクトリ,ファイル名正規表現
# 出力時刻入りのファイル名
expect/_45_collect_scripts/api,actual/_45_collect_scripts/api,balance_report_\d{14}\.json
# 固定のファイル名も書ける
expect/_45_collect_scripts/api,actual/_45_collect_scripts/api,summary\.csv
```

- 左ディレクトリは `expect/{収集系 process 名}/…`、右ディレクトリは `actual/{収集系 process 名}/…` を指すこと (`.` `..` は使えない)。外れるとエラー (exit 6)。右に `expect/` を書いて期待値どうしを比べる取り違えや、process 名の書き忘れを防ぐ
- 「1 行 = 1 レコード」で書く: 行の終わりは LF か CRLF (CR 単独は不可)、クォートは行内で閉じる (コメント行でも、カンマ直後の `"` はクォートを開く)。外れるとエラー (exit 6)
- 1 行は 3 列。正規表現にカンマを含めるときはその列を `"report_\d{1,3}\.csv"` のようにクォートする (クォート内の `"` は `""` と書く)
- 正規表現はファイル名全体に一致させる (前後に `^` `$` は不要)。`\d` や先読みが使える (compare-files の正規表現エンジン dlclark/regexp2)
- 探すのは**指定ディレクトリの直下だけ** (下の階層は探さない)。ファイル名順で**最初に一致した 1 件**を比較する
- 左右のディレクトリが両方あり、片側だけ一致した行は差分 (LeftOnly / RightOnly)、どちらにも一致しない行は比較エラー。片側のディレクトリが無い行は差分、両方無い行は比較エラー
- **CSV に書かなかったファイルは比較しない**
- resume は前 run の evidence を引き継ぐ。時刻入りのファイルを出す収集ステップは、前回のファイルを消してから書く (残すと溜まり、名前順で最初の古いファイルが比較される)
- ファイルが無い・データ行が 1 行も無い (コメントと空行だけ) 場合はエラー (exit 6)。対象 0 件を一致として素通りさせない
- BOM 付き UTF-8 はエラー (exit 6)。Excel の「CSV UTF-8」で保存すると BOM が付くので、BOM なしで保存する

## 比較ルール (比較レイアウト・起動設定)

compare と同じ比較ルールを使う。プロジェクト共通は **compare のディレクトリ**
`{proj}/config/plugins/process/compare/` (`compare_layout/*.json` など) を共有し、プロセス固有の上書きは
プロセスローカル `config/compare_layout/` に置く (後勝ち)。起動設定 `config/compare_files.{json,yaml,yml}` は
プロセスローカルにあれば `-config` で最優先に使う。欠落キーは compare-files の既定の探索 (`COMPAREFILES_CLASSPATH` → カレント → `./config`) で最初に見つかった起動設定から補われる。プロジェクト共通などに起動設定が無いとプロセスローカル自身が既定になり補われない (`{}` だけでは必須キー不足で exit 6)。compare-files 同梱の `config/compare_files.json` を複製して編集するのが確実。

`{proj}/config/plugins/process/compareRegex/` は stfw の設定 (`config.yml`) の上書き専用で、
ここに `compare_layout/` や起動設定を置いても読まれない。

## ディレクトリ規約 (エビデンスディレクトリ規約)

```
_{seq}_{group}_compareRegex/
├── config/
│   ├── config.yml
│   └── compare_target.csv           # git 管理 (比較対象設定・必須)
├── expect/                          # git 管理 (テスト作者が用意)
│   └── {収集系 process ディレクトリ名}/   # 同一 bizdate 内の収集系 process 名
│       └── ...                      # 当該 process の evidence/ 配下と同型 (ファイル名は任意の 1 例)
├── actual/                          # gitignore・自動生成 (evidence への symlink 群)
└── result/                          # gitignore (CompareSummary.csv 等の比較結果)
```

## 動作

1. `actual/` と `result/` を削除して再構築する (古い結果を残さない。バイナリ未配置・`on_mismatch` 不正のときはその前に終了する)
2. `config/compare_target.csv` の存在・BOM の有無・各データ行の左右ディレクトリを確かめる
3. `expect/` 直下の各 process 名に対応する evidence ツリーを、実ディレクトリ +
   ファイル単位 symlink で `actual/` に再現する (compare と同じ)
4. `compare_regex [-config config/compare_files.json] -od result config/compare_target.csv` を実行する

## 終了コード

| コード | 意味 |
|---|---|
| 0 | CSV の全行が一致 (OK / Ignore) |
| 3 | 差分あり (NG / LeftOnly / RightOnly) かつ `on_mismatch: warn` — Warn として記録され後続は継続する |
| 6 | 差分あり かつ `on_mismatch: error` (既定)、または比較エラー・設定不備・比較結果が出なかった (環境変数 `COMPAREFILES_OPT` に `-v` / `-h` がある等) |

## 既知の制約

- evidence 配下のファイル名に改行を含むケースは非対応
- `compare_files_version` を変更したときは `.stfw/cache/plugins/compareRegex/` を削除してから再 install する
