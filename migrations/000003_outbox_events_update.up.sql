ALTER TABLE outbox_events
ADD COLUMN locked_by TEXT,
ADD COLUMN locked_until TIMESTAMPTZ;