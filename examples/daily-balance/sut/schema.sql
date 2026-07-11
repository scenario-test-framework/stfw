-- daily-balance-sut のスキーマ。postgres コンテナの初期化時に実行される。
-- biz_calendar が業務日付 (単一行)、users が口座名義のマスタ (参照データ)、
-- accounts が口座の残高、transactions が取引履歴。

-- biz_calendar: SUT の業務日付。id='system' の単一行運用。
-- カスタムプラグイン updateBizdate が bizdate 階層ごとに stfw_bizdate の値で更新し、
-- API は取引記録時にここから業務日付を解決する (payload では受け取らない)。
CREATE TABLE IF NOT EXISTS biz_calendar (
    id      TEXT PRIMARY KEY,
    bizdate TEXT NOT NULL
);

-- users: 口座名義のマスタデータ。カスタムプラグイン importMasterData が
-- scenario の config 内 CSV から投入する (別ファイルに切り出さない小さな参照データの例)。
CREATE TABLE IF NOT EXISTS users (
    id    TEXT PRIMARY KEY,
    name  TEXT NOT NULL,
    email TEXT
);

CREATE TABLE IF NOT EXISTS accounts (
    id      TEXT   PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0
);

-- transactions: 取引履歴。連番 id を持たせない: エビデンス突合 (exportPostgres → expect)
-- を決定的にするため (TRUNCATE はシーケンスをリセットせず、再実行で id がずれる)。
CREATE TABLE IF NOT EXISTS transactions (
    account_id TEXT   NOT NULL,
    amount     BIGINT NOT NULL,
    bizdate    TEXT   NOT NULL
);
