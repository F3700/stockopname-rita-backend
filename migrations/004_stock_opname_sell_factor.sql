BEGIN;

ALTER TABLE stock_opname
    ADD COLUMN IF NOT EXISTS so_sell_factor SMALLINT NOT NULL DEFAULT 1;

ALTER TABLE stock_opname
    DROP CONSTRAINT IF EXISTS chk_stock_opname_sell_factor;

ALTER TABLE stock_opname
    ADD CONSTRAINT chk_stock_opname_sell_factor
    CHECK (so_sell_factor IN (-1, 0, 1));

CREATE INDEX IF NOT EXISTS idx_so_audit_batch
    ON stock_opname_audit(batch_id)
    WHERE batch_id IS NOT NULL;

COMMIT;
