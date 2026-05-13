-- name: CreateRun :one
INSERT INTO runs (id, repo, pr_number, commit_sha, engine, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRun :one
SELECT * FROM runs WHERE id = $1;

-- name: GetPendingRuns :many
SELECT * FROM runs WHERE status = 'pending' ORDER BY created_at ASC LIMIT 10;

-- name: UpdateRunStatus :one
UPDATE runs
SET status = $2, finished_at = $3
WHERE id = $1
RETURNING *;

-- name: CreateExperiment :one
INSERT INTO experiments (run_id, name, kind, type, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetExperimentsForRun :many
SELECT * FROM experiments WHERE run_id = $1;

-- name: UpdateExperimentStatus :one
UPDATE experiments
SET status = $2, started_at = COALESCE(started_at, $3), finished_at = $4
WHERE id = $1
RETURNING *;

-- name: CreateEvent :one
INSERT INTO events (run_id, experiment_id, level, message)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEventsForRun :many
SELECT * FROM events WHERE run_id = $1 ORDER BY ts ASC;

-- name: ListAllRuns :many
SELECT * FROM runs ORDER BY created_at DESC LIMIT 100;
