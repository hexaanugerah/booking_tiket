-- =============================================================================
-- 004: REFUND + RESCHEDULE
-- Opsional — aplikasi sudah memakai GORM AutoMigrate. File ini untuk setup SQL
-- manual. Hanya untuk MySQL/MariaDB.
-- =============================================================================

CREATE TABLE IF NOT EXISTS refunds (
    id             CHAR(36)     NOT NULL,
    created_at     DATETIME(3)  NULL,
    updated_at     DATETIME(3)  NULL,
    deleted_at     DATETIME(3)  NULL,
    booking_id     CHAR(36)     NOT NULL,
    user_id        CHAR(36)     NOT NULL,
    initiated_by   VARCHAR(10)  NOT NULL DEFAULT 'USER',
    destination_type VARCHAR(10) NOT NULL DEFAULT 'BANK',
    original_amount BIGINT      NOT NULL DEFAULT 0,
    deduction_percent INT       NOT NULL DEFAULT 0,
    deduction_amount BIGINT     NOT NULL DEFAULT 0,
    amount         BIGINT       NOT NULL DEFAULT 0,
    reason         VARCHAR(500) NOT NULL,
    bank_name      VARCHAR(100) NOT NULL,
    account_name   VARCHAR(150) NOT NULL,
    account_number VARCHAR(50)  NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    admin_note     VARCHAR(500) NULL,
    processed_by   CHAR(36)     NULL,
    processed_at   DATETIME(3)  NULL,
    PRIMARY KEY (id),
    KEY idx_refunds_booking_id (booking_id),
    KEY idx_refunds_user_id (user_id),
    KEY idx_refunds_status (status),
    KEY idx_refunds_processed_by (processed_by),
    KEY idx_refunds_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS booking_reschedules (
    id             CHAR(36)     NOT NULL,
    created_at     DATETIME(3)  NULL,
    updated_at     DATETIME(3)  NULL,
    deleted_at     DATETIME(3)  NULL,
    booking_id     CHAR(36)     NOT NULL,
    user_id        CHAR(36)     NOT NULL,
    old_visit_date VARCHAR(10)  NOT NULL,
    new_visit_date VARCHAR(10)  NOT NULL,
    reason         VARCHAR(500) NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'APPROVED',
    PRIMARY KEY (id),
    KEY idx_booking_reschedules_booking_id (booking_id),
    KEY idx_booking_reschedules_user_id (user_id),
    KEY idx_booking_reschedules_status (status),
    KEY idx_booking_reschedules_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
