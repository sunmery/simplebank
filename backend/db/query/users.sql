-- name: CreateUser :one
INSERT INTO users (username, full_name, hashed_password, email)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE username = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
     username = coalesce(sqlc.narg(username), username),
     full_name = coalesce(sqlc.narg(full_name), full_name),
     hashed_password = coalesce(sqlc.narg(hashed_password), hashed_password),
     email = coalesce(sqlc.narg(email), email)
WHERE username = sqlc.arg(username)
RETURNING *;

