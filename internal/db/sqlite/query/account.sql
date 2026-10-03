-- name: GetAccount :one
SELECT * FROM accounts WHERE id = ? LIMIT 1;


-- name: CreateAccount :one
INSERT INTO accounts (
    name, 
    account_type, 
    currency
) VALUES (
    ?, ?, ?
)
RETURNING *;