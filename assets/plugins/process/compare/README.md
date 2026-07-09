# compare プラグイン

期待値 (`expect/`) と収集エビデンス (`actual/`) をディレクトリ突合する Assert フェーズの組込みプラグイン。比較は外部 OSS [compare-files](https://github.com/scenario-test-framework/compare-files) で行う。

## 前提

- プロビジョニング: `stfw plugin install compare` (通常は `stfw init` が自動実行) が
  実行ホストの os_arch 版 `compare_files` をダウンロードし
  `.stfw/cache/plugins/compare/` へキャッシュする (install 時に `curl` / `tar` が必要)
- 接続情報は不要 (ローカルのファイル比較)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `compare_files_version` | - | 取得する compare-files のリリースタグ (既定 v2.2.0) |

compare-files の起動設定 (`compare_files.json`) と比較レイアウト (`compare_layout/`) は、
プロセスディレクトリの `config/` 配下に置くと自動探索される (無ければバイナリ同梱デフォルト)。

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
| 6 | 差分あり (NG / LeftOnly / RightOnly) またはエラー — 不一致はステップ失敗として後続を Blocked にする |

## 既知の制約

- evidence 配下のファイル名に改行を含むケースは非対応
- `compare_files_version` を変更したときは `.stfw/cache/plugins/compare/` を削除してから再 install する
