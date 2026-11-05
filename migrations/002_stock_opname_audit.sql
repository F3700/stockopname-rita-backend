-- ============================================
-- AUDIT: koreksi hasil stock opname
-- Additive. Tidak mengubah tabel lama.
-- ============================================

CREATE TABLE IF NOT EXISTS stock_opname_audit (
    audit_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    batch_id UUID,
    stock_opname_id INTEGER,
    sesi_id INTEGER NOT NULL,
    source_rak_id INTEGER,
    target_rak_id INTEGER,
    source_coor_id INTEGER,
    target_coor_id INTEGER,
    action VARCHAR(10) NOT NULL
        CHECK (action IN ('CREATE', 'UPDATE', 'DELETE', 'MOVE')),
    before_data JSONB,
    after_data JSONB,
    actor VARCHAR(100) NOT NULL,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_so_audit_session
    ON stock_opname_audit(sesi_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_so_audit_result
    ON stock_opname_audit(stock_opname_id);

CREATE INDEX IF NOT EXISTS idx_so_audit_source_rack
    ON stock_opname_audit(source_rak_id);

CREATE INDEX IF NOT EXISTS idx_so_audit_target_rack
    ON stock_opname_audit(target_rak_id);

CREATE INDEX IF NOT EXISTS idx_so_audit_source_coor
    ON stock_opname_audit(source_coor_id);

CREATE INDEX IF NOT EXISTS idx_so_audit_target_coor
    ON stock_opname_audit(target_coor_id);
