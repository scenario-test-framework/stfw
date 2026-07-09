# collectFile プラグイン

テスト対象ホストからファイル (外部 IF ファイル等) を収集する Collect フェーズの組込みプラグイン。

## 前提

- 必要コマンド: `ssh` / `scp` / `sshpass` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止):
  - ホスト: inventory グループ (`targets[].group`)
  - パスワード: `stfw secret set {host} {user}` で登録した secret (`{host}-{user}`)
  - SSH ホストキー: `stfw ssh trust {host|group}` で known_hosts へ事前登録

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `targets[].group` | ○ | inventory グループ名 (ホスト解決) |
| `targets[].user` | ○ | ログインユーザー (secret `{host}-{user}` でパスワード解決) |
| `targets[].paths[]` | ○ | 収集対象ファイルパスの正規表現 (posix-extended) |

> ⚠️ `targets` は先頭から連番で読み、**最初の `group` 未設定要素で打ち切る** (以降は黙って無視される)。実質 0 件なら何もせず成功する (no-op) ため設定漏れに注意。

```yaml
stfw:
  process:
    collectFile:
      targets:
        - group: web
          user: appuser
          paths:
            - /var/lib/app/out/.*\.csv
```

## 動作

1. グループ毎に inventory からホストを解決する
2. 各ホストで `find` (posix-extended regex) により対象ファイルを列挙する
3. `scp` でローカルへ収集する

## 出力

```
{process}/evidence/{host}/{収集元の絶対パスをそのまま再現}
```

エビデンスディレクトリ規約に従う (gitignore 対象。compare プラグインの期待値 `expect/` と同型)。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全 target の収集成功 |
| 6 | いずれかの収集に失敗 (ステップ失敗。残りの target は継続試行する) |
