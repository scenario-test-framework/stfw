# clearMysql プラグイン

MySQL のテーブルの全行を DELETE する Arrange フェーズ (初期化) の組込みプラグイン。

## 前提

- 必要コマンド: `mysql` (実行前に PATH ゲートされる)
- 接続情報は inventory / secret から解決する (config への直書きは禁止)。
  詳細は exportMysql の README を参照 (同じ接続モデル)

## 設定 (config/config.yml)

| キー | 必須 | 説明 |
|---|---|---|
| `host_group` | ○ | 接続先 DB ホストを解決する inventory グループ名 |
| `database` | ○ | 対象データベース名 |
| `user` | ○ | 接続ユーザー |
| `tables[]` | ○ | 削除対象テーブル名のリスト (**FK の子 → 親順**に列挙) |
| `port` | - | MySQL ポート (既定 3306) |

> ⚠️ `tables` が 0 件の場合は何もせず成功する (no-op)。設定漏れに注意。

```yaml
stfw:
  process:
    clearMysql:
      host_group: db
      database: appdb
      user: appuser
      tables:
        - orders
```

## 動作

`tables` の全テーブルを `DELETE FROM` で空にする。全削除は `START TRANSACTION`〜`COMMIT` の
1 トランザクションで実行され、途中で失敗した場合は何も削除されない (all-or-nothing)。
データ投入 (importMysql) の先行プロセスとして使う。

- **FK 制約は無効化しない** (`FOREIGN_KEY_CHECKS` に触れない)。行単位で FK 検証されるため、
  `tables` は **FK の子 → 親の順**に列挙する
- all-or-nothing の保証は対象テーブルが **InnoDB 等のトランザクション対応エンジン**である
  ことが前提 (MyISAM はトランザクション非対応のため各 DELETE が個別に確定する)
- 参照行が残る親を削除しようとすると MySQL の FK 違反エラー (1451) が stderr に出て失敗する
- `ON DELETE CASCADE` が宣言されているテーブルへは DDL どおり波及する
- AUTO_INCREMENT はリセットしない。完全リセットが必要ならカスタムプラグインで行う

## 終了コード

| コード | 意味 |
|---|---|
| 0 | 全テーブルの削除成功 |
| 6 | 削除失敗 (トランザクションごと巻き戻り、何も削除されない) |
