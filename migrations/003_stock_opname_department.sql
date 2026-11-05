-- ============================================
-- DEPARTMENT SNAPSHOT: laporan PDF memakai dept, bukan harga beli
-- Additive. Backfill dari master produk yang masih ada.
-- ============================================

ALTER TABLE stock_opname
    ADD COLUMN IF NOT EXISTS so_product_dept VARCHAR(10) NOT NULL DEFAULT '';

-- Backfill snapshot hasil SO dari master produk per sesi.
UPDATE stock_opname so
SET so_product_dept = p.product_department_code
FROM rak r
JOIN coordinator c ON c.coor_id = r.rak_coor_id
JOIN product p ON p.product_sesi_id = c.coor_sesi_id
WHERE so.stock_opname_rak_id = r.rak_id
  AND p.product_plu = so.so_product_plu
  AND so.so_product_dept = '';

-- Backfill dept ke snapshot upload lama (JSONB).
UPDATE rak r
SET rak_upload_snapshot = COALESCE((
    SELECT jsonb_agg(
        item || jsonb_build_object(
            'dept',
            COALESCE(p.product_department_code, item->>'dept', '')
        )
        ORDER BY e.ord
    )
    FROM jsonb_array_elements(r.rak_upload_snapshot) WITH ORDINALITY AS e(item, ord)
    LEFT JOIN coordinator c ON c.coor_id = r.rak_coor_id
    LEFT JOIN product p
        ON p.product_sesi_id = c.coor_sesi_id
       AND p.product_plu = item->>'plu'
), '[]'::jsonb)
WHERE jsonb_typeof(r.rak_upload_snapshot) = 'array'
  AND jsonb_array_length(r.rak_upload_snapshot) > 0;
