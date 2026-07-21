# parallel プラグイン

プロセスディレクトリ配下に定義した**子プロセス**を並走させる組込みプロセスタイプ。
子プロセス 1 件は実行ジャーナル上の 1 ステップとして記録され、
親のステータスは全子の悪い方 (Error > Warn > Success) になる。

Go ネイティブ実装のため外部コマンドへの依存はない。

## 使い方

```
_{seq}_{group}_parallel/
├── config/config.yml            # parallel 自体の設定 (max_parallel)
├── _10_db_exportPostgres/       # 子プロセス (親プロセスと同形式 _{seq}_{group}_{type})
│   ├── config/config.yml
│   └── evidence/                # 子ごとに分離される
└── _20_db_exportMysql/
    └── config/config.yml
```

- 子の走査規則は通常プロセスと同じ (`_` 始まりのみ・名前昇順)。seq は実行キューへの投入順
  (max_parallel 待機時の開始順) と表示順にのみ使う。同時起動される子同士の実開始順は不定
- 子のタイプに `parallel` は指定できない (入れ子禁止)
- 子は必ず全件実行される (先行子が Error でも他の子は止まらない。子同士に Blocked は無い)
- `stfw new process` は子の scaffold に対応しない (手動作成または `stfw scenario scaffold` で生成)

## 設定 (config/config.yml)

```yaml
stfw:
  process:
    parallel:
      max_parallel: 2   # 同時実行数の上限。0 = 上限なし
```

プロジェクト全体の既定は stfw.yml の `stfw.process.parallel.max_parallel` (同梱既定 0)。
プロセスの config/config.yml で個別に上書きできる。

## 動作

1. 子ディレクトリ名を昇順に列挙する (計画列挙。dry-run でも行う)
2. process 階層フック (setup) を親で 1 回実行する (子では実行しない)
3. 各子を goroutine で並走実行する (max_parallel 超過分は seq 順に待機)。
   子の実行内容は通常プロセスと同じ (config チェーン注入 + exec 契約 / scripts ネイティブ)
4. 子の完了ごとにステップとして記録する (Success / Warn / Error + exit code 0/3/6)
5. 全子の悪い方を親のステータスにする。親が Error なら後続の兄弟プロセスは実行されない
6. process 階層フック (teardown) を親で 1 回実行する

## 補足

- 実装は stfw 本体の Go ネイティブ (`internal/usecase/runscenario/parallel.go`)
- 子の stdout/stderr は並走のため行単位で交錯し得る (行内は壊れない)
- プロジェクト側 `plugins/process/parallel/` に同名プラグインを置いた場合はそちらが優先され、
  通常の exec 契約 (bin/run/{pre_execute,execute,post_execute}) で実行される
