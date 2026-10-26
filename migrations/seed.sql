-- ============================================
-- SEED DATA: Stock Opname Rita
-- Random dev data (BUKAN dari DBF legacy;
-- file PRODUK.DBF / BARCODE.DBF dipakai
-- sebagai fixture untuk test import)
--
-- Urutan wajib: sesi -> product -> barcode, karena
-- product dan barcode sekarang milik sesi.
-- ============================================

-- ============================================
-- SESI (4 sesi dengan variasi status)
-- ============================================
INSERT INTO sesi (sesi_location, sesi_code, sesi_status, sesi_startedat, sesi_endedat) VALUES
-- COMPLETED
('Gudang A - Blok Utara', 'SO-10-09-2026-01', 'COMPLETED', '2026-09-10 08:00:00+07', '2026-09-10 12:30:00+07'),
-- COMPLETED
('Gudang A - Blok Selatan', 'SO-10-09-2026-02', 'COMPLETED', '2026-09-10 08:00:00+07', '2026-09-10 13:00:00+07'),
-- IN_PROGRESS
('Gudang B', 'SO-11-09-2026-01', 'IN_PROGRESS', '2026-09-11 09:00:00+07', NULL),
-- CANCELLED
('Lantai 1', 'SO-12-09-2026-01', 'CANCELLED', '2026-09-12 07:30:00+07', '2026-09-12 08:00:00+07');

-- ============================================
-- PRODUCT (12 produk x sesi 1, 2, 3 = 36 baris)
-- Sesi 4 CANCELLED sengaja tidak punya produk.
-- ============================================
INSERT INTO product
  (product_sesi_id, product_plu, product_name, product_department_code,
   product_buyprice, product_sellprice)
SELECT s.sesi_id, v.plu, v.name, v.dept, v.buy, v.sell
FROM (VALUES
  ('900101', 'Sari Roti Tawar Kupas 400gr', '1138', 14500.00, 17200.00),
  ('900102', 'Ultra Milk Full Cream 1L',     '1139', 18900.00, 21500.00),
  ('900103', 'Chitato Sapi Panggang 68gr',   '1138',  9500.00, 11500.00),
  ('900104', 'Pocari Sweat 500ml',           '1139',  6500.00,  8000.00),
  ('900105', 'Daia Detergen Putih 800gr',    '1140', 17500.00, 20500.00),
  ('900106', 'Pepsodent Siwak 190gr',        '1140', 22000.00, 26500.00),
  ('900107', 'Beng-Beng Coklat 25gr',        '1138',  2800.00,  3500.00),
  ('900108', 'Floridina Orange 350ml',       '1139',  3800.00,  5000.00),
  ('900109', 'Krat Botol Kaca Deposit',      '1139',     0.00,     0.00),
  ('900110', 'Tolak Angin Cair 15ml',        '1141', 12500.00, 15200.00),
  ('900111', 'Mamy Poko Pants M34',          '1141', 52000.00, 61500.00),
  ('900112', 'ABC Kecap Manis 600ml',        '1138', 21000.00, 24800.00)
) AS v(plu, name, dept, buy, sell)
JOIN sesi s ON s.sesi_id IN (1, 2, 3);

-- ============================================
-- BARCODE (19 kode x sesi 1, 2, 3 = 57 baris)
-- Produk 900109 sengaja tanpa barcode.
-- ============================================
INSERT INTO barcode (barcode_sesi_id, barcode_product_id, barcode_code)
SELECT p.product_sesi_id, p.product_id, v.code
FROM (VALUES
  ('900101', '9001015550011'),
  ('900101', '8991001010016'),
  ('900102', '9001025550018'),
  ('900102', '8991001020022'),
  ('900103', '9001035550025'),
  ('900104', '9001045550032'),
  ('900104', '8991001040046'),
  ('900105', '9001055550049'),
  ('900105', '8991001050053'),
  ('900106', '9001065550056'),
  ('900107', '9001075550063'),
  ('900107', '8991001070077'),
  ('900108', '9001085550070'),
  ('900110', '9001105550094'),
  ('900110', '8991001100106'),
  ('900111', '9001115550100'),
  ('900111', '8991001110113'),
  ('900112', '9001125550117'),
  ('900112', '8991001120120')
) AS v(plu, code)
JOIN product p
  ON p.product_plu = v.plu
 AND p.product_sesi_id IN (1, 2, 3);

-- ============================================
-- COORDINATOR (nama orang Indonesia)
-- ============================================
INSERT INTO coordinator (coor_code, coor_sesi_id, coor_status) VALUES
-- Sesi 1 (COMPLETED) -> semua coordinator COMPLETED
('Budi', 1, 'COMPLETED'),
('Andi', 1, 'COMPLETED'),
-- Sesi 2 (COMPLETED) -> semua coordinator COMPLETED
('Dedi', 2, 'COMPLETED'),
('Rina', 2, 'COMPLETED'),
-- Sesi 3 (IN_PROGRESS) -> coordinator masih IN_PROGRESS/IN_REVIEW
('Eko', 3, 'IN_PROGRESS'),
('Siti', 3, 'IN_REVIEW');

