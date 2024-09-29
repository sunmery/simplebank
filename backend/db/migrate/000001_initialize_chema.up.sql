-- 账户表: 用户的基本信息, 一个账户只能属于一个用户
CREATE TABLE accounts
(
    id         bigserial PRIMARY KEY,
    owner      varchar                     NOT NULL, -- 用户,所有者
    balance    bigint                      NOT NULL, -- 余额
    currency   varchar                     NOT NULL, -- 货币类型
    created_at timestamptz DEFAULT (now()) NOT NULL
);

-- 条目表: 用于记录账户余额的所有更改
CREATE TABLE entries
(
    id         bigserial PRIMARY KEY,
    account_id bigint REFERENCES accounts (id) NOT NULL, -- 引用accounts表的用户id,
    amount     bigint                          NOT NULL, -- 金额, 可以是整数或者负数, 取决于是转出还是收入
    created_at timestamptz DEFAULT (now())     NOT NULL  -- 条目创建时间
);

CREATE INDEX idx_account_id ON entries (account_id);

-- 转账表: 记录两个账户之间的金额转入和转出
CREATE TABLE transfers
(
    id              bigserial PRIMARY KEY,
    from_account_id bigint REFERENCES accounts (id) NOT NULL, -- 发出转账的账户id
    to_account_id   bigint REFERENCES accounts (id) NOT NULL, -- 接收转账的账户id
    amount          bigint                          NOT NULL, -- 转账的金额, 必须是正数
    created_at      timestamptz DEFAULT (now())     NOT NULL  -- 转账创建的时间
);

CREATE INDEX accounts_owner ON accounts (owner);
CREATE INDEX transfers_from_account_id ON transfers (from_account_id);
CREATE INDEX transfers_to_account_id ON transfers (to_account_id);
CREATE INDEX entries_account_id_fkey ON entries (account_id);
CREATE INDEX transfers_compound ON transfers (from_account_id, to_account_id);
