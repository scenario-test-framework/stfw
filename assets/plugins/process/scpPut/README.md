# scpPut プラグイン

ローカルの `target/{グループ}/` 配下のディレクトリ構成を、リモートホスト群の配置先へ scp で**原子的に**配置する Arrange フェーズの組込みプラグイン。外部 IF ファイルの配置などに使う。

## 前提

- 必要コマンド: `ssh` / `scp` / `sshpass` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止):
  - ホスト: inventory グループ (`targets[].group`)
  - パスワード: secret (`{host}-{user}`)。`SSHPASS` 環境変数で渡すため argv に露出しない
  - SSH ホストキー: `stfw ssh trust {host|group}` で known_hosts へ事前登録

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `targets[].group` | ○ | inventory グループ名 (ホスト解決) |
| `targets[].user` | ○ | ログインユーザー (secret `{host}-{user}` でパスワード解決) |
| `targets[].dest` | ○ | リモートの配置先ディレクトリ (グループ毎に指定。安全文字 `A-Z a-z 0-9 . _ - /` のみ) |

> ⚠️ `targets` は先頭から連番で読み、**最初の `group` 未設定要素で打ち切る** (以降は黙って無視される)。実質 0 件なら何もせず成功する (no-op) ため設定漏れに注意。

```yaml
stfw:
  process:
    scpPut:
      targets:
        - group: web
          user: appuser
          dest: /opt/app/conf
```

## 配置物 (テスト作者が用意・git 管理)

```
_{seq}_{group}_scpPut/
└── target/
    └── {group}/          # このディレクトリ構成がそのまま dest へ配置される
        ├── app.conf      # 隠しファイル・空ディレクトリも保持される
        └── sub/nested.conf
```

## 動作 (原子的配置)

1. `target/{group}/` 配下を、`dest` の**隣接一時ディレクトリ** (同一ファイルシステム) へ scp put する
2. 一時ディレクトリを `dest` へディレクトリ丸ごと rename して入れ替える
   (既存 `dest` は退避 → 入れ替え → 退避を削除、の 2 段階 rename)
3. 入れ替えに失敗した場合は退避から `dest` を復元し、配置前の状態へロールバックする

`dest` に転送途中の不完全なファイルや新旧混在の状態が現れることはない
(転送中は一時ディレクトリ側に書かれ、`dest` へはディレクトリ丸ごとの rename でのみ反映される)。

- 既存 `dest` の置換時: 可観測状態は「配置前の完全な状態」「配置後の完全な状態」
  「2 回の rename の間の一時的不在」の 3 通り
- 初回配置時 (`dest` 未存在): 最終 rename の完了までは不在のままで、完了後に
  「配置後の完全な状態」が現れる

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全 target・全ホストへの配置成功 |
| 6 | いずれかの配置に失敗 (設定不備・dest 不正文字を含む。残りの target は継続試行する) |

## 既知の制約

- **全置換セマンティクス**: `dest` は source 内容で丸ごと置き換わる (source に無い既存ファイルは
  残らない)。`dest` には本プラグイン専用の配置先ディレクトリを指定すること
- `dest` にシェルメタ文字 (空白・引用符・`$` 等) は使えない (注入防止のため安全文字検証で拒否)
