ALTER TABLE booking_reschedules
    DROP KEY idx_booking_reschedules_processed_by,
    DROP COLUMN processed_at,
    DROP COLUMN processed_by,
    DROP COLUMN admin_note,
    MODIFY COLUMN status VARCHAR(20) NOT NULL DEFAULT 'APPROVED';
