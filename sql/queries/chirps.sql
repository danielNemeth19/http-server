-- name: CreateChirp :one
INSERT INTO CHIRPS (id, created_at, updated_at, body, user_id)
VALUES (
	gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING *;

-- name: GetChirps :many
SELECT ID, CREATED_AT, UPDATED_AT, BODY, USER_ID
FROM CHIRPS
ORDER BY CREATED_AT ASC;
