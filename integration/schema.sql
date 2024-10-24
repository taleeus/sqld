CREATE TABLE IF NOT EXISTS model (
    id          SERIAL      PRIMARY KEY,
    name        TEXT,
    created_at  TIMESTAMP   DEFAULT NOW()
);
