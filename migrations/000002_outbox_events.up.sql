CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,

    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,

    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,

    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,

    CONSTRAINT outbox_attempts_nonnegative
        CHECK (attempts >= 0)
);


CREATE INDEX idx_outbox_unpublished
ON outbox_events(created_at, id)
WHERE published_at IS NULL;