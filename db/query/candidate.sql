-- name: GetCandidate :one
SELECT 
  id,
  name,
  dob,
  bio_link,
  image_url,
  policy,
  vote_count,
  create_at
FROM candidates
WHERE id = $1 LIMIT 1;

-- name: ListCandidates :many
SELECT 
  id,
  name,
  dob,
  bio_link,
  image_url,
  policy,
  vote_count,
  create_at
FROM candidates
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: ListCandidatesResult :many
SELECT 
  id,
  name,
  dob,
  bio_link,
  image_url,
  policy,
  vote_count,
  CONCAT(percentage, '%')::text as percentage,
  create_at
 FROM candidates
ORDER BY vote_count DESC;

-- name: CreateCandidate :one
INSERT INTO candidates (
  name, dob, bio_link, image_url, policy, vote_count, percentage
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateCandidate :one
UPDATE candidates SET name = $2, dob = $3, bio_link = $4, image_url = $5, policy = $6
WHERE id = $1
RETURNING   
  id,
  name,
  dob,
  bio_link,
  image_url,
  policy,
  vote_count,
  create_at;

-- name: DeleteCandidate :exec
DELETE FROM candidates
WHERE id = $1;