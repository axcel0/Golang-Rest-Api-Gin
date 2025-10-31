-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: ListUsersPaginated :many
SELECT * FROM users
WHERE 
    (CASE WHEN @search != '' THEN (name LIKE '%' || @search || '%' OR email LIKE '%' || @search || '%') ELSE 1 END)
    AND (CASE WHEN @active >= 0 THEN active = @active ELSE 1 END)
ORDER BY 
    CASE WHEN @order_by = 'name' THEN name END,
    CASE WHEN @order_by = 'email' THEN email END,
    CASE WHEN @order_by = 'created_at' THEN created_at END DESC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE 
    (CASE WHEN @search != '' THEN (name LIKE '%' || @search || '%' OR email LIKE '%' || @search || '%') ELSE 1 END)
    AND (CASE WHEN @active >= 0 THEN active = @active ELSE 1 END);

-- name: CreateUser :one
INSERT INTO users (name, email, age, active)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET 
    name = COALESCE(?, name),
    email = COALESCE(?, email),
    age = COALESCE(?, age),
    active = COALESCE(?, active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
