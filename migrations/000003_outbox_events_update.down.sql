ALTER TABLE outbox_events
    DROP COLUMN locked_by,
    DROP COLUMN locked_until;