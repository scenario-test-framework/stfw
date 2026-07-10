# シナリオ: daily-balance

日次残高バッチのシナリオテスト (プラグインエコシステムの実プロジェクト例)。
業務日付をまたいで口座残高の繰越を検証する。
Arrange (clear/import) → Act (invokeRest) → Collect (exportPostgres) → Assert (compare)。

## 要求トレーサビリティ

| 要求仕様 | 検証する process |
|---|---|
| REQ-01 当日取引が口座残高へ正しく反映される | _10_20240101/_50_assert_compare, _20_20240102/_30_assert_compare |
| REQ-02 前業務日の残高が翌業務日へ繰り越される | _20_20240102/_30_assert_compare |

## _10_20240101 — Day1 (2024-01-01)。初期残高を投入し、当日の取引を反映して残高を検証する。

| # | process | グループ | プラグイン | 説明 |
|---|---|---|---|---|
| _10 | _10_arrange_clearPostgres | arrange | clearPostgres | users / accounts / transactions を truncate して初期状態にする (reset)。 |
| _15 | _15_arrange_importMasterData | arrange | importMasterData | シナリオ共通の口座名義マスタ (users) を投入する (seed)。 |
| _20 | _20_arrange_importPostgres | arrange | importPostgres | 初期残高 CSV (acc-001=1000 / acc-002=2000) を accounts へ投入する (seed)。 |
| _30 | _30_act_invokeRest | act | invokeRest | 取引を API へ POST する (acc-001 +500 / acc-002 +300)。 |
| _40 | _40_collect_exportPostgres | collect | exportPostgres | 取引反映後の残高を evidence/appdb/accounts.csv へ収集する。 |
| _50 | _50_assert_compare | assert | compare | 当日残高が期待値 (acc-001=1500 / acc-002=2300) と一致することを検証する。 |

### _10_arrange_clearPostgres

- グループ: arrange
- 要求仕様: -
- 設定:

    ```yaml
    tables:
        - transactions
        - accounts
        - users
    ```

### _15_arrange_importMasterData

- グループ: arrange
- 要求仕様: -
- 設定:

    ```yaml
    tables:
        - users
    ```

### _20_arrange_importPostgres

- グループ: arrange
- 要求仕様: -
- 設定:

    ```yaml
    tables:
        - accounts
    ```

### _30_act_invokeRest

- グループ: act
- 要求仕様: -
- 設定:

    ```yaml
    host_group: api
    script: script.js
    ```

### _40_collect_exportPostgres

- グループ: collect
- 要求仕様: -
- 設定:

    ```yaml
    tables:
        - accounts
    ```

### _50_assert_compare

- グループ: assert
- 要求仕様: REQ-01 当日取引が口座残高へ正しく反映される
- 設定:

    ```yaml
    compare_files_version: v2.2.0
    ```

## _20_20240102 — Day2 (2024-01-02)。Day1 の残高を「繰り越して」当日の取引を反映する。

| # | process | グループ | プラグイン | 説明 |
|---|---|---|---|---|
| _10 | _10_act_invokeRest | act | invokeRest | 前業務日の残高に対して取引を POST する (acc-001 -200 / acc-002 +100)。 |
| _20 | _20_collect_exportPostgres | collect | exportPostgres | 取引反映後の累積残高を evidence/appdb/accounts.csv へ収集する。 |
| _30 | _30_assert_compare | assert | compare | 繰越後の累積残高が期待値 (acc-001=1300 / acc-002=2400) と一致することを検証する。 |

### _10_act_invokeRest

- グループ: act
- 要求仕様: -
- 設定:

    ```yaml
    host_group: api
    script: script.js
    ```

### _20_collect_exportPostgres

- グループ: collect
- 要求仕様: -
- 設定:

    ```yaml
    tables:
        - accounts
    ```

### _30_assert_compare

- グループ: assert
- 要求仕様: REQ-01 当日取引が口座残高へ正しく反映される, REQ-02 前業務日の残高が翌業務日へ繰り越される
- 設定:

    ```yaml
    compare_files_version: v2.2.0
    ```
