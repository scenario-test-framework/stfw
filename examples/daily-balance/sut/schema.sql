-- daily-balance-sut のスキーマ。postgres コンテナの初期化時に実行される。
-- users が口座名義のマスタ (参照データ)、accounts が口座の残高、transactions が取引履歴。

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

CREATE TABLE IF NOT EXISTS transactions (
    id         BIGSERIAL PRIMARY KEY,
    account_id TEXT      NOT NULL,
    amount     BIGINT    NOT NULL,
    bizdate    TEXT      NOT NULL
);
