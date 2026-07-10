-- daily-balance-sut のスキーマ。postgres コンテナの初期化時に実行される。
-- accounts が口座マスタ (残高)、transactions が取引履歴。
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
