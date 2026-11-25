# collectLog プラグイン

テスト対象ホストから「業務日付の実行開始日時以降」のログ行だけを収集する Collect フェーズの組込みプラグイン。時刻フィルタには外部 OSS [logfilter](https://github.com/scenario-test-framework/logfilter) を使う。

## 前提

- 必要コマンド: `ssh` / `scp` / `sshpass` (実行前に PATH ゲートされる)。install 時は `curl` / `tar`
- プロビジョニング: `stfw plugin install collectLog` (通常は `stfw init` が自動実行) が
  logfilter バイナリを arch 別にダウンロードし `.stfw/cache/plugins/collectLog/` へキャッシュする。
  `stfw:full` イメージは既定の版・arches の logfilter を同梱しており、キャッシュに無い arch は
  同梱版で補うため install 不要 (`logfilter_version` を上書きした場合や同梱に無い arch は install が必要)
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
| `targets[].log_format` | - | ログの形式。`text` (既定) / `json` (JSON Lines) / `logfmt`。構造化ログはログ内の時刻フィールドで絞り込む |
| `targets[].timestamp_key` | - | 構造化ログの時刻フィールドのキー (JSON Lines はドット区切りで入れ子をたどれる。例: `log.ts`。logfmt ではドットもキー名の一部)。省略時は `time` / `timestamp` / `ts` / `@timestamp` を順に探す |
| `targets[].timestamp_format` | - | 構造化ログの時刻の形式。`auto` (既定) / `unix` / `unixms` / `unixus` / `unixns` |
| `logfilter_version` | - | 取得する logfilter のリリースタグ (既定 v2.1.0)。**プロジェクト共通 `config/plugins/process/collectLog/config.yml` で指定する** (プロセスの `config/config.yml` に書いても使わない。`logfilter_arches` も同じ) |
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

### 構造化ログ (JSON Lines / logfmt)

`targets[]` のエントリごとに `log_format` などを書くと、そのエントリの `paths` を構造化ログとして読み、
ログ内の時刻フィールドで基準時刻以降のイベントに絞り込む (logfilter v2.1.0 以降の機能)。書かないエントリは従来どおりテキスト形式。

```yaml
stfw:
  process:
    collectLog:
      targets:
        - group: web            # テキストのログ (従来どおり)
          user: appuser
          paths:
            - /var/log/app/.*\.log
        - group: web            # {"log":{"ts":1704153600000,...}} のような JSON Lines
          user: appuser
          log_format: json
          timestamp_key: log.ts
          timestamp_format: unixms
          paths:
            - /var/log/app/.*\.jsonl
```

- 値が不正なら、どのホストにも接続する前にエラー (exit 6)。`timestamp_key` / `timestamp_format` は `json` / `logfmt` のときだけ書ける
- 時刻の読み方: `T` 区切りの ISO 8601 形式はオフセットを読む。空白区切りの日時は収集先ホストのローカル時刻 (オフセット付きの表記は確実には読めない: `+00:00` はローカル時刻扱い、`-05:00` は変換エラー。オフセットを使うなら `T` 区切りにする)。Unix 時刻 (数値・`unix*`) は絶対時刻
- logfilter v2.1.0 以降が必要。プロジェクト共通の `logfilter_version` で旧版を指定しているなら、まず v2.1.0 以降にする (または外す)。v2.1.0 より前を install 済みなら、版を上書きしていなくても `.stfw/cache/plugins/collectLog/logfilter/` を消してから `stfw plugin install collectLog` をやり直す (キャッシュは版を照合せず優先される)

## 動作

1. フィルタ基準時刻 = `stfw_bizdate_start_ts` (bizdate 階層の実行開始日時)
2. ホストのログインシェルが非対話の接続で標準出力に何も出さないことを確かめる (出すホストはエラー)。ホストに `mktemp -d` で専用ディレクトリ (`/tmp/stfw.XXXXXXXX`) を作り、arch 版 logfilter を scp 転送する
3. ssh で logfilter を実行し、基準時刻以降の行 (構造化ログはイベント) に絞り込む (出力も専用ディレクトリに置く)。該当が 0 件 (logfilter は exit 3) なら空のファイルになり、それをそのまま収集する (「その日はログが出ない」を空の期待値で確かめられる)
4. scp でローカルへ収集し、最後に専用ディレクトリごと削除する (削除できなければ、残った場所を示してエラーにする)

パスワードは環境変数 `SSHPASS` で `sshpass` に渡し、コマンドの引数には載せない。相手のホストで実行するコマンドに埋め込む値 (パターン・ファイル名・基準時刻) は、すべて単一引用符で囲んで渡す。そのため `'` や `;` を含むログのファイル名も、その名前のまま収集し、名前の一部がコマンドとして実行されることはない。

## 出力

```
{process}/evidence/{host}/{収集元の絶対パスをそのまま再現}
```

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全 target の収集成功 (基準時刻以降のログが 0 件のファイルは、空のファイルとして収集して成功扱い) |
| 6 | いずれかの収集に失敗 (残りの target は継続試行する) |

## 補足

- `logfilter_version` を変更したときは `.stfw/cache/plugins/collectLog/` を削除してから
  再 install する (キャッシュは「初回 install 優先」)
