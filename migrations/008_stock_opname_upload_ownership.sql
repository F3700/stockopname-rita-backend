-- ============================================
-- UPLOAD OWNERSHIP: pemilik baris hasil SO
--
-- so_upload_id = upload yang memiliki baris ini.
-- stock_opname_rak_id tetap menjadi LOKASI baris.
--
-- Invariant: bila so_upload_id terisi, koordinator upload pemilik harus
-- sama dengan koordinator rak tempat baris berada. Invariant ini dijaga
-- oleh kode (klaim rak dibatasi satu koordinator), BUKAN constraint
-- database — tidak ada FK yang bisa menyatakan aturan itu.
--
-- rak.rak_upload_id berubah arti menjadi "host": upload terakhir yang
-- mengklaim rak. NULL berarti rak belum pernah di-upload.
-- ============================================

BEGIN;

ALTER TABLE stock_opname
    ADD COLUMN IF NOT EXISTS so_upload_id INTEGER;

ALTER TABLE stock_opname
    DROP CONSTRAINT IF EXISTS fk_stock_opname_upload;

ALTER TABLE stock_opname
    ADD CONSTRAINT fk_stock_opname_upload
    FOREIGN KEY (so_upload_id) REFERENCES upload(upload_id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_stock_opname_upload
    ON stock_opname(so_upload_id);

-- Backfill: baris yang saat ini berada di rak milik sebuah upload.
-- Perkiraan terbaik: item yang ditambahkan admin ke rak yang sudah upload
-- ikut dianggap milik upload tersebut.
UPDATE stock_opname so
SET so_upload_id = r.rak_upload_id
FROM rak r
WHERE r.rak_id = so.stock_opname_rak_id
  AND r.rak_upload_id IS NOT NULL
  AND so.so_upload_id IS NULL;

-- Bersihkan host menggantung: rak bertanda upload tetapi tidak punya baris.
-- Tanpa ini, rak lama seperti itu menjadi tidak bisa di-upload setelah
-- aturan klaim berubah menjadi "rak_upload_id IS NULL".
UPDATE rak r
SET rak_upload_id = NULL
WHERE r.rak_upload_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM stock_opname so
      WHERE so.stock_opname_rak_id = r.rak_id);

COMMIT;
