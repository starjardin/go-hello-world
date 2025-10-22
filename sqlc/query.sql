-- sqlc/query.sql
-- name: GetMessage :one
SELECT content FROM messages WHERE id = $1;