-- =============================================================================
-- 005: PERSETUJUAN ADMIN UNTUK JADWAL ULANG
-- Pengajuan jadwal ulang kini berstatus PENDING hingga admin memutuskan.
-- Opsional — aplikasi sudah memakai GORM AutoMigrate. Hanya MySQL/MariaDB.
-- =============================================================================

ALTER TABLE booking_reschedules
    MODIFY COLUMN status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    ADD COLUMN admin_note   VARCHAR(500) NULL AFTER status,
    ADD COLUMN processed_by CHAR(36)     NULL AFTER admin_note,
    ADD COLUMN processed_at DATETIME(3)  NULL AFTER processed_by,
    ADD KEY idx_booking_reschedules_processed_by (processed_by);