-- ============================================
-- RAK
-- ============================================
INSERT INTO rak (rak_name, rak_updatedat, rak_coor_id) VALUES
-- Sesi 1, Budi
('Rak A1', '2026-09-10 08:30:00+07', 1),
('Rak A2', '2026-09-10 08:35:00+07', 1),
('Rak B1', '2026-09-10 08:35:00+07', 1),
-- Sesi 1, Andi
('Rak C1', '2026-09-10 08:40:00+07', 2),
-- Sesi 2, Dedi
('Rak E1', '2026-09-10 08:30:00+07', 3),
('Rak E2', '2026-09-10 08:35:00+07', 3),
-- Sesi 2, Rina
('Rak F1', '2026-09-10 08:45:00+07', 4),
-- Sesi 3, Eko (IN_PROGRESS)
('Rak G1', '2026-09-11 09:20:00+07', 5);

-- ============================================
-- STOCK OPNAME (snapshot, tanpa FK ke product)
-- ============================================
INSERT INTO stock_opname (stock_opname_quantity, so_product_plu, so_product_name, so_barcode, so_buyprice, so_sellprice, stock_opname_rak_id) VALUES
-- Sesi 1, Rak A1
(48, '900101', 'Sari Roti Tawar Kupas 400gr', '8991001010016', 14500.00, 17200.00, 1),
(36, '900103', 'Chitato Sapi Panggang 68gr', '9001035550025', 9500.00, 11500.00, 1),
-- Sesi 1, Rak A2 (produk sama di rak berbeda)
(12, '900101', 'Sari Roti Tawar Kupas 400gr', '9001015550011', 14500.00, 17200.00, 2),
(24, '900102', 'Ultra Milk Full Cream 1L', '8991001020022', 18900.00, 21500.00, 2),
-- Sesi 1, Rak B1
(60, '900104', 'Pocari Sweat 500ml', '8991001040046', 6500.00, 8000.00, 3),
-- Sesi 1, Rak C1
(30, '900105', 'Daia Detergen Putih 800gr', '9001055550049', 17500.00, 20500.00, 4),
-- Sesi 2, Rak E1
(10, '900107', 'Beng-Beng Coklat 25gr', '8991001070077', 2800.00, 3500.00, 5),
(40, '900112', 'ABC Kecap Manis 600ml', '8991001120120', 21000.00, 24800.00, 5),
-- Sesi 2, Rak E2
(15, '900106', 'Pepsodent Siwak 190gr', '9001065550056', 22000.00, 26500.00, 6),
-- Sesi 2, Rak F1
(25, '900108', 'Floridina Orange 350ml', '9001085550070', 3800.00, 5000.00, 7),
-- Sesi 3 (IN_PROGRESS), Rak G1
(8, '900110', 'Tolak Angin Cair 15ml', '9001105550094', 12500.00, 15200.00, 8),
(5, '900111', 'Mamy Poko Pants M34', '8991001110113', 52000.00, 61500.00, 8);

-- ============================================
-- TANDAI RAK YANG SUDAH DIUPLOAD + SNAPSHOT
-- Satu upload per koordinator; satu rak hanya menunjuk satu upload.
-- Harus setelah seluruh INSERT stock_opname.
-- ============================================
INSERT INTO upload (upload_coor_id, upload_createdat)
SELECT c.coor_id, MAX(so.stock_opname_updatedat)
FROM coordinator c
JOIN rak r ON r.rak_coor_id = c.coor_id
JOIN stock_opname so ON so.stock_opname_rak_id = r.rak_id
GROUP BY c.coor_id;

UPDATE rak r
SET rak_upload_id = u.upload_id
FROM upload u
WHERE u.upload_coor_id = r.rak_coor_id;

UPDATE rak r
SET rak_upload_snapshot = (
        SELECT jsonb_agg(jsonb_build_object(
            'plu', so.so_product_plu,
            'barcode', so.so_barcode,
            'name', so.so_product_name,
            'qty', so.stock_opname_quantity
        ) ORDER BY so.so_product_name)
        FROM stock_opname so
        WHERE so.stock_opname_rak_id = r.rak_id
    )
WHERE r.rak_upload_id IS NOT NULL;

-- ============================================
-- VERIFIKASI
-- ============================================
-- SELECT product_sesi_id, COUNT(*) FROM product GROUP BY product_sesi_id ORDER BY 1;
-- SELECT barcode_sesi_id, COUNT(*) FROM barcode GROUP BY barcode_sesi_id ORDER BY 1;
-- SELECT COUNT(*) AS total_sesi FROM sesi;
