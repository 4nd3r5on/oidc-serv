-- name: CreateUser :one
INSERT INTO users (username, password_hash, locale)
VALUES (@username, @password_hash, @locale)
RETURNING id;

-- name: GetUserByID :one
SELECT id, username, password_hash, locale
FROM users
WHERE id = @id;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, locale
FROM users
WHERE username = @username;

-- name: UpdateUser :exec
UPDATE users SET
    username      = COALESCE(sqlc.narg('username'), username),
    locale        = COALESCE(sqlc.narg('locale'), locale),
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    updated_at    = now()
WHERE id = @id;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = @id;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = @id) AS exists;
