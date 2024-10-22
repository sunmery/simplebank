// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package db

import (
	"context"
)

const CreateUser = `-- name: CreateUser :one
INSERT INTO users (username, full_name, hashed_password, email)
VALUES ($1, $2, $3, $4)
RETURNING username, full_name, hashed_password, email, password_changed_at, created_at, updated_at
`

type CreateUserParams struct {
	Username       string `json:"username"`
	FullName       string `json:"fullName"`
	HashedPassword string `json:"hashedPassword"`
	Email          string `json:"email"`
}

// CreateUser
//
//	INSERT INTO users (username, full_name, hashed_password, email)
//	VALUES ($1, $2, $3, $4)
//	RETURNING username, full_name, hashed_password, email, password_changed_at, created_at, updated_at
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (Users, error) {
	row := q.db.QueryRow(ctx, CreateUser,
		arg.Username,
		arg.FullName,
		arg.HashedPassword,
		arg.Email,
	)
	var i Users
	err := row.Scan(
		&i.Username,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const GetUser = `-- name: GetUser :one
SELECT username, full_name, hashed_password, email, password_changed_at, created_at, updated_at
FROM users
WHERE username = $1
LIMIT 1
`

// GetUser
//
//	SELECT username, full_name, hashed_password, email, password_changed_at, created_at, updated_at
//	FROM users
//	WHERE username = $1
//	LIMIT 1
func (q *Queries) GetUser(ctx context.Context, username string) (Users, error) {
	row := q.db.QueryRow(ctx, GetUser, username)
	var i Users
	err := row.Scan(
		&i.Username,
		&i.FullName,
		&i.HashedPassword,
		&i.Email,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
