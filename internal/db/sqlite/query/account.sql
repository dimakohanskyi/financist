-- name: GetAccount :one
SELECT * FROM accounts WHERE id = ? LIMIT 1;
