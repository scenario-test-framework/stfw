# collectLog プラグイン

テスト対象ホストから「業務日付の実行開始日時以降」のログ行だけを収集する Collect フェーズの組込みプラグイン。時刻フィルタには外部 OSS [logfilter](https://github.com/scenario-test-framework/logfilter) を使う。

## 前提

- 必要コマンド: `ssh` / `scp` / `sshpass` (実行前に PATH ゲートされる)。install 時は `curl` / `tar`
- プロビジョニング: `stfw plugin install collectLog` (通常は `stfw init` が自動実行) が
  logfilter バイナリを arch 別にダウンロードし `.stfw/cache/plugins/collectLog/` へキャッシュする
- 接続情報は inventory / secret から解決する (config への直書きは禁止):
  - ホスト: inventory グループ (`targets[].group`)
  - **arch**: inventory の構造化ホストエントリ `arch` (logfilter の送り分けに必須)
  - パスワード: secret (`{host}-{user}`)
  - SSH ホストキー: `stfw ssh trust` で事前登録
- 収集先ホストと stfw 実行ホストの時刻が同期しており、**タイムゾーン解釈も一致**していること
  (フィルタ基準時刻は stfw 実行ホストのローカル時刻の「日時」部分だけを logfilter へ渡すため、
  ログの時刻表記が別タイムゾーンだと絞り込みがずれる)

```yaml
# config/inventory/staging.yml (arch 指定つきホストエントリ)
stfw_inventory:
  - web:
    - host: web01
      arch: linux_amd64
```

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `targets[].group` | ○ | inventory グループ名 (ホスト解決) |
| `targets[].user` | ○ | ログインユーザー (secret でパスワード解決) |
| `targets[].paths[]` | ○ | 収集対象ログのパス正規表現 (posix-extended) |
| `logfilter_version` | - | 取得する logfilter のリリースタグ (既定 v2.0.0) |
| `logfilter_arches` | - | install 時にダウンロードする `{os}_{arch}` (スペース区切り) |

> ⚠️ `targets` は先頭から連番で読み、**最初の `group` 未設定要素で打ち切る** (以降は黙って無視される)。実質 0 件なら何もせず成功する (no-op) ため設定漏れに注意。

```yaml
stfw:
  process:
    collectLog:
      targets:
        - group: web
          user: appuser
          paths:
            - /var/log/app/.*\.log
```

## 動作

1. フィルタ基準時刻 = `stfw_bizdate_start_ts` (bizdate 階層の実行開始日時)
2. ホストの arch 版 logfilter を scp 転送する
3. ssh で logfilter を実行し、基準時刻以降の行に絞り込む
4. scp でローカルへ収集し、転送したバイナリを削除する

## 出力

```
{process}/evidence/{host}/{収集元の絶対パスをそのまま再現}
```

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全 target の収集成功 |
| 6 | いずれかの収集に失敗 (残りの target は継続試行する) |

## 補足

- `logfilter_version` を変更したときは `.stfw/cache/plugins/collectLog/` を削除してから
  再 install する (キャッシュは「初回 install 優先」)
