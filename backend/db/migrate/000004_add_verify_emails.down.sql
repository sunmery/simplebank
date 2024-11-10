-- 添加cascade确保如果其它表中引用了该表的记录时他们将被全部删除
DROP TABLE IF EXISTS "verify_emails" CASCADE;

alter table users
    DROP
        column is_verify_email;
