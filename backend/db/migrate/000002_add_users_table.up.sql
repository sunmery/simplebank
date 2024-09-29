-- 用户表: 一个用户可以有多个账户, 每个用户只能创建不同货币的账户
CREATE TABLE users
(
    username            varchar PRIMARY KEY,
    full_name           varchar                     NOT NULL,
    hashed_password     varchar                     NOT NULL,
    email               varchar UNIQUE              NOT NULL,
    password_changed_at timestamptz DEFAULT (now()) NOT NULL,
    created_at          timestamptz DEFAULT (now()) NOT NULL,
    updated_at          timestamptz DEFAULT (now()) NOT NULL
);

-- 关联用户表与账户表, 关系为一对多, 即1个用户能创建多个不同货币的账户
ALTER TABLE accounts
    ADD
        FOREIGN KEY ("owner") REFERENCES users ("username");

-- 复合唯一索引, 用于限制每个用户只能创建不同货币的账户
-- CREATE UNIQUE INDEX ON accounts("owner","currency");

-- accounts表上添加一个约束，使得owner和currency的组合必须唯一, 即1个用户能创建多个不同货币的账户
ALTER TABLE accounts
    ADD
        CONSTRAINT "owner_currency_key"
            UNIQUE ("owner", "currency");
