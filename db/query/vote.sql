-- name: CreateVote :one
INSERT INTO votes (
  vote_national_id, candidate_id
) VALUES (
  $1, $2
)
RETURNING *;


-- name: ListVoteOrderByCandidate :many
SELECT 
 candidate_id,
 vote_national_id 
 FROM votes
ORDER BY candidate_id;