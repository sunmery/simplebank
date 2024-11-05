CREATE TABLE verify_emails
(
    id          bigserial primary key,
    username    varchar                                             NOT NULL, --用户名
    email       varchar                                             NOT NULL, -- 邮件
    secret_code varchar                                             NOT NULL, -- 一次性密码
    is_used     bool        DEFAULT false                           NOT NULL, --密码是否被使用
    created_at  timestamptz DEFAULT (now())                         NOT NULL, --创建时间
    expired_at  timestamptz DEFAULT (now() + INTERVAL '15 minutes') NOT NULL  -- 过期时间
);

-- 外键. 引用至用户表的用户名
alter table "verify_emails"
    add
        foreign key (username)
            references users (username);

-- 在用户表添加是否以验证邮件的列
alter table "users"
    add
        column is_email_verified bool DEFAULT false NOT NULL;
