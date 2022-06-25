-- name: GetUser :one
SELECT * FROM users
WHERE national_id = $1 LIMIT 1;


-- name: CreateUser :one
INSERT INTO users (
  national_id, hashed_password, full_name, email, permission, has_voted
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;