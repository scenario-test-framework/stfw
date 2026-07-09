# sshExec プラグイン

自プロセスの `scripts/` 配下のスクリプトを、リモートホスト群へ ssh 経由で送ってファイル名昇順に一括実行する Act フェーズの組込みプラグイン。テスト対象システムのバウンダリスクリプト実行などに使う。

## 前提

- 必要コマンド: `ssh` / `sshpass` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止):
  - ホスト: inventory グループ (`host_group`)
  - パスワード: secret (`{host}-{user}`)。`SSHPASS` 環境変数で渡すため argv に露出しない
  - SSH ホストキー: `stfw ssh trust {host|group}` で known_hosts へ事前登録

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | inventory グループ名 (ホスト解決) |
| `user` | ○ | ログインユーザー (secret `{host}-{user}` でパスワード解決) |

```yaml
stfw:
  process:
    sshExec:
      host_group: web
      user: appuser
```

## 配置物 (テスト作者が用意・git 管理)

```
_{seq}_{group}_sshExec/
└── scripts/          # リモートで実行するスクリプト (ファイル名昇順)
    ├── 10_setup.sh
    └── 20_kick_batch.sh
```

## 動作

解決した各ホストに対し、`scripts/` 直下の各ファイルを昇順に
`ssh {user}@{host} 'bash -s' < script` (標準入力パイプ送信) でリモート実行する。
スクリプトはリモートホストに保存されない。

**fail-fast**: スクリプトが 1 つでも失敗したら即時終了し、以降のスクリプト・ホストは実行しない。

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全ホスト・全スクリプトの実行成功 |
| 6 | いずれかのスクリプトが失敗 (即時終了)、または設定不備・scripts/ 不在 |

## 既知の制約

- スクリプトは `bash -s` の標準入力として送るため、リモート側で bash が必要
- スクリプト自体が標準入力を読む処理には向かない (スクリプト本文の伝送に使用するため)
