# scripts プラグイン

`scripts/` ディレクトリ配下のスクリプト群をファイル名昇順に一括実行する、最も基本的な組込みプロセスタイプ。任意のフェーズ (Arrange / Act / Collect / Assert) に使える。

Go ネイティブ実装のため外部コマンドへの依存はない (このディレクトリの `bin/install` は前提コマンド確認のみ)。

## 使い方

```console
$ stfw new process 10 web scripts    # プロセスの scaffold を生成
```

```
_{seq}_{group}_scripts/
├── config/config.yml   # 任意の設定 (フラット化して環境変数で公開)
└── scripts/            # ステップスクリプト (昇順に逐次実行)
    ├── 100_1st_step
    └── 200_2nd_step
```

## 設定 (config/config.yml)

任意のキーを定義できる。`stfw.process.scripts.*` 配下がフラット化され、
`stfw_process_scripts_{キー}` としてステップスクリプトの環境変数に公開される。

```yaml
stfw:
  process:
    scripts:
      some_key: value   # -> 環境変数 stfw_process_scripts_some_key
```

## 動作

1. `scripts/` 直下のファイルを昇順に列挙する (計画列挙。dry-run でも行う)
2. 各スクリプトへ実行権限を付与する (pre_execute 相当)
3. 昇順に逐次実行する。作業ディレクトリは `scripts/`、環境変数は実行契約どおり注入される
4. あるステップがエラー終了すると、以降のステップは実行せず `Blocked` として記録する
5. `stfw run --dry-run` では列挙のみ行い、実行はスキップする

## 終了コード (ステップ)

| コード | ステップ状態 |
|---|---|
| 0 | Success |
| 非 0 | Error (以降のステップは Blocked) |

## 補足

- 実装は stfw 本体の Go ネイティブ (`internal/usecase/runscenario/process.go`)。
  このディレクトリの `template/` は `stfw new process` の scaffold 元。
- プロジェクト側 `plugins/process/scripts/` に同名プラグインを置いた場合はそちらが優先され、
  通常の exec 契約 (bin/run/{pre_execute,execute,post_execute}) で実行される。
