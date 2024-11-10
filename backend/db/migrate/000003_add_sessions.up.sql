-- 会话表: 存储和控制用户的刷新令牌
CREATE TABLE sessions
(
    id            uuid PRIMARY KEY,                     -- 会话ID
    username      varchar                     NOT NULL, -- 用户名, 关联用户表的用户名
    refresh_token varchar                     NOT NULL, -- 刷新令牌
    user_agent    varchar                     NOT NULL, -- 用户代理
    client_ip     varchar                     NOT NULL, -- 客户端IP
    is_blocked    boolean     DEFAULT false   NOT NULL,-- 是否锁定, 防止刷新令牌泄露
    expires_at    timestamptz,                          -- 刷新令牌过期时间
    created_at    timestamptz DEFAULT (now()) NOT NULL  -- 创建时间
);

ALTER TABLE users
    ADD
        FOREIGN KEY ("username") REFERENCES users ("username");