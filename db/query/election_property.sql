-- name: GetElectionProperty :one
SELECT * FROM election_properties
WHERE name = $1 LIMIT 1;

-- name: UpdateElectionProperty :one
UPDATE election_properties SET value = $2
WHERE name = $1
RETURNING *;