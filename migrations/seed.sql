-- ============================================
-- SEED DATA: Stock Opname Rita
-- Realistic Indonesian retail data
-- ============================================

-- ============================================
-- DEPARTMENT (perusahaan/principal)
-- ============================================
INSERT INTO department (department_code, department_name, department_desc) VALUES
('INDF', 'INDOFOOD', 'PT Indofood Sukses Makmur Tbk - Produk makanan dan minuman'),
('ASTR', 'ASTRA', 'PT Astra International Tbk - Otomotif dan elektronik'),
('MYOR', 'MAYORA', 'PT Mayora Indah Tbk - Snack, kopi, dan biskuit'),
('UNVR', 'UNILEVER', 'PT Unilever Indonesia Tbk - Produk perawatan rumah tangga'),
('KLBF', 'KALBE', 'PT Kalbe Farma Tbk - Produk kesehatan dan suplemen');

-- ============================================
-- CATEGORY
-- ============================================
INSERT INTO category (category_name, category_desc) VALUES
('Makanan', 'Produk makanan kering dan basah'),
('Minuman', 'Produk minuman kemasan dan serbuk'),
('Rumah Tangga', 'Peralatan dan perlengkapan rumah tangga'),
('Elektronik', 'Perangkat elektronik dan aksesoris'),
('Kesehatan', 'Obat-obatan dan produk kesehatan'),
('Perlengkapan Bayi', 'Produk untuk bayi dan anak');

-- ============================================
-- PRODUCT (20 produk)
-- ============================================
INSERT INTO product (product_barcode, product_name, product_buyprice, product_sellprice, product_category_id, product_department_id) VALUES
-- INDOFOOD (department_id = 1)
('8991234567890', 'Indomie Goreng Original', 2500.00, 3500.00, 1, 1),
('8991234567891', 'Indomie Kuah Soto Mie', 2500.00, 3500.00, 1, 1),
('8991234567892', 'Teh Sosro Botol 450ml', 3000.00, 4000.00, 2, 1),
('8991234567893', 'Le Minerale 600ml', 4000.00, 5000.00, 2, 1),

-- MAYORA (department_id = 3)
('8991234567894', 'Beras Premium 5kg', 55000.00, 65000.00, 1, 3),
('8991234567895', 'Gula Pasir 1kg', 14000.00, 17000.00, 1, 3),
('8991234567896', 'Kopi Kapal Api 200g', 18000.00, 22000.00, 2, 3),
('8991234567897', 'Good Day Cappuccino 10s', 15000.00, 18000.00, 2, 3),

-- UNILEVER (department_id = 4)
('8991234567898', 'Rinso Anti Noda 800g', 18000.00, 22000.00, 3, 4),
('8991234567899', 'Sunlight 755ml', 7500.00, 9000.00, 3, 4),
('8991234567900', 'Sabun Lifebuoy 100g', 3500.00, 4500.00, 3, 4),

-- ASTRA (department_id = 2)
('8991234567901', 'TWS Bluetooth Earphone', 85000.00, 125000.00, 4, 2),
('8991234567902', 'Charger HP 20W Type-C', 65000.00, 95000.00, 4, 2),
('8991234567903', 'Powerbank 10000mAh', 120000.00, 175000.00, 4, 2),

-- KALBE (department_id = 5)
('8991234567904', 'Extra Joss 10s', 12000.00, 15000.00, 2, 5),
('8991234567905', 'Hydro Coco 500ml', 8000.00, 10000.00, 2, 5),
('8991234567906', 'Paracetamol 500mg 10 Tablet', 8000.00, 12000.00, 5, 5),

-- MIX (beberapa produk dari berbagai principal)
('8991234567907', 'Minyak Goreng 2L', 28000.00, 32000.00, 1, 1),
('8991234567908', 'Tepung Terigu Segitiga Biru 1kg', 12000.00, 15000.00, 1, 1),
('8991234567909', 'Aqua 1500ml', 6000.00, 7500.00, 2, 1);

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
-- Sesi 1, Budi Santoso
('Ahmad', '2026-09-10 08:15:00+07', '2026-09-10 11:00:00+07', 1),
('Dewi', '2026-09-10 08:15:00+07', '2026-09-10 11:30:00+07', 1),
-- Sesi 1, Andi Wijaya
('Rudi', '2026-09-10 08:20:00+07', '2026-09-10 11:15:00+07', 2),
('Maya', '2026-09-10 08:20:00+07', '2026-09-10 12:00:00+07', 2),
-- Sesi 2, Dedi Kurniawan
('Hendra', '2026-09-10 08:15:00+07', '2026-09-10 12:00:00+07', 3),
-- Sesi 2, Rina Marlina
('Lina', '2026-09-10 08:25:00+07', '2026-09-10 12:30:00+07', 4),
-- Sesi 3, Eko Prasetyo (IN_PROGRESS) → inspector masih jalan
('Fajar', '2026-09-11 09:10:00+07', NULL, 5),
-- Sesi 3, Siti Nurhaliza (IN_REVIEW) → inspector sudah selesai
('Yuni', '2026-09-11 09:10:00+07', '2026-09-11 11:30:00+07', 6);

