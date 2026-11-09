-- ============================================
-- PRODUCT CARTON: jumlah produk per karton
-- ============================================

ALTER TABLE product
    ADD COLUMN IF NOT EXISTS product_per_carton INTEGER NOT NULL DEFAULT 0;

ALTER TABLE product
    DROP CONSTRAINT IF EXISTS chk_product_per_carton;

ALTER TABLE product
    ADD CONSTRAINT chk_product_per_carton
    CHECK (product_per_carton >= 0);
