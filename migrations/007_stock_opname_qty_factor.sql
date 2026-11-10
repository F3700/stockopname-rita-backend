-- ============================================
-- QTY FACTOR: faktor mengalikan quantity, bukan harga
-- Menggantikan arti so_sell_factor dari migration 004.
-- ============================================

BEGIN;

ALTER TABLE stock_opname
    RENAME COLUMN so_sell_factor TO so_qty_factor;

ALTER TABLE stock_opname
    DROP CONSTRAINT IF EXISTS chk_stock_opname_sell_factor;

ALTER TABLE stock_opname
    ADD CONSTRAINT chk_stock_opname_qty_factor
    CHECK (so_qty_factor IN (-1, 0, 1));

-- Data lama memakai faktor harga; reset ke netral.
UPDATE stock_opname SET so_qty_factor = 1;

COMMIT;
