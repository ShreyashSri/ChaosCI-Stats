CREATE TABLE runs (
    id          TEXT PRIMARY KEY,
    repo        TEXT NOT NULL,
    pr_number   INT  NOT NULL,
    commit_sha  TEXT NOT NULL,
    engine      TEXT NOT NULL CHECK (engine IN ('chaosmesh','litmus')),
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE experiments (
    id            BIGSERIAL PRIMARY KEY,
    run_id        TEXT REFERENCES runs(id),
    name          TEXT NOT NULL,
    kind          TEXT NOT NULL CHECK (kind IN ('essential','extended')),
    type          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    started_at    TIMESTAMPTZ,
    finished_at   TIMESTAMPTZ
);

CREATE TABLE events (
    id              BIGSERIAL PRIMARY KEY,
    run_id          TEXT REFERENCES runs(id),
    experiment_id   BIGINT REFERENCES experiments(id),
    ts              TIMESTAMPTZ DEFAULT now(),
    level           TEXT,
    message         TEXT
);

CREATE INDEX idx_events_run_id ON events(run_id);
CREATE INDEX idx_experiments_run_id ON experiments(run_id);
