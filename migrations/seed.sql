-- ============================================
-- SEED DATA: Stock Opname Rita
-- Random dev data (BUKAN dari DBF legacy;
-- file PRODUK.DBF / BARCODE.DBF dipakai
-- sebagai fixture untuk test import)
-- ============================================

-- ============================================
-- PRODUCT (random)
-- ============================================
INSERT INTO product (product_plu, product_name, product_department_code, product_buyprice, product_sellprice) VALUES
('900101', 'Sari Roti Tawar Kupas 400gr', '1138', 14500.00, 17200.00),
('900102', 'Ultra Milk Full Cream 1L', '1139', 18900.00, 21500.00),
('900103', 'Chitato Sapi Panggang 68gr', '1138', 9500.00, 11500.00),
('900104', 'Pocari Sweat 500ml', '1139', 6500.00, 8000.00),
('900105', 'Daia Detergen Putih 800gr', '1140', 17500.00, 20500.00),
('900106', 'Pepsodent Siwak 190gr', '1140', 22000.00, 26500.00),
('900107', 'Beng-Beng Coklat 25gr', '1138', 2800.00, 3500.00),
('900108', 'Floridina Orange 350ml', '1139', 3800.00, 5000.00),
('900109', 'Krat Botol Kaca Deposit', '1139', 0.00, 0.00),
('900110', 'Tolak Angin Cair 15ml', '1141', 12500.00, 15200.00),
('900111', 'Mamy Poko Pants M34', '1141', 52000.00, 61500.00),
('900112', 'ABC Kecap Manis 600ml', '1138', 21000.00, 24800.00);

-- ============================================
-- BARCODE (1-2 per produk, 1 produk tanpa barcode)
-- ============================================
INSERT INTO barcode (barcode_product_id, barcode_code) VALUES
(1, '9001015550011'), (1, '8991001010016'),
(2, '9001025550018'), (2, '8991001020022'),
(3, '9001035550025'),
(4, '9001045550032'), (4, '8991001040046'),
(5, '9001055550049'), (5, '8991001050053'),
(6, '9001065550056'),
(7, '9001075550063'), (7, '8991001070077'),
(8, '9001085550070'),
-- produk 9 (Krat Botol) sengaja tanpa barcode
(10, '9001105550094'), (10, '8991001100106'),
(11, '9001115550100'), (11, '8991001110113'),
(12, '9001125550117'), (12, '8991001120120');

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
-- COORDINATOR (nama orang Indonesia)
-- ============================================
INSERT INTO coordinator (coor_code, coor_sesi_id, coor_status) VALUES
-- Sesi 1 (COMPLETED) → semua coordinator COMPLETED
('Budi', 1, 'COMPLETED'),
('Andi', 1, 'COMPLETED'),
-- Sesi 2 (COMPLETED) → semua coordinator COMPLETED
('Dedi', 2, 'COMPLETED'),
('Rina', 2, 'COMPLETED'),
-- Sesi 3 (IN_PROGRESS) → coordinator masih IN_PROGRESS/IN_REVIEW
('Eko', 3, 'IN_PROGRESS'),
('Siti', 3, 'IN_REVIEW');

-- ============================================
-- INSPECTOR (nama orang Indonesia)
-- ============================================
INSERT INTO inspector (inspector_code, inspector_startedat, inspector_endedat, inspector_coor_id) VALUES
-- Sesi 1, Budi
('Ahmad', '2026-09-10 08:15:00+07', '2026-09-10 11:00:00+07', 1),
('Dewi', '2026-09-10 08:15:00+07', '2026-09-10 11:30:00+07', 1),
-- Sesi 1, Andi
('Rudi', '2026-09-10 08:20:00+07', '2026-09-10 11:15:00+07', 2),
('Maya', '2026-09-10 08:20:00+07', '2026-09-10 12:00:00+07', 2),
-- Sesi 2, Dedi
('Hendra', '2026-09-10 08:15:00+07', '2026-09-10 12:00:00+07', 3),
-- Sesi 2, Rina
('Lina', '2026-09-10 08:25:00+07', '2026-09-10 12:30:00+07', 4),
-- Sesi 3, Eko (IN_PROGRESS) → inspector masih jalan
('Fajar', '2026-09-11 09:10:00+07', NULL, 5),
-- Sesi 3, Siti (IN_REVIEW) → inspector sudah selesai
('Yuni', '2026-09-11 09:10:00+07', '2026-09-11 11:30:00+07', 6);

-- ============================================
-- RAK
-- ============================================
INSERT INTO rak (rak_name, rak_updatedat, rak_inspector_id) VALUES
-- Sesi 1, Ahmad
('Rak A1', '2026-09-10 08:30:00+07', 1),
('Rak A2', '2026-09-10 08:35:00+07', 1),
-- Sesi 1, Dewi
('Rak B1', '2026-09-10 08:35:00+07', 2),
-- Sesi 1, Rudi
('Rak C1', '2026-09-10 08:40:00+07', 3),
-- Sesi 2, Hendra
('Rak E1', '2026-09-10 08:30:00+07', 5),
('Rak E2', '2026-09-10 08:35:00+07', 5),
-- Sesi 2, Lina
('Rak F1', '2026-09-10 08:45:00+07', 6),
-- Sesi 3, Fajar (IN_PROGRESS)
('Rak G1', '2026-09-11 09:20:00+07', 7);

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
-- VERIFIKASI
-- ============================================
-- SELECT COUNT(*) AS total_products FROM product;
-- SELECT COUNT(*) AS total_barcodes FROM barcode;
-- SELECT COUNT(*) AS total_sesi FROM sesi;
-- SELECT COUNT(*) AS total_coordinators FROM coordinator;
-- SELECT COUNT(*) AS total_inspectors FROM inspector;
-- SELECT COUNT(*) AS total_raks FROM rak;
-- SELECT COUNT(*) AS total_stock_opname FROM stock_opname;
