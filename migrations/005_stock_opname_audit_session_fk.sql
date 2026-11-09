-- ============================================
-- AUDIT FK: audit ikut terhapus saat sesi dihapus
-- Additive, tapi membersihkan audit yatim lebih dulu.
-- ============================================

BEGIN;

-- Audit lama yang sesinya sudah tidak ada tidak bisa dipetakan lagi.
DELETE FROM stock_opname_audit a
WHERE NOT EXISTS (
    SELECT 1 FROM sesi s WHERE s.sesi_id = a.sesi_id
);

ALTER TABLE stock_opname_audit
    DROP CONSTRAINT IF EXISTS fk_so_audit_sesi;

ALTER TABLE stock_opname_audit
    ADD CONSTRAINT fk_so_audit_sesi
    FOREIGN KEY (sesi_id) REFERENCES sesi(sesi_id) ON DELETE CASCADE;

COMMIT;
