-- name: CreateVerifyEmail :one
INSERT INTO verify_emails(username, email, secret_code)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateVerifyEmail :one
-- 更新is_used为已使用(TRUE),
-- 条件是一次性密码(secret_code)相同且没有使用过(is_used = FALSE)和在有效期内(expired_at > now())
UPDATE verify_emails
SET is_used = TRUE
WHERE id = @id
  AND secret_code = @secret_code
  AND is_used = FALSE
  AND expired_at > now()
RETURNING *;