-- ============================================
-- RAK (15 rak)
-- ============================================
INSERT INTO rak (rak_name, rak_updatedat, rak_inspector_id) VALUES
-- Sesi 1, Ahmad Fauzi
('Rak A1', '2026-09-10 08:30:00+07', 1),
('Rak A2', '2026-09-10 08:35:00+07', 1),
('Rak A3', '2026-09-10 08:40:00+07', 1),
-- Sesi 1, Dewi Sari
('Rak B1', '2026-09-10 08:35:00+07', 2),
('Rak B2', '2026-09-10 08:45:00+07', 2),
-- Sesi 1, Rudi Hartono
('Rak C1', '2026-09-10 08:40:00+07', 3),
('Rak C2', '2026-09-10 08:50:00+07', 3),
-- Sesi 1, Maya Putri
('Rak D1', '2026-09-10 08:45:00+07', 4),
('Rak D2', '2026-09-10 08:55:00+07', 4),
-- Sesi 2, Hendra Susanto
('Rak E1', '2026-09-10 08:30:00+07', 5),
('Rak E2', '2026-09-10 08:35:00+07', 5),
('Rak E3', '2026-09-10 08:40:00+07', 5),
-- Sesi 2, Lina Agustina
('Rak F1', '2026-09-10 08:45:00+07', 6),
('Rak F2', '2026-09-10 08:50:00+07', 6),
-- Sesi 3, Fajar Nugroho (IN_PROGRESS)
('Rak G1', '2026-09-11 09:20:00+07', 7);

-- ============================================
-- STOCK OPNAME (40 entries)
-- ============================================
INSERT INTO stock_opname (stock_opname_quantity, stock_opname_product_id, stock_opname_rak_id) VALUES
-- Sesi 1, Rak A1 - Mie Instan
(48, 1, 1),   -- Indomie Goreng
(36, 2, 1),   -- Indomie Kuah Soto
-- Sesi 1, Rak A3 - Minuman Kemasan
(24, 3, 3),   -- Teh Sosro
(60, 4, 3),   -- Le Minerale
-- Sesi 1, Rak C1 - Sabun & Deterjen
(30, 9, 6),   -- Rinso
(50, 11, 6),  -- Sabun Lifebuoy
(40, 10, 6),  -- Sunlight
-- Sesi 1, Rak C2 - Produk Kesehatan
(20, 16, 7),  -- Paracetamol
(25, 19, 7),  -- Masker ( produk belum ada, skip)

-- Sesi 2, Rak E1 - Beras & Gula
(10, 5, 10),  -- Beras Premium
(40, 6, 10),  -- Gula Pasir
-- Sesi 2, Rak E2 - Minyak Goreng
(15, 17, 11), -- Minyak Goreng
-- Sesi 2, Rak E3 - Tepung & Bahan Kue
(20, 18, 12), -- Tepung Terigu
-- Sesi 2, Rak F1 - Kopi & Teh
(25, 7, 13),  -- Kapal Api
(30, 8, 13),  -- Good Day Cappuccino
-- Sesi 2, Rak F2 - Air Mineral
(100, 19, 14), -- Aqua

-- Sesi 3 (IN_PROGRESS), Rak G1 - Elektronik (baru sebagian)
(8, 12, 15),  -- TWS Bluetooth
(5, 13, 15),  -- Charger HP

-- Beberapa produk di rak berbeda (realistis untuk retail)
(12, 1, 2),   -- Indomie Goreng juga di Rak A2
(20, 3, 2),   -- Teh Sosro juga di Rak A2
(10, 7, 3),   -- Kopi Kapal Api juga di Rak A3
(15, 9, 12),  -- Rinso juga di Rak E3
(8, 16, 6),   -- Paracetamol juga di Rak C1
(6, 12, 11),  -- TWS juga di Rak E2
(4, 13, 10),  -- Charger HP juga di Rak E1
(7, 10, 3),   -- Lifebuoy juga di Rak A3
(3, 14, 15),  -- Powerbank juga di Rak G1
(25, 15, 3),  -- Extra Joss juga di Rak A3
(18, 16, 1);  -- Paracetamol juga di Rak A1

-- ============================================
-- VERIFIKASI
-- ============================================
-- SELECT COUNT(*) AS total_departments FROM department;
-- SELECT COUNT(*) AS total_categories FROM category;
-- SELECT COUNT(*) AS total_products FROM product;
-- SELECT COUNT(*) AS total_sesi FROM sesi;
-- SELECT COUNT(*) AS total_coordinators FROM coordinator;
-- SELECT COUNT(*) AS total_inspectors FROM inspector;
-- SELECT COUNT(*) AS total_raks FROM rak;
-- SELECT COUNT(*) AS total_stock_opname FROM stock_opname;


TRUNCATE TABLE
    stock_opname,
    deleted_product,
    rak,
    inspector,
    coordinator,
    sesi,
    product,
    category,
    department
CASCADE;


ALTER SEQUENCE department_department_id_seq RESTART WITH 1;
ALTER SEQUENCE category_category_id_seq RESTART WITH 1;
ALTER SEQUENCE product_product_id_seq RESTART WITH 1;
ALTER SEQUENCE sesi_sesi_id_seq RESTART WITH 1;
ALTER SEQUENCE coordinator_coor_id_seq RESTART WITH 1;
ALTER SEQUENCE inspector_inspector_id_seq RESTART WITH 1;
ALTER SEQUENCE rak_rak_id_seq RESTART WITH 1;
ALTER SEQUENCE stock_opname_stock_opname_id_seq RESTART WITH 1;