# Plan: Master Produk per Sesi

Status: **siap implementasi** — backend + admin; D1–D9 sudah diputuskan. Sebagian langkah 0/3/6 sudah ada di working tree (lihat catatan baseline).
Tanggal: 2026-10-23
Scope: **backend + admin web**. Client mobile out of scope (didesain ulang terpisah).
Revisi: 2026-10-23 — D3 full read-only, completion guard, delete/purge sesi, D9; D3 diperluas ke update status coordinator.
Revisi: 2026-10-23 — guard `POST /stockopname/racks`; real two-connection lock test; import phase 2 hanya untuk set barcode yang berubah; constraint `fk_product_sesi` dinamai eksplisit.
Revisi: 2026-10-25 — clear hanya `POST /products/clear?sessionId=&confirm=true` (`DELETE /products/clear` dihapus karena httprouter panic: static `/products/clear` tidak boleh satu method tree dengan `DELETE /products/:id`); catatan baseline implementasi parsial; aturan duplikat barcode per PLU disamakan dengan parser; `sesi_id` divalidasi `gte=1`; catatan race import-import dan delete sesi yang sudah punya inspector/rak.
Revisi: 2026-10-25 — scope **backend + admin saja** (mobile out of scope, didesain ulang terpisah); `productCount` admin dibaca dari `pagination.total_items` + `onProductsChanged`; delete sesi mengembalikan `200` JSON; `coor-page` read-only saat sesi final; `GET /results?sessionId=` untuk sesi tidak ada tetap `200` kosong.

> **Ringkasan revisi**
>
> - Tidak ada lagi halaman Master Data global. `/master`, `MasterPage`, dan menu sidebar dihapus.
> - Produk dikelola **di dalam sesi**: detail sesi punya section produk; pembuatan sesi memakai wizard 2 langkah (sesi + koordinator, lalu import produk).
> - Sesi tanpa produk **belum siap**: join inspector dan pembuatan hasil stock opname baru ditolak `409`.
> - Sesi `COMPLETED`/`CANCELLED` bersifat **final**: semua write (produk, join inspector, create rak, stock opname, update status coordinator) ditolak `409`; read tetap boleh.
> - Sesi final juga tidak bisa di-`PATCH`/reopen; satu-satunya write adalah delete sesi (D9).
> - Sesi tanpa produk tidak bisa di-`COMPLETED`; tombol selesai disabled dan backend mengembalikan `409`.
> - Sesi `COMPLETED`/`CANCELLED` boleh dihapus (hard purge). Sesi `IN_PROGRESS` hanya boleh dihapus jika 0 produk dan 0 hasil stock opname.
> - `PATCH/DELETE /products/:id` wajib `?sessionId=`; produk beda sesi -> `404`.
> - Clear memakai satu path: `POST /products/clear?sessionId=&confirm=true`. `DELETE /products` dan `DELETE /products/clear` tidak didaftarkan.
> - Cek status write memakai lock sesi `FOR SHARE`; completion/delete memakai `FOR UPDATE`.
> - D4-A: import dua fase; multi-pengklaim mengikuti aturan final di bawah.
> - D1 sudah selesai. Sebagian langkah 3 (lock `FindByIdForShare/ForUpdate`, `FindSesiIdByInspector`) dan 6 (guard `POST /stockopname/racks`) sudah dikerjakan di working tree; commit/stash dulu sebelum lanjut. D2, D4, D5, D7–D9 belum diimplementasikan; DB existing boleh dihapus.
> - Rollout: backend + admin saja. Client mobile didesain ulang terpisah (out of scope).

## Tujuan

- Sistem dipakai di beberapa tempat secara bersamaan.
- PLU dan barcode yang sama boleh punya arti berbeda di tempat lain.
- Satu tempat punya satu sesi aktif; master produk di-import dari DBF per sesi.
- Master produk menjadi milik sesi, bukan milik global.
- Admin membuat sesi dan langsung menyiapkan produk lewat wizard; tidak ada katalog global.
- `stock_opname` tetap snapshot dan tidak berubah saat produk dihapus/di-clear; snapshot ikut terhapus hanya jika sesinya dihapus (D9).
- Kontrak stock opname tetap memakai `rak_id`; `InspectorJoinResponse` hanya bertambah `sesi_id` (additive).
- Endpoint produk berubah: `sessionId` wajib, termasuk untuk `PATCH/DELETE` by ID.
- Client mengambil daftar produk penuh per sesi; tidak ada incremental sync.
- Sesi tanpa produk tidak bisa dipakai untuk join/scan sampai produk diimpor atau ditambahkan.

## Non-goals

- Tidak membuat tabel `location` baru.
- Tidak mengubah struktur/history `stock_opname` untuk operasi produk; hard delete sesi tetap menghapus snapshot sesi itu (D9).
- Tidak menambah status `DRAFT`; kesiapan sesi dihitung dari jumlah produk.
- Tidak membuat endpoint gabungan create-session + import; wizard memanggil dua endpoint berurutan.
- Tidak menambah autentikasi. Isolasi per sesi di sini adalah isolasi logis, bukan batas keamanan.
- Tidak mengubah export PDF/Excel/DBF.
- Tidak menyediakan shim/versi API lama untuk endpoint produk.
- Tidak ada migrasi data existing. DB lama dihapus dan dibuat ulang.
- Tidak ada tombstone/`deleted_product`.
- Tidak ada sync produk berdasarkan tanggal.
- Tidak ada halaman katalog produk lintas sesi.
- Tidak mengimplementasikan atau menyesuaikan client mobile; mobile didesain ulang terpisah.

## Keputusan yang sudah disepakati

1. `product` memiliki `product_sesi_id`; `barcode` memiliki `barcode_sesi_id`.
2. `product_id` tetap identity global. Yang unik per sesi hanya PLU dan barcode.
3. Uniqueness baru:
   - `UNIQUE (product_sesi_id, product_plu)`
   - `UNIQUE (barcode_sesi_id, barcode_code)`
4. `barcode` memakai composite FK ke `product(product_sesi_id, product_id)` supaya sesi barcode tidak bisa berbeda dari sesi produknya. Barcode tetap child dari `product`; tidak ada FK langsung `barcode -> sesi`.
5. `deleted_product` dihapus. Delete produk langsung menghapus baris; PLU yang sama boleh dibuat/di-import lagi.
6. `GET /products/sync`, `updated_after`, dan `GET /products/last-session` dihapus. Client melakukan full fetch per sesi.
7. `stock_opname` tetap snapshot; sesi diturunkan dari `rak -> inspector -> coordinator -> sesi`.
8. `InspectorJoinResponse` ditambah `sesi_id` agar client bisa fetch produk per sesi.
9. Import DBF memvalidasi barcode duplikat antar-PLU dalam satu sesi; aturan multi-pengklaim ada di D4-A.
10. `sessionId` wajib di endpoint produk yang membaca/menulis master. `PATCH/DELETE /products/:id` juga wajib `?sessionId=`.
11. D1: `GET /products/last-session` dihapus. Dashboard admin dihapus; statistik sesi pindah ke Stock Opname dan `Total Produk` tidak ditampilkan lagi.
12. D2: fresh schema + seed. DB existing boleh dihapus; tidak ada backfill atau sesi legacy.
13. Query memakai `sessionId` (camelCase), body create dan response join inspector memakai `sesi_id`.
14. D3: sesi `COMPLETED`/`CANCELLED` bersifat **final/read-only untuk semua write**: create/update/delete/import/clear produk, create inspector/rak saat join, create rak (`POST /stockopname/racks`), update status coordinator (`PATCH /stockopname/coordinators/:id`), dan create/update/delete hasil `stock_opname` ditolak `409`; `PATCH /stockopname/sessions/:id` juga ditolak `409`. Read tetap boleh. Satu-satunya write yang diizinkan adalah delete sesi (D9).
15. D4-A: import memakai dua fase. Upsert field produk; hanya produk yang set barcode barunya berbeda dari set lama yang dihapus dan di-insert ulang barcode-nya. Produk dengan set barcode sama tidak disentuh, sehingga `barcode_id` dan primary barcode stabil saat re-import. Semua dalam satu transaksi.
16. D5: halaman `/master` dihapus. Produk dikelola di detail sesi dan di wizard pembuatan sesi. Wizard: langkah 1 sesi + koordinator, langkah 2 import produk. Import adalah langkah wajib di wizard.
17. D6: sesi dengan jumlah produk `0` belum siap. Join inspector dan pembuatan hasil stock opname baru ditolak `409` sampai ada produk. Ini juga berlaku lagi setelah clear/delete menghapus semua produk.
18. D7: clear memakai satu path `POST /products/clear?sessionId=&confirm=true`; `DELETE /products` dan `DELETE /products/clear` tidak didaftarkan. `PATCH/DELETE /products/:id` wajib `?sessionId=` dan mismatch -> `404`.
19. D8: backend + admin saja. Client mobile lama tidak kompatibel untuk list produk dan akan didesain ulang terpisah (out of scope).
20. Dokumentasi yang diupdate: `docs/apispec.json` backend dan QA docs. Mobile `apispec.json` tidak diubah di plan ini.
21. D9: sesi `COMPLETED`/`CANCELLED` boleh dihapus permanen. Sesi `IN_PROGRESS` hanya boleh dihapus jika belum punya produk dan belum punya hasil stock opname. Delete memakai lock `FOR UPDATE`; child data dihapus lewat cascade, termasuk `stock_opname`.
22. Completion guard: update ke `COMPLETED` ditolak `409` jika `COUNT(product) = 0`; pengecekan dan update memakai lock `FOR UPDATE`.

---

## 1. Schema

### 1.1 Model baru

```sql
-- product
product_sesi_id INTEGER NOT NULL
-- FK: CONSTRAINT fk_product_sesi REFERENCES sesi(sesi_id) ON DELETE CASCADE (lihat 1.2)

-- barcode
barcode_sesi_id INTEGER NOT NULL
```

Tidak ada tabel `deleted_product`.

`barcode_sesi_id` adalah kolom denormalisasi untuk uniqueness per sesi. Barcode tetap milik `product`; composite FK di bawah yang menjamin `barcode_sesi_id` selalu sama dengan `product_sesi_id`.

### 1.2 Constraint

```sql
-- product: PLU unik per sesi
ALTER TABLE product
  ADD CONSTRAINT uq_product_sesi_plu
  UNIQUE (product_sesi_id, product_plu);

-- dipakai composite FK barcode
ALTER TABLE product
  ADD CONSTRAINT uq_product_sesi_product
  UNIQUE (product_sesi_id, product_id);

-- barcode: kode unik per sesi
ALTER TABLE barcode
  ADD CONSTRAINT uq_barcode_sesi_code
  UNIQUE (barcode_sesi_id, barcode_code);

ALTER TABLE barcode
  ADD CONSTRAINT fk_barcode_sesi_product
  FOREIGN KEY (barcode_sesi_id, barcode_product_id)
  REFERENCES product (product_sesi_id, product_id)
  ON DELETE CASCADE;

-- product: sesi induk, nama constraint eksplisit
ALTER TABLE product
  ADD CONSTRAINT fk_product_sesi
  FOREIGN KEY (product_sesi_id)
  REFERENCES sesi (sesi_id)
  ON DELETE CASCADE;
```

Catatan:

- Nama `fk_product_sesi` eksplisit supaya `constraintMessages`/`MapPgError` tidak bergantung pada nama auto-generated PostgreSQL.

- `UNIQUE (product_sesi_id, product_id)` memang redundant terhadap PK `product_id`, tapi PostgreSQL mensyaratkannya untuk composite FK.
- FK lama `barcode_product_id -> product(product_id)` (`fk_barcode_product`) dihapus; cukup composite FK.
- `barcode_sesi_id` tidak perlu FK langsung ke `sesi`; composite FK ke product sudah menjamin sesinya valid.
- Satu tempat satu sesi aktif tetap konvensi aplikasi. Tanpa tabel `location`, database tidak bisa enforce aturan itu.
- Untuk D9, FK anak sesi diubah menjadi cascade agar satu `DELETE FROM sesi` menghapus seluruh data operasional:
  ```sql
  fk_coordinator_sesi      ON DELETE CASCADE
  fk_inspector_coordinator ON DELETE CASCADE
  fk_rak_inspector         ON DELETE CASCADE
  fk_stock_opname_rak      ON DELETE CASCADE
  ```
  `product -> sesi` dan `barcode -> product` juga cascade. Ini **hard purge**: `stock_opname` sesi tersebut ikut terhapus. Snapshot tetap dipertahankan saat produk dihapus/di-clear, tetapi tidak saat sesinya dihapus.

### 1.3 Index

```sql
-- index lama tetap untuk search
CREATE INDEX idx_product_name ON product(product_name);
CREATE INDEX idx_product_dept ON product(product_department_code);

-- dipakai attachBarcodes
CREATE INDEX idx_barcode_product ON barcode(barcode_product_id);
```

Index sync `idx_product_updatedat` tidak dibuat karena tidak ada endpoint sync berdasarkan tanggal.

Unique constraint `(product_sesi_id, product_plu)` sudah membuat index yang dipakai untuk count/filter sesi. Setelah semua list memfilter sesi, cek query plan dengan `EXPLAIN`. Kalau search per sesi masih mahal, tambahkan composite index:

```sql
-- opsional, tergantung hasil EXPLAIN
CREATE INDEX idx_product_sesi_name
  ON product (product_sesi_id, product_name);

CREATE INDEX idx_product_sesi_dept
  ON product (product_sesi_id, product_department_code);
```

Unique constraint `(barcode_sesi_id, barcode_code)` sudah membuat index untuk lookup barcode per sesi.

### 1.4 File schema

`migrations/schema1.sql` diperbarui untuk instalasi baru:

- pindahkan tabel `sesi` **sebelum** `product`;
- hapus `UNIQUE` inline pada `product.product_plu` dan `barcode.barcode_code`;
- hapus constraint lama `fk_barcode_product`;
- tambah `product_sesi_id` dan `barcode_sesi_id`;
- tambah `uq_product_sesi_plu`, `uq_product_sesi_product`, `uq_barcode_sesi_code`, `fk_barcode_sesi_product`, dan `fk_product_sesi` (nama eksplisit);
- ubah `fk_coordinator_sesi`, `fk_inspector_coordinator`, `fk_rak_inspector`, dan `fk_stock_opname_rak` menjadi `ON DELETE CASCADE` untuk D9;
- hapus tabel `deleted_product`;
- hapus index `idx_product_updatedat`;
- tambah index baru bila diperlukan.

`migrations/seed.sql` ditulis ulang untuk DB kosong (lihat bagian 5). Tidak ada `schema2_session_master.sql`.

### 1.5 Kesiapan sesi

Tidak ada kolom/status baru. Kesiapan dihitung:

```text
session ready = COUNT(product WHERE product_sesi_id = sesi_id) > 0
```

Konsekuensi:

- Sesi baru yang belum berhasil import tetap `IN_PROGRESS`, tetapi belum siap.
- Sesi tanpa produk tidak bisa di-`COMPLETED`; completion guard mengembalikan `409`.
- Clear/delete yang menghapus semua produk membuat sesi kembali belum siap.
- Import parsial yang menghasilkan minimal satu produk valid membuat sesi siap.
- Tambah produk manual juga membuat sesi siap.
- `products_without_barcode` tidak menghalangi kesiapan; produk masih bisa dicari lewat PLU.

### 1.6 Fresh install

Karena tidak ada migrasi data, alur deploy DB adalah:

```bash
# Dev/test saja. Semua data lama akan hilang.
psql "$DATABASE_URL" -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'
psql "$DATABASE_URL" -f migrations/schema1.sql
psql "$DATABASE_URL" -f migrations/seed.sql
```

Prasyarat:

- Backup hanya jika masih ada data yang ingin disimpan secara manual.
- Pastikan tidak ada client yang sedang menulis saat schema dibuat ulang.
- `DROP SCHEMA public CASCADE` menghapus semua tabel, function, dan trigger di schema public.

### 1.7 Hapus sesi

`DELETE /stockopname/sessions/:id` memakai aturan:

| Status | Boleh dihapus? | Syarat |
|---|---|---|
| `COMPLETED` | ya | hard purge seluruh child data |
| `CANCELLED` | ya | hard purge seluruh child data |
| `IN_PROGRESS` | hanya cleanup | `COUNT(product) = 0` **dan** `COUNT(stock_opname) = 0` |
| lainnya | tidak | `409` |

- Delete adalah **hard delete**, bukan soft delete/arsip.
- Sesi `IN_PROGRESS` yang masih punya hasil tetapi produknya sudah 0 harus di-`CANCELLED` dulu, baru boleh dihapus.
- Karena FK anak memakai cascade, `stock_opname`, `rak`, `inspector`, `coordinator`, `product`, dan `barcode` sesi itu ikut terhapus.
- Ini pengecualian eksplisit dari aturan snapshot: `stock_opname` tetap ada setelah produk dihapus/di-clear, tetapi hilang jika sesinya dihapus.
- Kalau suatu saat audit history wajib disimpan, jangan pakai hard delete; perlu arsip/export atau desain soft delete terpisah.
- UI wajib memakai konfirmasi ketik kode sesi; API tetap menolak sesi aktif yang masih punya data.

---

## 2. Perubahan kode

### 2.1 Model

- `internal/model/product.go`
  - tambah `ProductSesiID int`
- `internal/model/barcode.go`
  - tambah `BarcodeSesiID int`
- Hapus model/service/repository/handler `deleted_product`.

`ProductResponse` tidak perlu menambah `sesi_id`; client sudah memegang `sessionId` yang dipakai untuk request. Kalau nanti dibutuhkan untuk debugging, bisa ditambahkan terpisah.

### 2.2 Repository

#### `ProductRepository`

```go
Save(ctx, tx, product)                                  // product.ProductSesiID diisi
Update(ctx, tx, product)
Delete(ctx, tx, id)
Clear(ctx, tx, sesiId int)                              // per sesi
FindById(ctx, tx, id)
FindByPLU(ctx, tx, sesiId int, plu string)
FindByBarcode(ctx, tx, sesiId int, code string)
FindAllForImport(ctx, tx, sesiId int)
FindAllInPageSearch(ctx, sesiId int, limit, offset int, search string)
CountBySesi(ctx, tx, sesiId int) (int, error)           // readiness + completion guard
```

`FindAllUpdatedAfter` tidak ada lagi. `Clear` hanya menghapus produk sesi, tanpa tombstone.

Perubahan SQL:

- `productColumns` dan `scanProduct` menambah `p.product_sesi_id`.
- `FindByPLU`:
  ```sql
  WHERE p.product_sesi_id = $1 AND p.product_plu = $2
  ```
- `FindByBarcode`:
  ```sql
  JOIN barcode b
    ON b.barcode_product_id = p.product_id
   AND b.barcode_sesi_id = p.product_sesi_id
  WHERE b.barcode_sesi_id = $1 AND b.barcode_code = $2
  ```
- `FindAllInPageSearch`: `WHERE p.product_sesi_id = $1 AND (...)`; query count juga memakai filter yang sama.
- `FindAllForImport`: `WHERE p.product_sesi_id = $1`.
- `Save`: insert `product_sesi_id`.
- `Clear`: jangan `TRUNCATE`, karena tidak bisa difilter dan akan mereset identity sesi lain:
  ```sql
  DELETE FROM product WHERE product_sesi_id = $1;
  ```
  Barcode terhapus lewat cascade.
- `Delete`: langsung `DELETE FROM product WHERE product_id = $1`; barcode cascade. Tidak ada tombstone.
- `FindById` tetap global karena `product_id` identity global. Pemanggil yang butuh isolasi sesi harus mengecek `ProductSesiID` setelah load.
- `CountBySesi`:
  ```sql
  SELECT COUNT(*) FROM product WHERE product_sesi_id = $1
  ```
  Dipakai untuk readiness; menerima `pgx.Tx` supaya bisa dibaca di transaksi yang sama.

#### `BarcodeRepository`

```go
SaveBatch(ctx, tx, sesiId int, productID int, codes []string)
Replace(ctx, tx, sesiId int, productID int, codes []string)
```

`INSERT INTO barcode (barcode_sesi_id, barcode_product_id, barcode_code)`.

`Replace` menghapus barcode produk di sesi itu lalu insert ulang. Untuk import dua fase, `Replace(ctx, tx, sesiId, productID, nil)` dipakai sebagai fase hapus.

#### `InspectorRepository`

Tambah:

```go
FindSesiIdByInspector(ctx, tx, inspectorId int) (int, error)
```

SQL:

```sql
SELECT c.coor_sesi_id
FROM inspector i
JOIN coordinator c ON c.coor_id = i.inspector_coor_id
WHERE i.inspector_id = $1
```

Kalau tidak ada baris, return `NotFoundError{Resource: "Inspector", ID: inspectorId}`. Dipakai `RackService.CreateRack` sebelum lock sesi.

#### `StockOpnameRepository`

Tambah:

```go
FindSesiIdByRak(ctx, tx, rakID int) (int, error)
FindSesiIdByStockOpnameID(ctx, tx, stockOpnameID int) (int, error)
CountBySesi(ctx, tx, sesiId int) (int, error)
```

SQL `FindSesiIdByRak`:

```sql
SELECT c.coor_sesi_id
FROM rak r
JOIN inspector i ON i.inspector_id = r.rak_inspector_id
JOIN coordinator c ON c.coor_id = i.inspector_coor_id
WHERE r.rak_id = $1
```

SQL `FindSesiIdByStockOpnameID`:

```sql
SELECT c.coor_sesi_id
FROM stock_opname so
JOIN rak r ON r.rak_id = so.stock_opname_rak_id
JOIN inspector i ON i.inspector_id = r.rak_inspector_id
JOIN coordinator c ON c.coor_id = i.inspector_coor_id
WHERE so.stock_opname_id = $1
```

Kalau tidak ada baris, return `NotFoundError{Resource: "Rack"}` atau `NotFoundError{Resource: "Stock opname"}` sesuai konteks.

SQL `CountBySesi`:

```sql
SELECT COUNT(*)
FROM stock_opname so
JOIN rak r ON r.rak_id = so.stock_opname_rak_id
JOIN inspector i ON i.inspector_id = r.rak_inspector_id
JOIN coordinator c ON c.coor_id = i.inspector_coor_id
WHERE c.coor_sesi_id = $1
```

`CountBySesi` dipakai untuk guard delete sesi `IN_PROGRESS` (D9).

#### `SesiRepository`

```go
FindById(ctx, id int) (*model.Sesi, error)              // pool; read/pre-check
FindByIdForShare(ctx, tx, id int) (*model.Sesi, error)  // SELECT ... FOR SHARE
FindByIdForUpdate(ctx, tx, id int) (*model.Sesi, error) // SELECT ... FOR UPDATE
Delete(ctx, tx, id int) error
```

- `FindByIdForShare` dipakai semua operasi tulis produk, stock opname, join inspector, dan update status coordinator untuk menahan update status/delete sesi tanpa memblokir sesi tulis lain.
- `FindByIdForUpdate` dipakai `SesiService.Update` (completion guard) dan `SesiService.Delete`; lock eksklusif supaya count dan perubahan status/delete konsisten.
- `FindById` tetap untuk read dan validasi keberadaan tanpa lock, memakai pool (tanpa tx).
- `Delete` cukup `DELETE FROM sesi WHERE sesi_id = $1`; child data terhapus lewat cascade (D9).
- Jangan memakai nama `FindByIdForUpdate` untuk `FOR SHARE`; itu sumber bug.

### 2.3 Service

#### Validasi sesi bersama

`ProductServiceImpl`, `ImportServiceImpl`, `StockOpnameServiceImpl`, `InspectorServiceImpl` (join inspector/rak), dan `CoordinatorServiceImpl` (update status) mendapat dependency `SesiRepository` (sudah ada di router). Ada dua jalur:

- **Read** (list): `SesiRepository.FindById(ctx, sesiId)`.
  - tidak ada -> `NotFoundError{Resource: "Session", ID: sesiId}` -> 404;
  - status tidak dicek; sesi selesai tetap boleh dibaca.
- **Write** (create/update/delete/clear/import produk; create/update/delete stock opname; create inspector/rak saat join; create rak lewat `POST /stockopname/racks`; update status coordinator):
  1. mulai transaksi;
  2. `FindByIdForShare(ctx, tx, sesiId)` -> lock `FOR SHARE`;
  3. tidak ada -> 404;
  4. status `COMPLETED`/`CANCELLED` -> `ConflictError` -> 409 (D3);
  5. lanjutkan operasi tulis di transaksi yang sama;
  6. commit/rollback melepas lock.

Jangan memakai `FindById` dari pool untuk operasi tulis. Pengecekan status harus terjadi di transaksi yang sama dengan tulis, bukan sebelum transaksi.

Urutan lock: **sesi dulu, baru produk/hasil/inspector/rak**. Ini berlaku untuk `PATCH/DELETE` by ID juga, supaya tidak ada deadlock antara update produk dan update status sesi. Completion dan delete sesi memakai `FindByIdForUpdate` (`FOR UPDATE`), bukan `FOR SHARE`.

#### `ProductService`

Interface berubah:

```go
Create(ctx, req)
Update(ctx, id, sesiId int, req)
Delete(ctx, id, sesiId int)
Clear(ctx, sesiId int)
FindAll(ctx, sesiId int, pagination, search)
```

- `Create(ctx, req)` — `ProductCreateRequest` mendapat `SesiID`.
  - mulai tx; `FindByIdForShare` + cek status (D3);
  - set `product.ProductSesiID = req.SesiID`;
  - duplicate PLU di sesi yang sama -> unique constraint -> 409.
- `Update(ctx, id, sesiId, req)`:
  - mulai tx; `FindByIdForShare` untuk `sesiId`;
  - load produk; kalau produk tidak ada **atau** `product.ProductSesiID != sesiId` -> 404;
  - setelah kepemilikan cocok, cek status `COMPLETED`/`CANCELLED` -> 409 (D3);
  - sesi produk tidak boleh pindah;
  - kalau PLU diubah ke PLU aktif lain di sesi yang sama -> unique constraint -> 409;
  - `BarcodeRepository.Replace` memakai `sesiId`.
- `Delete(ctx, id, sesiId)`:
  - mulai tx; `FindByIdForShare`;
  - load produk; produk tidak ada/mismatch -> 404;
  - setelah kepemilikan cocok, cek status -> 409 (D3);
  - langsung hapus; tidak ada tombstone.
- `Clear(ctx, sesiId)`:
  - mulai tx; `FindByIdForShare` + cek status;
  - hapus semua produk sesi; tidak ada tombstone.
  - Kalau produk menjadi 0, sesi otomatis belum siap (D6); UI harus memberi warning.
- `FindAll(ctx, sesiId, pagination, search)`:
  - validasi keberadaan sesi via `FindById`;
  - sesi selesai tetap boleh dibaca;
  - filter sesi di repository.
- `FindAllUpdatedAfter` dihapus dari interface dan implementasi.

`PATCH/DELETE` mismatch mengembalikan **404**, bukan 403, supaya tidak membocorkan keberadaan produk sesi lain.

#### `ImportService`

```go
Import(ctx, sesiId int, produkData, barcodeData []byte, dryRun bool)
```

- Parse dan validasi bentuk file; kumpulkan produk dan kandidat PLU per barcode.
- Duplicate PLU: pertahankan last occurrence, baris sebelumnya invalid. Duplicate barcode untuk PLU yang sama: baris pertama tetap dipakai, baris duplikat berikutnya invalid (perilaku parser sekarang via `seenBarcode`). Jangan mengubah perilaku ini diam-diam karena D4 memakai hasil akhirnya.
- Mulai tx; `FindByIdForShare(ctx, tx, sesiId)`; cek status D3.
- `FindAllForImport(ctx, tx, sesiId)` untuk mendapatkan pemilik barcode saat ini.
- Tentukan valid/invalid barcode sebelum menulis (D4-A):
  - kumpulkan semua PLU per barcode dari file;
  - cari pemilik barcode di DB untuk sesi itu;
  - tentukan valid/invalid tanpa bergantung urutan file;
  - barcode yang sudah dimiliki PLU lain di DB hanya boleh pindah kalau PLU pemilik lama ikut di-import dan barcode itu tidak ada lagi di set barunya;
  - produk yang akhirnya tidak punya barcode masuk `products_without_barcode`.
- `upsertProduct` mengisi `ProductSesiID = sesiId`; semua `SaveBatch`/`Replace` memakai sesi itu.

Aturan final multi-pengklaim (D4-A):

```text
Untuk setiap barcode:
  candidates = PLU di file yang memuat barcode itu
  owner      = produk di DB (sesi ini) yang saat ini memiliki barcode itu

1. candidates = 1:
   - tidak ada owner -> terima.
   - candidate == owner -> terima.
   - candidate != owner:
       - owner ada di file DAN barcode tidak ada di set baru owner -> terima (pindah).
       - selain itu -> invalid.

2. candidates > 1:
   - owner ada di candidates DAN barcode masih ada di set baru owner
       -> pertahankan owner, invalidkan kandidat lain.
   - selain itu -> invalidkan semua kandidat.
```

Konsekuensi yang sudah dipilih:

- Jika A adalah owner, A ikut import tetapi melepas barcode, lalu B dan C sama-sama mengklaim:
  - B dan C invalid;
  - pelepasan A tetap berlaku;
  - barcode tidak dimiliki siapa pun setelah import.
- Barcode baru yang hanya muncul untuk satu PLU dan belum punya owner -> diterima.
- Barcode baru yang diklaim beberapa PLU tanpa owner -> semua invalid.
- Jika A masih memiliki barcode dan B mengklaim, A dipertahankan dan B invalid.
- Jika A tidak ada di file dan B mengklaim, B invalid.
- Hasil tidak boleh bergantung urutan baris file.

Jangan memakai map `code -> PLU` yang diisi sambil jalan; kumpulkan dulu semua kandidat, owner, dan set baru.

Konflik dengan data yang sudah ada di DB **tidak cukup** hanya diserahkan ke unique constraint. Di PostgreSQL, satu unique violation membatalkan seluruh transaksi, sehingga `invalid_rows` tidak bisa dilaporkan per baris. Validasi harus dilakukan sebelum write. Kalau tetap terjadi race dan constraint memukul, kembalikan 409 dengan pesan jelas; jangan biarkan menjadi 500.

Dua import bersamaan pada sesi yang sama sama-sama memakai `FOR SHARE`, jadi keduanya bisa lolos pre-validasi; yang kalah akan kena unique constraint dan mengembalikan 409. Ini diterima. Jangan menaikkan lock import ke `FOR UPDATE` tanpa mengukur dampaknya ke write produk lain.

Urutan write (D4-A, dua fase):

1. Upsert semua produk di file dan kumpulkan `product_id` + set barcode baru.
2. Hapus barcode lama hanya untuk produk yang set barcode barunya berbeda (`Replace` dengan `codes=nil`).
3. Insert barcode baru hanya untuk produk yang set barcode barunya berbeda (`SaveBatch`).

Pecah `upsertProduct` menjadi dua bagian: upsert field produk dan penulisan barcode. Jangan biarkan `upsertProduct` memanggil `Replace`/`SaveBatch` sebelum fase 2. Produk dengan set barcode sama tidak disentuh; ini menjaga `barcode_id`/primary barcode stabil dan membuat re-import idempotent.

Dengan urutan ini, barcode yang pindah dari PLU A ke PLU B dalam satu file tidak bentrok dengan unique constraint immediate.

`dryRun` tetap menjalankan validasi dan phase 1 di dalam transaksi, lalu rollback sebelum phase 2; tidak ada perubahan permanen.

#### `StockOpnameService`

- `StockOpnameServiceImpl` mendapat dependency `SesiRepository`.
- `resolveProduct` menjadi:

```go
resolveProduct(ctx, tx, sesiId int, barcode, plu string)
```

- `Create`:
  1. mulai transaksi;
  2. `sesiId := StockOpnameRepository.FindSesiIdByRak(ctx, tx, req.RakID)`;
  3. `FindByIdForShare(ctx, tx, sesiId)`; status `COMPLETED`/`CANCELLED` -> `ConflictError` -> 409;
  4. `CountBySesi(ctx, tx, sesiId)`; kalau `0` -> `ConflictError` -> 409 `Session has no products yet`;
  5. resolve produk dalam sesi itu;
  6. snapshot dan simpan.
- `CreateByRack`:
  1. mulai transaksi;
  2. ambil `sesiId` sekali dari `req.RakID`;
  3. `FindByIdForShare` + cek status sekali;
  4. cek `CountBySesi > 0` sekali;
  5. loop item memakai sesi yang sama.
- `Update`:
  1. mulai transaksi;
  2. `sesiId := StockOpnameRepository.FindSesiIdByStockOpnameID(ctx, tx, id)`;
  3. `FindByIdForShare` + cek status; `COMPLETED`/`CANCELLED` -> 409;
  4. update kuantitas, load response, commit.
- `Delete`:
  1. mulai transaksi;
  2. `sesiId := StockOpnameRepository.FindSesiIdByStockOpnameID(ctx, tx, id)`;
  3. `FindByIdForShare` + cek status; `COMPLETED`/`CANCELLED` -> 409;
  4. delete, commit.
- `rak_id` pada tiap item tetap diabaikan seperti sekarang.
- Tidak perlu menambah `so_sesi_id`; hasil tetap diturunkan lewat join.
- Readiness hanya memblokir **pembuatan hasil baru**. `GET` hasil lama tetap boleh karena memakai snapshot dan tidak melakukan lookup produk.
- D3 sekarang mengunci **semua write**, termasuk stock opname. `POST /results`, `POST /results/racks`, `PATCH /results/:id`, dan `DELETE /results/:id` pada sesi `COMPLETED`/`CANCELLED` -> 409. Read `GET` tetap 200.
- Karena update/delete hasil memakai `FindSesiIdByStockOpnameID`, hasil lama yang rack/session-nya sudah tidak konsisten akan menghasilkan 404 yang jelas, bukan panic/foreign-key 500.
- `FindAll` (list) tidak berubah: filter `sessionId` tanpa validasi eksistensi; sesi tidak ada -> 200 list kosong (bukan 404).

#### `SesiService`

`SesiServiceImpl` mendapat dependency baru:

```go
ProductRepository     repository.ProductRepository
StockOpnameRepository repository.StockOpnameRepository
```

- `Update(ctx, id, req)`:
  1. mulai transaksi;
  2. `FindByIdForUpdate(ctx, tx, id)` -> lock `FOR UPDATE`; tidak ada -> 404;
  3. kalau `current.SesiStatus` adalah `COMPLETED`/`CANCELLED` -> `ConflictError` -> 409 `Session is final`; tidak ada reopen/update metadata;
  4. kalau `req.Status == "COMPLETED"` dan `ProductRepository.CountBySesi(ctx, tx, id) == 0` -> `ConflictError` -> 409 `Session has no products yet`;
  5. `SesiRepository.Update`, commit.
  - Lock `FOR UPDATE` mencegah race: clear/import produk tidak bisa berjalan bersamaan dengan completion.
  - `CANCELLED` tidak butuh produk; sesi `IN_PROGRESS` kosong boleh dibatalkan.
  - Sesi final hanya boleh dihapus (D9), bukan diubah statusnya.
- `Delete(ctx, id)`:
  1. mulai transaksi;
  2. `FindByIdForUpdate(ctx, tx, id)` -> lock `FOR UPDATE`; tidak ada -> 404;
  3. kalau status `IN_PROGRESS`:
     - `ProductRepository.CountBySesi > 0` **atau** `StockOpnameRepository.CountBySesi > 0` -> `ConflictError` -> 409 `Session still has data`;
     - jika keduanya `0` -> boleh cleanup;
  4. kalau status `COMPLETED`/`CANCELLED` -> boleh hard delete;
  5. `SesiRepository.Delete`, commit.
  - `SesiRepository.Delete` cukup `DELETE FROM sesi`; semua child terhapus lewat cascade.
  - `stock_opname` sesi itu ikut terhapus. Ini konsekuensi hard purge yang sudah dipilih (D9).
- Error 409 harus sampai ke handler sebagai `ConflictError`, bukan 500.

#### `InspectorService`

- `InspectorJoinResponse` tambah `SesiID int json:"sesi_id"`.
- `InspectorService` mendapat dependency `ProductRepository` untuk count dan `SesiRepository` untuk status/lock.
- `CreateInspector`: setelah coordinator ditemukan:
  1. `FindByIdForShare(ctx, tx, coordinator.CoorSesiID)`;
  2. status `COMPLETED`/`CANCELLED` -> 409 `Session is final`;
  3. cek status coordinator aktif; kalau tidak -> 409;
  4. cek `ProductRepository.CountBySesi(ctx, tx, coordinator.CoorSesiID) > 0`; kalau `0` -> `ConflictError` -> 409 `Session has no products yet`;
  5. baru simpan inspector/rak.
- `CreateInspectorByCoordinatorQR`: sama, memakai `coordinator.CoorSesiID`.
- Isi `SesiID` dari `coordinator.CoorSesiID` di response.
- Lock `FOR SHARE` menahan completion/delete selama join; kalau sesi sudah final, join ditolak 409. Ini menutup race join vs complete.
- Ini additive; client lama tetap bisa join kalau sesi sudah punya produk, tetapi fetch produk tetap butuh client yang mengirim `sessionId`.

#### `RackService`

- `RackServiceImpl` mendapat dependency `InspectorRepository` dan `SesiRepository`.
- `CreateRack` (`POST /stockopname/racks`):
  1. mulai transaksi;
  2. `sesiId := InspectorRepository.FindSesiIdByInspector(ctx, tx, request.InspectorID)`; tidak ada -> 404;
  3. `SesiRepository.FindByIdForShare(ctx, tx, sesiId)`; status `COMPLETED`/`CANCELLED` -> `ConflictError` -> 409 `Session is final`;
  4. simpan rack, commit.
- Ini menutup jalur create rak mandiri yang sebelumnya tidak memeriksa status sesi. `GET` rak tidak berubah.
- Urutan lock tetap: sesi dulu, lalu rak.

#### `CoordinatorService`

- `CoordinatorServiceImpl` mendapat dependency `SesiRepository`.
- `Update` (`PATCH /stockopname/coordinators/:id`):
  1. mulai transaksi;
  2. `CoordinatorRepository.FindById(ctx, tx, id)` (method yang sudah ada, dipakai join QR) untuk mendapat `CoorSesiID`; tidak ada -> 404;
  3. `FindByIdForShare(ctx, tx, coordinator.CoorSesiID)` + cek status; `COMPLETED`/`CANCELLED` -> `ConflictError` -> 409 `Session is final`;
  4. `CoordinatorRepository.Update`, commit.
- Tanpa guard ini, sesi final masih bisa diubah lewat status coordinator (D3 bocor).

Tidak ada `DeletedProductService`.

### 2.4 Handler dan route

Query param memakai `sessionId` agar konsisten dengan endpoint stock opname yang sudah ada. Body create memakai `sesi_id`.

| Method | Path | Perubahan |
|---|---|---|
| POST | `/products` | body tambah `sesi_id` |
| GET | `/products` | wajib `?sessionId=` |
| POST | `/products/import` | wajib `?sessionId=` |
| POST | `/products/clear` | satu-satunya path clear; wajib `?sessionId=&confirm=true` |
| PATCH / DELETE | `/products/:id` | wajib `?sessionId=`; mismatch -> 404 |
| POST | `/stockopname/results` | 409 kalau sesi belum punya produk atau sesi final |
| POST | `/stockopname/results/racks` | 409 kalau sesi belum punya produk atau sesi final |
| PATCH / DELETE | `/stockopname/results/:id` | 409 kalau sesi final |
| DELETE | `/stockopname/sessions/:id` | boleh untuk `COMPLETED`/`CANCELLED`; `IN_PROGRESS` hanya jika 0 produk + 0 hasil; 200/409 |
| PATCH | `/stockopname/sessions/:id` | 409 kalau sesi final; 409 kalau transisi `IN_PROGRESS -> COMPLETED` dengan produk 0 |
| PATCH | `/stockopname/coordinators/:id` | 409 kalau sesi final |
| POST | `/stockopname/inspectors` | response tambah `sesi_id`; 409 kalau belum ada produk atau sesi final |
| POST | `/stockopname/inspectors/join-by-coordinator-qr` | response tambah `sesi_id`; 409 kalau belum ada produk atau sesi final |
| POST | `/stockopname/racks` | 409 kalau sesi final; tidak ada perubahan kontrak request/response |

Catatan route: `DELETE /products/clear` **tidak** boleh didaftarkan. httprouter v1.3.0 panic di `NewRouter` kalau static segment (`/products/clear`) berada di method tree yang sama dengan param segment (`/products/:id`): `'clear' in new path '/products/clear' conflicts with existing wildcard ':id'`. Karena `DELETE /products/:id` dipakai, clear hanya tersedia sebagai `POST /products/clear`.

Handler:

- `ProductHandlerImpl.GetProducts`, `ImportProducts`, `ClearProducts`, `UpdateProduct`, `DeleteProduct`: parse `sessionId` dengan `helper.ParseIntQueryNotNull`.
- `ProductHandlerImpl.ClearProducts`: `confirm=true` tetap wajib; tidak ada perubahan body.
- `ProductHandlerImpl.CreateProduct`: `sesi_id` divalidasi lewat `ProductCreateRequest`.
- `SesiHandlerImpl.UpdateSesi`: 409 final session dan 409 completion guard diteruskan apa adanya.
- `SesiHandlerImpl.DeleteSesi`: sukses -> `200` + `{message, data:null}` (pakai helper response JSON yang sudah ada; `api.ts` admin selalu mem-parse JSON); sesi aktif yang masih punya data -> `409`.
- `StockOpnameHandler` meneruskan 409 status final dari service; jangan mengubahnya menjadi 500.
- Hapus `ProductHandlerImpl.SyncProducts`, `DeletedProductHandler`, dan route `/products/sync`, `/deleted/products`.
- `Router` mendaftarkan `POST /products/clear`; **tidak** mendaftarkan `DELETE /products` maupun `DELETE /products/clear`.
- `product_handler_test.go` dan `inspector_handler_*_test.go` mengikuti query/body/response baru; test deleted product dihapus.
- `router_test` memastikan `DELETE /products` dan `DELETE /products/clear` tidak ada, `POST /products/clear` ada, dan `DELETE /stockopname/sessions/:id` tetap terdaftar.

Error mapping yang harus eksplisit:

| Kondisi | Status |
|---|---|
| `sessionId` hilang atau bukan integer | 400 |
| `sessionId` valid tetapi sesi tidak ada | 404 |
| Produk tidak ada / tidak ada di sesi itu | 404 |
| PLU atau barcode duplikat dalam satu sesi | 409 |
| Sesi `COMPLETED`/`CANCELLED` untuk operasi tulis produk, join inspector, create rak, stock opname, atau update status coordinator | 409 |
| `PATCH /stockopname/sessions/:id` pada sesi `COMPLETED`/`CANCELLED` (reopen/edit) | 409 |
| Sesi belum punya produk (join/scan baru) | 409 |
| Update ke `COMPLETED` saat produk sesi masih 0 | 409 |
| Delete sesi `IN_PROGRESS` yang masih punya produk/hasil | 409 |
| Rak / hasil stock opname tidak ditemukan | 404 |

Import tidak memakai 409 untuk duplikat baris yang terdeteksi pre-validasi; duplikat tersebut dilaporkan di `invalid_rows` dengan HTTP 200. 409 hanya untuk race/constraint yang tidak terduga.

`model.MapPgError` perlu entri baru di `constraintMessages`:

```go
"uq_product_sesi_plu":       "Product PLU already exists in this session",
"uq_barcode_sesi_code":      "Barcode already exists in this session",
"fk_product_sesi":           "Session not found",
"fk_barcode_sesi_product":   "Product/session mismatch",
```

Hapus entri lama:

```go
"product_product_plu_key"
"barcode_barcode_code_key"
"fk_barcode_product"
```

Catatan: `ForeignKeyError` yang sudah ada dipetakan ke 409. Karena itu, untuk sesi yang tidak ada, gunakan `SesiRepository.FindById`/`FindByIdForShare`/`FindByIdForUpdate` dan kembalikan `NotFoundError` (404), bukan mengandalkan FK.

### 2.5 DTO

- `ProductCreateRequest`:
  ```go
  SesiID int `json:"sesi_id" validate:"required,gte=1"`
  ```
  `gte=1` dipakai supaya nilai negatif/0 ditolak 400, bukan lolos ke lookup sesi.
- `InspectorJoinResponse`:
  ```go
  SesiID int `json:"sesi_id"`
  ```
- Hapus DTO/handler/service untuk deleted product.
- Hapus parameter `updated_after` dari kontrak produk.

### 2.6 API spec dan docs

`docs/apispec.json`:

- hapus seluruh path `/products/last-session` (sudah dilakukan);
- hapus path `/products/sync` dan `/deleted/products`;
- tambah parameter `sessionId` pada `GET /products`, `POST /products/import`, `POST /products/clear`, dan `PATCH/DELETE /products/:id`;
- hapus `DELETE /products` dan `DELETE /products/clear`; `POST /products/clear` tetap satu-satunya path clear;
- tambah field `sesi_id` pada body create product;
- tambah `sesi_id` pada response join inspector;
- tambah response 400 (sessionId hilang/invalid), 404 (session/product tidak ada), dan 409 (duplikat/sesi terkunci/sesi belum punya produk/sesi aktif masih punya data);
- dokumentasikan 409 untuk `POST/PATCH/DELETE` hasil stock opname pada sesi `COMPLETED`/`CANCELLED`;
- dokumentasikan 409 untuk join inspector pada sesi final dan untuk sesi belum punya produk;
- dokumentasikan 409 untuk `PATCH /stockopname/coordinators/:id` pada sesi `COMPLETED`/`CANCELLED`;
- dokumentasikan `PATCH /stockopname/sessions/:id`: 409 untuk sesi `COMPLETED`/`CANCELLED` (tidak ada reopen) dan 409 untuk `COMPLETED` dengan produk 0;
- dokumentasikan `DELETE /stockopname/sessions/:id`: 200 untuk `COMPLETED`/`CANCELLED` atau `IN_PROGRESS` kosong, 409 untuk `IN_PROGRESS` yang masih punya produk/hasil;
- hapus schema/response deleted product;
- perbarui contoh request/response.

Mobile `apispec.json` **tidak diubah** di plan ini; mobile akan didesain ulang terpisah.

QA docs di root proyek (`D:\Programming\2026-KP-RITA\QA-TEST-CASES.md`, `QA-TEST-RESULTS.md`; di luar repo backend):

- hapus skenario `last-session`, tombstone, dan deleted product;
- hapus sisa skenario Dashboard, Master Data, `TC-NAV`, `TC-DASH`, `TC-MAST`, `Total Produk`, dan `GET /products` global;
- tambah skenario PATCH/DELETE dengan `sessionId`, clear satu path, readiness, wizard, D4-A multi-pengklaim, completion guard, dan delete sesi;
- tambah skenario sesi final: write produk, join inspector, write stock opname, dan update status coordinator -> 409, read -> 200, `PATCH` sesi -> 409 (tidak ada reopen);
- tambah skenario admin: sesi final read-only, tidak ada edit/status/reopen, hanya hapus; `coor-page` juga read-only;
- tambah skenario delete: `COMPLETED`/`CANCELLED` purge, `IN_PROGRESS` kosong boleh, `IN_PROGRESS` berisi ditolak; sukses `200`;
- tambah skenario count: tambah/import/clear produk langsung mengubah `productCount` dan status tombol "Selesaikan Sesi";
- jangan menyisakan klaim "client harus update bersamaan".

### 2.7 Router wiring

Saat ini `productService` dibuat sebelum `sesiRepository`. Ubah urutan:

```go
sesiRepository := repository.NewSesiRepository(pool)
coordinatorRepository := repository.NewCoordinatorRepository(pool)
productRepository := repository.NewProductRepository(pool)
barcodeRepository := repository.NewBarcodeRepository(pool)
inspectorRepository := repository.NewInspectorRepository(pool)
rackRepository := repository.NewRackRepository(pool)
stockOpnameRepository := repository.NewStockOpnameRepository(pool)

productService := service.NewProductService(
    productRepository, barcodeRepository, sesiRepository, pool, validate,
)
importService := service.NewImportService(
    productRepository, barcodeRepository, sesiRepository, pool,
)
productHandler := handler.NewProductHandler(productService, importService)

coordinatorService := service.NewCoordinatorService(
    coordinatorRepository, sesiRepository, pool,
)
sesiService := service.NewSesiService(
    sesiRepository, coordinatorRepository, productRepository, stockOpnameRepository, pool, validate,
)
stockOpnameService := service.NewStockOpnameService(
    stockOpnameRepository, productRepository, sesiRepository, pool, validate,
)
inspectorService := service.NewInspectorService(
    inspectorRepository, coordinatorRepository, rackRepository, productRepository, sesiRepository, pool,
)
```

Hapus seluruh pembuatan `deletedProductRepository`, `deletedProductService`, dan `deletedProductHandler`.

---

## 3. Dampak ke client

### 3.1 Admin web

#### Halaman global dihapus

- Hapus route `/master` dari `src/App.tsx`.
- Hapus menu "Master Data" dan import icon `Package` dari `src/components/layout/app-sidebar.tsx`.
- Hapus `src/pages/master/master-page.tsx`.
- Pindahkan komponen produk dari `src/pages/master/product/` ke `src/pages/stock-opname/session/product/` dan update import. Nama `ProductTabs` boleh tetap; yang penting tidak ada route `/master`.
- Tidak ada session selector global. Session ID selalu datang dari route detail sesi.

#### Section produk di detail sesi

`src/pages/stock-opname/session-page.tsx` menampilkan section "Master Produk" (pakai `ProductTabs`) dengan props:

```tsx
<ProductTabs
  sessionId={id}
  sessionStatus={session.status}
  onProductsChanged={refreshProductCount}
/>
```

- `ProductTabs` menghapus prop `isActive`.
- Semua panggilan produk memakai `sessionId`:
  - list (`GET /products?sessionId=...`);
  - create (body `sesi_id`);
  - update/delete (`?sessionId=...`);
  - import (`?sessionId=...`);
  - clear (`?sessionId=...&confirm=true`).
- Jika `sessionStatus !== "IN_PROGRESS"`:
  - tampilkan data produk read-only;
  - nonaktifkan/hilangkan tambah, edit, hapus, import, dan clear.
  - nonaktifkan tombol edit/status sesi; satu-satunya aksi tulis adalah hapus sesi (D9).
  - `StockOpnameTable` juga harus menyembunyikan/mematikan aksi tambah/edit/hapus hasil. Full lock D3 berlaku untuk semua write, bukan hanya master produk.
  - Aksi ubah status koordinator juga dinonaktifkan saat sesi final; backend menolak `PATCH /stockopname/coordinators/:id` dengan 409.
- `coor-page.tsx` memakai `sessionData.status` (sudah di-fetch lewat `getSession`) sebagai `readOnly`. Saat `readOnly`:
  - `createResultColumns` tidak menampilkan aksi update/delete;
  - `StockOpnameAddItemButton` disabled;
  - tombol status koordinator dan QR tetap memakai status coordinator (backend meng-cascade `COMPLETED`/`CANCELLED`).
- Jika `sessionStatus === "IN_PROGRESS"` dan `productCount === 0`:
  - tampilkan warning: "Sesi belum punya produk. Join inspector dan scan hasil stock opname diblokir sampai produk diimpor atau ditambahkan. Sesi juga tidak bisa diselesaikan."
  - tetap tampilkan tombol import dan tambah manual.
  - tombol "Selesaikan Sesi" disabled.
- Sumber `productCount`: jangan menambah endpoint. `session-page.tsx` memakai helper yang membaca `pagination.total_items`:
  ```ts
  // product.service.ts
  export async function getProductCount(sessionId: number) {
    const { pagination } = await getProducts(sessionId, 1, 1, "");
    return pagination.total_items;
  }
  ```
  `session-page.tsx` menambahkan `getProductCount(id)` ke `Promise.all` di `loadData`, dan menyediakan `refreshProductCount` yang hanya memanggil `getProductCount(id)` lalu `setProductCount`.
- Setelah mutasi produk di `ProductTabs` (create/update/delete/import/clear), panggil prop `onProductsChanged` supaya parent me-refresh `productCount` (bukan hanya saat mount). Tanpa ini tombol "Selesaikan Sesi" tetap disabled setelah produk pertama ditambahkan.
- `ProductClearCard` memberi warning bahwa menghapus semua produk akan membuat sesi tidak siap sampai ada produk lagi.
- `ProductAddButton`, `ProductBarcodeTable`, `ProductImportCard`, dan `ProductClearCard` menerima `sessionId` dari parent.
- `so-add-item-btn.tsx` yang sudah menerima `sessionId` harus meneruskannya ke `getProducts`:
  ```ts
  getProducts(sessionId, 1, 20, productSearch)
  ```
- `StockOpnameAddItemButton` disabled untuk `COMPLETED`/`CANCELLED`; ini sekarang sesuai D3 full lock, bukan pengecualian.
- `StockOpnameTable`/`createResultColumns` tidak boleh menampilkan aksi update/delete saat sesi final. Backend tetap mengembalikan 409 jika ada client lain yang mencoba.

#### Completion guard dan delete sesi

`session-page.tsx`:

- Tombol "Selesaikan Sesi" disabled jika status `IN_PROGRESS` dan `productCount === 0`; tampilkan alasan di tooltip/alert.
- Backend tetap sumber kebenaran: `PATCH /stockopname/sessions/:id` dengan status `COMPLETED` dan produk 0 -> 409.
- Sesi yang sudah `COMPLETED`/`CANCELLED` tidak punya tombol edit/status/reopen; backend juga menolak `PATCH` dengan 409. Satu-satunya aksi adalah hapus.
- Tambah tombol "Hapus Sesi" (danger zone) di detail sesi:
  - tampil untuk `COMPLETED`/`CANCELLED`;
  - untuk `IN_PROGRESS`, tampil hanya jika `productCount === 0` (cleanup wizard gagal); kalau masih ada hasil stock opname, backend akan menolak 409 dan admin harus membatalkan sesi dulu;
  - catat: sesi aktif dengan 0 produk tetapi sudah punya inspector/rak tetap boleh dihapus backend (lihat risiko 25); jelaskan cakupan purge di dialog;
  - konfirmasi wajib mengetik kode sesi. `AlertDialog` sekarang tidak punya input: tambahkan prop opsional `requireText` (atau dialog lokal) dan disable tombol konfirmasi sampai nilainya sama dengan `session.code`;
  - sukses (`200` + JSON) -> kembali ke `/stock-opname` dan reload list;
  - jelaskan bahwa delete bersifat permanen dan ikut menghapus hasil stock opname sesi tersebut.
- Tombol delete tidak diletakkan di list sesi; taruh di detail agar tidak sengaja.

#### Service frontend

Ubah signature agar `sessionId` tidak bisa lupa:

```ts
getProducts(sessionId, page, limit, search)
getProductCount(sessionId)
createProduct(sessionId, data)
updateProduct(id, sessionId, data)
deleteProduct(id, sessionId)
updateProductBarcodes(id, sessionId, barcodes)
importProducts(produk, barcode, sessionId, dryRun)
clearProducts(sessionId)
deleteSession(id)
```

`ProductImportCard` dan `ProductClearCard` memakai `sessionId` dari props, bukan state global.

- `clearProducts(sessionId)` memanggil `POST /products/clear?sessionId=<id>&confirm=true`. Jangan memakai `DELETE /products` atau `DELETE /products/clear`.

#### Wizard pembuatan sesi

`SessionAddForm` / `SessionAddBtn` diubah menjadi wizard 2 langkah:

1. **Langkah 1 — Sesi + koordinator**
   - kode, lokasi, daftar kode koordinator (field yang sudah ada);
   - tombol "Lanjut".
2. **Langkah 2 — Import produk**
   - upload `PRODUK.DBF` dan `BARCODE.DBF`;
   - kedua file wajib; tidak ada tombol "Lewati";
   - tombol final "Buat Sesi".

Alur submit:

1. Simpan dulu data langkah 1 dan file di state; **jangan** membuat sesi saat pindah ke langkah 2.
2. Saat tombol final ditekan:
   - `const session = await createSession(step1)`;
   - `await importProducts(produk, barcode, session.id)`;
   - kalau sukses: tutup dialog, reload daftar sesi, arahkan ke detail sesi, tampilkan toast sukses;
   - kalau import gagal: sesi **tetap ada**; tutup dialog, arahkan ke detail sesi, tampilkan error, dan biarkan admin retry import atau tambah produk manual;
   - kalau `createSession` gagal: tidak ada sesi; tetap di wizard dan tampilkan error.
3. Import parsial dengan minimal satu produk valid dianggap sukses; tampilkan `invalid_rows` sebagai warning.
4. `dry_run` tidak perlu di wizard; validasi/import langsung. Dry-run tetap tersedia di section produk.

Konsekuensi yang diterima:

- "Wajib import" adalah aturan wizard, bukan constraint database. Karena create dan import dua endpoint terpisah, backend tidak bisa menolak `POST /stockopname/sessions` hanya karena produk belum ada.
- Sesi yang gagal import tetap `IN_PROGRESS`, tetapi belum siap; join/scan diblokir oleh readiness check.
- Admin boleh menambah produk manual dari detail sesi untuk memulihkan sesi, atau menghapus sesi kosong itu (D9).
- Sesi yang gagal import tidak bisa di-`COMPLETED` sampai punya minimal satu produk.

### 3.2 Client mobile — out of scope

Mobile lama **tidak diimplementasikan di plan ini**; client mobile akan didesain ulang terpisah. Yang perlu dijaga hanya kontrak backend untuk client berikutnya:

- `InspectorJoinResponse` berisi `sesi_id`; client menyimpannya setelah join.
- Produk dibaca per sesi: `GET /products?sessionId=...`, tanpa `/products/sync`, `updated_after`, atau tombstone.
- `POST /stockopname/results` dan `POST /stockopname/results/racks` bisa `409` saat sesi belum punya produk atau sesi final.
- `PATCH/DELETE /results/:id` bisa `409` saat sesi final.
- `POST /stockopname/inspectors` bisa `409` saat sesi belum siap atau final; client menampilkan pesan, bukan retry buta.
- Sesi yang dihapus admin membuat request produk berikutnya `404`.

Tidak ada pekerjaan kode, DB, atau spec mobile di plan ini.

### 3.3 Rollout

- Backend + admin web dirilis bersama.
- Mobile tidak termasuk rollout ini; mobile lama bisa `400` di `/products` karena `sessionId` wajib. Jangan deploy backend ke lapangan jika mobile lama masih menjadi satu-satunya client produk.
- Setelah backend + admin naik, admin harus membuat sesi + import produk sebelum inspector boleh join. Readiness check menolak join/scan pada sesi tanpa produk.
- Client mobile berikutnya mengikuti kontrak di §3.2; desain ulang dan jadwalnya terpisah dari plan ini.

---

## 4. Perilaku yang harus dijaga

- `stock_opname` tetap bisa dibaca setelah produk dihapus atau master di-clear.
- `product_id` tetap unik global; cache client berdasarkan `product_id` tidak bentrok.
- PLU yang sama di dua sesi menghasilkan `product_id` berbeda. Lookup client yang memakai PLU atau barcode harus di-scope dengan `sesi_id`.
- `Clear` sesi A tidak boleh menghapus produk sesi B.
- `Delete` produk langsung menghapus baris; PLU yang sama boleh dibuat/di-import lagi.
- Scan barcode di rak sesi A tidak boleh menemukan produk sesi B.
- Import sesi A tidak boleh mengubah produk sesi B.
- `PATCH/DELETE /products/:id` tanpa `sessionId` -> 400; produk beda sesi -> 404.
- `DELETE /products` dan `DELETE /products/clear` tidak ada; clear hanya lewat `POST /products/clear?sessionId=`.
- Jika barcode yang sama dipakai beberapa PLU dalam satu file import:
  - kalau salah satunya pemilik di DB dan pemilik masih memakai barcode itu -> pemilik dipertahankan, yang lain invalid;
  - kalau tidak ada pemilik -> semua invalid;
  - kalau pemilik melepas barcode dan ada lebih dari satu pengklaim -> semua pengklaim invalid, pelepasan pemilik tetap berlaku.
- Import tidak digagalkan oleh baris invalid; hasilnya `invalid_rows`.
- Perpindahan barcode dari PLU A ke PLU B dalam satu import harus berhasil lewat dua fase (D4-A).
- Sesi `COMPLETED`/`CANCELLED`: semua operasi tulis (produk, join inspector, dan stock opname) ditolak 409; `PATCH` sesi juga 409; read tetap boleh (D3). Satu-satunya write adalah delete sesi (D9).
- Sesi `IN_PROGRESS` dengan `COUNT(product) = 0`:
  - join inspector ditolak 409;
  - pembuatan hasil stock opname baru ditolak 409;
  - list produk tetap 200 (kosong);
  - tambah/import produk tetap boleh;
  - update ke `COMPLETED` ditolak 409.
- Import parsial dengan minimal satu produk valid -> sesi siap.
- Tambah produk manual -> sesi siap.
- Clear/delete semua produk -> sesi kembali belum siap.
- Delete sesi:
  - `COMPLETED`/`CANCELLED` -> 200, hard purge semua child termasuk `stock_opname`;
  - `IN_PROGRESS` dengan 0 produk dan 0 hasil -> 200;
  - `IN_PROGRESS` yang masih punya produk/hasil -> 409.
- `stock_opname` tetap ada setelah produk dihapus/di-clear, tetapi ikut terhapus saat sesinya dihapus (D9).
- `GET /products` dengan `sessionId` yang tidak ada -> 404, bukan list kosong. `GET /stockopname/results?sessionId=` tetap 200 kosong untuk sesi yang tidak ada (tidak divalidasi; hanya filter).
- Tidak ada endpoint sync atau deleted product.
- `Total Produk` tidak lagi ditampilkan; statistik Stock Opname hanya `Total Sesi`, `Selesai`, dan `Proses`.

---

## 5. Seed

`migrations/seed.sql` ditulis ulang untuk DB kosong. Urutan:

1. Insert 4 sesi terlebih dahulu.
2. Insert katalog produk untuk sesi 1, 2, dan 3 (sesi 4 `CANCELLED` tidak perlu).
3. Insert barcode untuk tiap salinan produk.
4. Insert coordinator, inspector, rak, dan `stock_opname` seperti sekarang. `stock_opname` tetap snapshot dan tidak memakai FK ke product.
5. Tidak ada `deleted_product`.

Contoh pola agar tidak menulis 36 baris manual:

```sql
INSERT INTO product
  (product_sesi_id, product_plu, product_name, product_department_code,
   product_buyprice, product_sellprice)
SELECT s.sesi_id, v.plu, v.name, v.dept, v.buy, v.sell
FROM (VALUES
  ('900101', 'Sari Roti Tawar Kupas 400gr', '1138', 14500.00, 17200.00),
  ('900102', 'Ultra Milk Full Cream 1L',    '1139', 18900.00, 21500.00)
  -- lanjutkan 12 produk
) AS v(plu, name, dept, buy, sell)
JOIN sesi s ON s.sesi_id IN (1, 2, 3);

INSERT INTO barcode (barcode_sesi_id, barcode_product_id, barcode_code)
SELECT p.product_sesi_id, p.product_id, v.code
FROM (VALUES
  ('900101', '9001015550011'),
  ('900101', '8991001010016')
  -- lanjutkan semua barcode
) AS v(plu, code)
JOIN product p ON p.product_plu = v.plu
              AND p.product_sesi_id IN (1, 2, 3);
```

Hasil yang diharapkan: 36 produk (12 x 3 sesi), 57 barcode (19 x 3 sesi), dan barcode yang sama muncul di tiga sesi. Ini sekaligus menjadi bukti isolasi.

Catatan: contoh ini mengasumsikan DB kosong dan identity sesi dimulai dari 1. Kalau tidak, gunakan `RETURNING sesi_id` atau sesuaikan ID sesi secara eksplisit.

Verifikasi seed:

```sql
SELECT product_sesi_id, COUNT(*) FROM product GROUP BY product_sesi_id ORDER BY 1;
SELECT barcode_sesi_id, COUNT(*) FROM barcode GROUP BY barcode_sesi_id ORDER BY 1;
```

---

## 6. Test

### 6.1 Test yang harus diperbarui

- `internal/service/fakes_test.go` — tambah `SesiRepository.FindByIdForShare`, `SesiRepository.FindByIdForUpdate`, `SesiRepository.Delete`, `ProductRepository.CountBySesi`, `StockOpnameRepository.FindSesiIdByStockOpnameID`, dan `StockOpnameRepository.CountBySesi`; hapus fake deleted product.
- `internal/handler/fakes_test.go` — hapus fake deleted product.
- Semua test repository yang memakai pgxmock dan mengecek SQL/argumen.
- `internal/service/so_service_write_test.go` — tambah ekspektasi `FindSesiIdByRak`/`FindSesiIdByStockOpnameID`, `FindByIdForShare`, dan `CountBySesi`; tambah kasus sesi final 409 untuk create/update/delete.
- `internal/service/sesi_service_test.go` — tambah completion guard, final-session update 409, dan delete policy; update constructor dengan dependency baru.
- `internal/service/import_service_test.go` — signature baru, pre-validasi, dan duplicate barcode.
- `internal/service/product_service_write_test.go` — `sesiId`, session validation, ownership mismatch.
- `internal/service/inspector_service_test.go` dan `inspector_service_qr_test.go` — assert `sesi_id` di response, readiness 409, final 409, dan constructor dengan `SesiRepository`.
- `internal/service/coordinator_service_test.go` — update status pada sesi final -> 409; sesi aktif tetap sukses; constructor dengan `SesiRepository`.
- `internal/handler/product_handler_test.go`, `sesi_handler_test.go`, dan `inspector_handler_*_test.go` — query/body/response baru; status delete 200/409.
- `internal/model/errors_test.go` — pesan untuk constraint baru.
- `internal/repository/product_repository_*_test.go`, `barcode_repository_test.go`, `sesi_repository_test.go`, `so_repository_*_test.go` — filter sesi, argumen SQL, `FindByIdForShare`/`FindByIdForUpdate`, `CountBySesi`.
- Hapus `deleted_product_service_test.go`, `deleted_product_repository_test.go`, `deleted_product_handler_test.go`, dan test sync.
- `router_test.go` — `DELETE /products` dan `DELETE /products/clear` tidak ada; `POST /products/clear` ada; `DELETE /stockopname/sessions/:id` ada.

### 6.2 Test baru

1. **Isolasi list**: produk sesi A tidak muncul di list sesi B.
2. **Isolasi resolve**: barcode hanya ada di sesi B; scan di rak sesi A menghasilkan 404.
3. **Barcode sama, sesi beda**: dua produk di dua sesi memakai barcode yang sama; masing-masing rak resolve ke produk yang benar.
4. **Import per sesi**: import sesi A tidak mengubah produk sesi B.
5. **Duplicate barcode antar-PLU**: file import berisi satu barcode untuk dua PLU. Kalau salah satunya pemilik di DB, baris pemilik dipertahankan dan baris lain masuk `invalid_rows`; kalau tidak ada pemilik, semua baris kode itu masuk `invalid_rows`. Import tetap jalan dan hasilnya tidak bergantung urutan file.
6. **Multi-pengklaim dengan owner melepas**: A pemilik, A ikut import tetapi melepas barcode, B dan C mengklaim. B/C invalid, A release tetap berlaku, barcode tidak dimiliki siapa pun. Ini test khusus keputusan final.
7. **Duplicate barcode vs DB**: barcode sudah dipakai produk lain di sesi yang sama; baris masuk `invalid_rows`, import tetap jalan.
8. **Barcode pindah PLU**: barcode pindah dari PLU A ke PLU B dalam satu file; import berhasil dan barcode hanya dimiliki B. Kalau PLU A tidak ada di file atau masih memakai barcode itu, baris B masuk `invalid_rows`.
9. **Barcode baru tanpa owner**: satu PLU mengklaim barcode yang belum ada di DB -> diterima.
10. **Clear per sesi**: clear A tidak menghapus produk B.
11. **Delete lalu create/import ulang**: delete produk, lalu PLU yang sama bisa dibuat/di-import lagi tanpa tombstone.
12. **Join inspector**: response manual dan QR berisi `sesi_id` yang benar.
13. **Readiness join**: sesi 0 produk -> 409; setelah tambah produk -> join berhasil.
14. **Readiness scan**: sesi 0 produk -> `POST /results` dan `POST /results/racks` 409.
15. **Readiness clear**: sesi punya produk; clear semua; join/scan kembali 409; tambah produk -> siap lagi.
16. **Readiness import parsial**: import menghasilkan minimal satu produk valid dan beberapa `invalid_rows`; sesi dianggap siap.
17. **CreateByRack**: semua item memakai sesi dari `rak_id` terluar.
18. **Session tidak ada**: missing/invalid `sessionId` -> 400; valid tapi tidak ada -> 404.
19. **Sesi terkunci**: create/update/delete/import/clear produk di sesi `COMPLETED`/`CANCELLED` ditolak 409 (D3).
20. **PATCH/DELETE ownership**: missing `sessionId` -> 400; produk beda sesi -> 404; sesi benar -> sukses.
21. **Clear route**: `POST /products/clear?sessionId=&confirm=true` sukses; `DELETE /products` dan `DELETE /products/clear` tidak terdaftar (httprouter panic kalau static `/products/clear` satu method tree dengan `/products/:id`).
22. **Fresh seed**: DB kosong + `schema1.sql` + `seed.sql` menghasilkan 36 produk dan 57 barcode.
23. **Full list pagination**: semua halaman `GET /products` bisa diambil dan hanya berisi sesi yang diminta.
24. **dryRun D4**: `dryRun=true` tidak mengubah barcode meski file berisi perpindahan barcode.
25. **D3 full lock stock opname**: create/update/delete hasil di sesi `COMPLETED`/`CANCELLED` -> 409; `GET` hasil -> 200.
26. **Lock D3**: test/mock memastikan operasi tulis produk dan stock opname memakai `FindByIdForShare` dengan `FOR SHARE`, bukan `FindById` pool; real two-connection test membuktikan lock menahan `UPDATE sesi`.
27. **Completion guard**: `PATCH session -> COMPLETED` saat produk 0 -> 409; setelah tambah produk -> sukses. Mock memastikan `FindByIdForUpdate` (`FOR UPDATE`) dipakai sebelum count/update.
28. **Delete completed**: sesi `COMPLETED` dengan produk, barcode, coordinator, inspector, rak, dan `stock_opname` -> 200; semua child hilang; tidak ada orphan.
29. **Delete cancelled**: sama untuk `CANCELLED`; hasil stock opname ikut terhapus.
30. **Delete active**: `IN_PROGRESS` dengan 0 produk/0 hasil -> 200; dengan produk -> 409; dengan hasil stock opname -> 409.
31. **Delete tidak ada**: sesi tidak ditemukan -> 404; tidak ada efek samping.
32. **Delete lock**: `SesiService.Delete` memakai `FindByIdForUpdate` (`FOR UPDATE`), bukan `FindById`.
33. **Completion vs clear**: test/smoke memastikan clear yang berjalan bersamaan tidak bisa membuat sesi `COMPLETED` dengan produk 0 (lock `FOR UPDATE`).
34. **Final session update**: `PATCH` sesi `COMPLETED`/`CANCELLED` (termasuk ke `IN_PROGRESS`) -> 409; delete sesi tetap 200.
35. **Final join**: sesi final dengan coordinator aktif (data inconsistent) -> join 409; tidak ada inspector/rak baru.
36. **Final coordinator update**: `PATCH /stockopname/coordinators/:id` pada sesi final -> 409; status coordinator tidak berubah. Di sesi `IN_PROGRESS` tetap sukses.
37. **Final rack create**: `POST /stockopname/racks` pada sesi `COMPLETED`/`CANCELLED` -> 409; tidak ada rak baru. Di sesi `IN_PROGRESS` tetap 201.

### 6.3 Test manual (wajib)

`pgxmock` tidak bisa membuktikan unique constraint asli atau row lock. Jalankan smoke test di Postgres dan real lock test di bawah. Jangan hardcode `product_id` atau `sesi_id`; ambil dari `RETURNING`. Untuk langkah stock opname, gunakan sesi yang dibuat lewat API/seed sehingga punya coordinator, inspector, dan rak; setup SQL di bawah hanya untuk menguji unique/index.

**Real two-connection lock test (wajib).** Test ini membuka dua koneksi: tx1 memanggil `FindByIdForShare`, tx2 mencoba `UPDATE sesi` dengan timeout 300ms. tx2 harus timeout selama tx1 memegang lock; setelah tx1 rollback, update harus sukses. Ini membuktikan `FOR SHARE` benar-benar menahan completion, bukan hanya mencocokkan string SQL.

```bash
TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/repository -run TestSesiFindByIdForShareBlocksUpdate -v
```

Contoh aman:

```sql
-- Jalankan di DB test, jangan di produksi.
-- Setup data; commit supaya bisa dicek lewat API.
BEGIN;

-- Buat dua sesi dan catat ID-nya.
WITH s AS (
  INSERT INTO sesi (sesi_location, sesi_code, sesi_status)
  VALUES ('Tempat A', 'SMOKE-A', 'IN_PROGRESS')
  RETURNING sesi_id
)
SELECT sesi_id FROM s;

-- Ganti :sesi_a dengan hasil RETURNING di atas.
-- Ulangi untuk sesi B.

-- Produk sama PLU/barcode, beda sesi.
WITH p AS (
  INSERT INTO product
    (product_sesi_id, product_plu, product_name, product_department_code,
     product_buyprice, product_sellprice)
  VALUES (:sesi_a, '777001', 'Produk A', '1', 1000, 2000)
  RETURNING product_id, product_sesi_id
)
INSERT INTO barcode (barcode_sesi_id, barcode_product_id, barcode_code)
SELECT product_sesi_id, product_id, '777001000001' FROM p;

-- Ulangi untuk sesi B dengan PLU dan barcode yang sama.

COMMIT;
```

Cek resolusi per sesi:

```sql
SELECT p.product_sesi_id, p.product_plu
FROM product p
JOIN barcode b ON b.barcode_product_id = p.product_id
WHERE b.barcode_code = '777001000001'
ORDER BY p.product_sesi_id;
```

Cek unique per sesi; statement ini harus gagal. Jalankan terpisah, jangan di dalam transaksi setup:

```sql
INSERT INTO barcode (barcode_sesi_id, barcode_product_id, barcode_code)
VALUES (
  :sesi_a,
  (SELECT product_id FROM product WHERE product_sesi_id = :sesi_a AND product_plu = '777001'),
  '777001000001'
);
```

Cek delete lalu create/import ulang:

```sql
DELETE FROM product
WHERE product_sesi_id = :sesi_a AND product_plu = '777001';

-- Ini harus berhasil karena tidak ada tombstone.
INSERT INTO product
  (product_sesi_id, product_plu, product_name, product_department_code,
   product_buyprice, product_sellprice)
VALUES (:sesi_a, '777001', 'Produk A Baru', '1', 1000, 2000);
```

Cek readiness lewat API/manual:

1. Buat sesi baru tanpa produk; join inspector -> 409; `POST /stockopname/results` -> 409.
2. Import produk parsial (minimal satu valid); join/scan -> berhasil.
3. Clear semua produk; join/scan -> 409 lagi.
4. Tambah produk manual; join/scan -> berhasil lagi.
5. PATCH/DELETE produk dengan `sessionId` sesi lain -> 404.
6. `DELETE /products` -> route tidak ditemukan; `POST /products/clear?sessionId=...&confirm=true` -> sukses.

7. Coba `PATCH session -> COMPLETED` saat produk 0 -> 409; tambah satu produk -> selesai berhasil.
8. Di sesi `COMPLETED` yang punya rak dan hasil: `PATCH -> IN_PROGRESS` -> 409; join inspector -> 409; `PATCH /stockopname/coordinators/:id` -> 409; `POST /stockopname/results` -> 409; `GET` hasil -> 200.
9. Delete sesi `COMPLETED` yang punya hasil -> 200; cek `product`, `barcode`, `stock_opname`, `rak`, `inspector`, `coordinator` sudah tidak ada.
10. Buat sesi `IN_PROGRESS` kosong -> delete -> 200; buat sesi aktif dengan produk -> delete -> 409.
11. `GET /products?sessionId=<tidak ada>` -> 404; `GET /stockopname/results?sessionId=<tidak ada>` -> 200 dengan list kosong.
12. Admin: tambah/import/clear produk di detail sesi; `productCount` dan tombol "Selesaikan Sesi" langsung ikut berubah tanpa reload halaman.
13. Admin: hapus sesi hanya bisa dikonfirmasi setelah kode sesi diketik; `coor-page` tidak menampilkan aksi tulis saat sesi final.

Setelah selesai, reset DB test atau hapus sesi smoke; cascade menghapus seluruh child:

```sql
DELETE FROM sesi WHERE sesi_code IN ('SMOKE-A', 'SMOKE-B');
```

Perintah verifikasi kode:

```bash
go build ./...
go test ./...
go vet ./...
```

Verifikasi admin web:

```bash
npm run build
```

---

## 7. Urutan implementasi

0. Konsolidasikan pekerjaan parsial yang sudah ada di working tree (D1 cleanup, `SesiRepository.FindByIdForShare/ForUpdate`, `InspectorRepository.FindSesiIdByInspector`, guard `RackService.CreateRack`, `sesi_lock_integration_test.go`); commit/stash dulu. Hapus sisa kode sync/tombstone/deleted_product; tidak ada migrasi.
1. Update `schema1.sql` dan tulis ulang `seed.sql`.
   - Pindahkan `sesi` sebelum `product`.
   - Hapus global unique, old FK, `deleted_product`, `idx_product_updatedat`.
   - Tambah kolom/constraint baru.
   - Ubah FK anak sesi + `stock_opname` menjadi `ON DELETE CASCADE`.
2. Update model (`ProductSesiID`, `BarcodeSesiID`).
3. Ubah interface dan implementasi repository + test repository:
   - `SesiRepository.FindByIdForShare`, `FindByIdForUpdate`, `Delete`;
   - `ProductRepository` filter sesi, `Clear(sesiId)`, `CountBySesi`;
   - `BarcodeRepository` argumen sesi;
   - `StockOpnameRepository.FindSesiIdByRak`, `FindSesiIdByStockOpnameID`, `CountBySesi`.
4. Ubah `ProductService`:
   - lock `FindByIdForShare` + status D3 di transaksi;
   - `PATCH/DELETE` wajib `sessionId`, mismatch 404;
   - clear per sesi;
   - hapus sync/deleted.
5. Ubah `ImportService`:
   - signature `sesiId`;
   - pre-validasi D4-A multi-pengklaim;
   - dua fase;
   - dry-run.
6. Ubah `StockOpnameService`, `InspectorService`, `RackService`, `CoordinatorService`, dan `SesiService`:
   - stock opname: status D3 untuk create/update/delete + readiness count;
   - `FindSesiIdByRak` / `FindSesiIdByStockOpnameID`;
   - `sesi_id` di join response;
   - join inspector: `FindByIdForShare` + final 409 + readiness;
   - create rak mandiri: `FindSesiIdByInspector` + `FindByIdForShare` + final 409;
   - update status coordinator: `FindByIdForShare` + final 409;
   - completion guard (`FOR UPDATE`) + final session update 409;
   - delete policy + cascade.
7. Ubah DTO, handler, dan router:
   - query/body `sessionId`/`sesi_id`;
   - satu path clear;
   - `DeleteSesi` 200/409;
   - reorder wiring/dependency baru.
8. Update `docs/apispec.json` dan QA docs. Mobile `apispec.json` tidak diubah.
9. Update semua fake/test; tambah test isolasi, readiness, ownership, D4 multi-pengklaim, completion guard, final update 409, final join 409, final rack 409, delete/cascade.
10. Jalankan `go test ./...`, `go vet ./...`, lalu smoke test Postgres + real two-connection lock test (`TEST_DATABASE_URL=... go test ./internal/repository -run TestSesiFindByIdForShareBlocksUpdate -v`).
11. Reset DB test, jalankan `schema1.sql`, `seed.sql`, dan smoke test.
12. Update admin web:
    - hapus `/master`, `MasterPage`, menu;
    - pindahkan section produk ke detail sesi;
    - wizard 2 langkah;
    - service frontend memakai `sessionId`;
    - `getProductCount` + `onProductsChanged` supaya count/tombol selesai selalu segar;
    - warning readiness/read-only;
    - disable tombol selesai saat `productCount === 0`;
    - tombol hapus sesi + konfirmasi ketik kode (`requireText`);
    - disable edit/status/reopen sesi final;
    - disable semua aksi stock opname saat sesi final, termasuk `coor-page`.
13. Build admin web dan cek manual.
14. Deploy backend + admin.

Perkiraan dampak: sekitar 40 file Go (termasuk test dan fake), 1 file schema, seed, API spec, plus admin web. Perubahan paling mekanis ada di test/fake; perubahan logika utama ada di repository, import, readiness, stock opname, completion guard, dan delete/cascade.

---

## 8. Acceptance criteria

- [ ] Dua sesi bisa menyimpan PLU dan barcode yang sama dengan produk berbeda.
- [ ] Scan barcode di rak sesi A selalu resolve ke produk sesi A.
- [ ] List, import, dan clear hanya menyentuh satu sesi.
- [ ] `stock_opname` lama tetap terbaca setelah master dihapus/di-clear.
- [ ] `stock_opname` tetap snapshot; tidak ada FK baru ke product; hasil ikut terhapus hanya saat sesi dihapus (D9).
- [ ] `InspectorJoinResponse` mengembalikan `sesi_id`.
- [ ] `sessionId` hilang/invalid -> 400; sesi tidak ada -> 404; duplikat/sesi terkunci -> 409.
- [ ] `PATCH/DELETE /products/:id` tanpa `sessionId` -> 400; produk beda sesi -> 404.
- [ ] `DELETE /products` dan `DELETE /products/clear` tidak ada; `POST /products/clear?sessionId=` adalah satu-satunya path clear.
- [ ] Import dengan barcode duplikat terhadap data existing melaporkan `invalid_rows` dan tidak menggagalkan seluruh import.
- [ ] Barcode yang pindah PLU dalam satu import berhasil lewat dua fase (D4-A).
- [ ] Re-import dengan set barcode yang sama tidak menghapus/meng-insert ulang barcode; `barcode_id` dan primary barcode stabil.
- [ ] Multi-pengklaim dengan owner melepas: B/C invalid, pelepasan A tetap berlaku (D4-A).
- [ ] Barcode baru tanpa owner dengan satu pengklaim diterima; dengan beberapa pengklaim semua invalid.
- [ ] Duplikat barcode antar-PLU mengikuti aturan pemilik DB dan hasilnya tidak bergantung urutan file.
- [ ] Delete produk langsung menghapus baris; PLU yang sama bisa dibuat/di-import lagi.
- [ ] Sesi `COMPLETED`/`CANCELLED` full read-only: create/update/delete/import/clear produk, create rak, dan create/update/delete stock opname ditolak 409 (D3).
- [ ] Join inspector pada sesi final ditolak 409; inspector/rak tidak bertambah.
- [ ] `POST /stockopname/racks` pada sesi final ditolak 409; tidak ada rak baru.
- [ ] `PATCH /stockopname/coordinators/:id` pada sesi final ditolak 409; status coordinator tidak berubah.
- [ ] `PATCH /stockopname/sessions/:id` pada sesi final ditolak 409; tidak ada reopen; delete sesi tetap 200 (D9).
- [ ] Pengecekan status D3 terjadi di transaksi yang sama dan memakai lock `FOR SHARE`; real two-connection lock test lulus.
- [ ] Constraint `fk_product_sesi` ada dengan nama eksplisit dan dipetakan di `constraintMessages`.
- [ ] `GET` produk dan `GET` hasil stock opname tetap 200 pada sesi final.
- [ ] Update ke `COMPLETED` saat produk 0 ditolak 409; tombol selesai disabled.
- [ ] `DELETE /stockopname/sessions/:id`: `COMPLETED`/`CANCELLED` -> 200 purge; `IN_PROGRESS` kosong -> 200; `IN_PROGRESS` berisi -> 409.
- [ ] Delete sesi menghapus product, barcode, stock_opname, rak, inspector, dan coordinator tanpa orphan (cascade).
- [ ] `dryRun=true` pada import tidak mengubah barcode (D4-A).
- [ ] Sesi dengan 0 produk: join inspector -> 409; `POST /results` dan `POST /results/racks` -> 409.
- [ ] Import parsial dengan minimal satu produk valid membuat sesi siap.
- [ ] Tambah produk manual membuat sesi siap.
- [ ] Clear/delete semua produk membuat sesi kembali belum siap.
- [ ] Admin web tidak punya route `/master` dan menu Master Data.
- [ ] Admin web punya section produk di detail sesi dengan CRUD + import + clear.
- [ ] Wizard create sesi 2 langkah; import wajib; gagal import -> sesi tetap ada dan bisa retry/manual/delete kosong.
- [ ] Admin web mengirim `sessionId`/`sesi_id` di semua operasi produk, termasuk PATCH/DELETE.
- [ ] Admin web memakai `getProductCount` (`pagination.total_items`) dan me-refresh count setelah create/update/delete/import/clear.
- [ ] Admin web menonaktifkan tombol selesai saat `productCount === 0` dan semua aksi write stock opname saat sesi final.
- [ ] `coor-page` juga read-only saat sesi final (aksi edit/delete hasil dan tambah item tidak tampil/disabled).
- [ ] Admin web tidak menampilkan edit/status/reopen untuk sesi final; hanya hapus.
- [ ] Admin web punya tombol hapus sesi dengan konfirmasi ketik kode sesi (`requireText`).
- [ ] `DELETE /stockopname/sessions/:id` sukses mengembalikan `200` JSON (bukan 204) sehingga `api.ts` tidak gagal parse.
- [ ] Client mobile tidak diimplementasikan di plan ini; kontrak backend untuk client berikutnya terdokumentasi.
- [ ] Tidak ada endpoint `/products/sync`, `/deleted/products`, atau tabel `deleted_product`.
- [ ] Tidak ada lagi referensi `last-session` di kode, test, atau spec.
- [ ] `go build ./...`, `go test ./...`, `go vet ./...` lulus.
- [ ] `docs/apispec.json` sesuai dengan kontrak baru; mobile `apispec.json` tidak diubah.
- [ ] QA docs tidak lagi mengklaim Dashboard, Master Data, `Total Produk`, `last-session`, tombstone, atau deleted product.
- [ ] Fresh install dari DB kosong berhasil: `schema1.sql` + `seed.sql` menghasilkan 36 produk dan 57 barcode.
- [ ] Rollout backend + admin; mobile didesain ulang terpisah.

---

## 9. Risiko dan blocker

1. **Breaking change client contract** — endpoint produk mewajibkan `sessionId`, termasuk PATCH/DELETE. Mobile lama tidak kompatibel dan akan didesain ulang terpisah. Mitigasi: rollout backend + admin saja; jangan deploy backend ke lapangan jika mobile lama masih menjadi satu-satunya client produk.
2. **Sesi `IN_PROGRESS` tanpa produk** — tidak ada status `DRAFT`; sesi tetap muncul di daftar walau belum siap. Mitigasi: readiness check memblokir join/scan; UI detail menampilkan warning; admin retry import/manual atau delete sesi kosong.
3. **Wizard tidak atomic** — create sesi dan import dua endpoint. Kalau import gagal, sesi tetap ada. Mitigasi: redirect ke detail, retry import, manual add, atau delete sesi kosong (D9). Kalau ingin atomic, perlu endpoint gabungan/transaksi baru; tidak dipilih.
4. **D3 lock** — `FOR SHARE` menahan update status/delete sesi sampai transaksi tulis selesai. Mitigasi: lock sesi dulu, lalu produk/hasil; transaksi pendek; real two-connection lock test dan smoke. Completion/delete memakai `FOR UPDATE`, bukan `FOR SHARE`.
5. **D4 multi-pengklaim** — owner A melepas, B/C invalid, barcode bisa hilang dari semua produk. Ini keputusan final; UI/import result harus menampilkan `invalid_rows` agar admin sadar.
6. **Global unique lama** — kalau `product_plu`/`barcode_code` masih `UNIQUE`, PLU/barcode yang sama di sesi berbeda gagal. Mitigasi: fresh schema tanpa global unique; smoke dua sesi.
7. **`barcode_sesi_id` denormalisasi** — bisa salah isi kalau insert manual. Mitigasi: composite FK `fk_barcode_sesi_product`; jangan hapus constraint ini.
8. **Clear semua produk** — sesi langsung belum siap; join/scan diblokir. Mitigasi: warning di `ProductClearCard` dan empty state.
9. **PATCH/DELETE ownership** — kalau `sessionId` tidak divalidasi, produk sesi lain bisa diubah. Mitigasi: wajib query param, mismatch 404, test.
10. **Readiness query tambahan** — satu `COUNT(*)` per join/scan. Mitigasi: index `(product_sesi_id, product_plu)` sudah ada; transaksi pendek.
11. **Count tidak refresh setelah mutasi** — kalau `productCount` hanya di-fetch saat mount, tombol "Selesaikan Sesi" tetap disabled setelah produk pertama ditambahkan. Mitigasi: `getProductCount` + prop `onProductsChanged` dari `ProductTabs`.
12. **`coor-page` ikut punya aksi tulis** — halaman koordinator punya tabel hasil sendiri; kalau tidak di-gate, tombolnya 409 saat sesi final. Mitigasi: `readOnly` dari `sessionData.status`; hide/disable aksi hasil.
13. **Import conflict vs existing DB** — unique violation membatalkan transaksi. Mitigasi: pre-validasi; race -> 409 jelas.
14. **Error mapping** — FK error dipetakan 409, bukan 404. Mitigasi: validasi sesi eksplisit lewat `SesiRepository`; readiness memakai `ConflictError`.
15. **Index/search** — filter sesi + `ILIKE` bisa mahal. Mitigasi: `EXPLAIN`, lalu tambah composite index kalau perlu.
16. **Admin tanpa katalog global** — tidak ada halaman lintas sesi. Mitigasi: laporan/export tetap memakai snapshot; kalau audit lintas sesi dibutuhkan, tambahkan read-only report terpisah.
17. **Reset DB destruktif** — fresh install menghapus semua data. Mitigasi: backup manual jika perlu; pastikan DB dev/test.
18. **Tidak ada audit delete produk** — tanpa tombstone, riwayat delete produk hanya ada di snapshot `stock_opname` jika pernah discan. Mitigasi: audit log terpisah kalau nanti dibutuhkan.
19. **Delete sesi = hard purge** — `stock_opname`, rak, inspector, coordinator, produk, dan barcode ikut terhapus; tidak bisa di-undo. Mitigasi: hanya `COMPLETED`/`CANCELLED` atau `IN_PROGRESS` kosong; konfirmasi ketik kode sesi; export dulu kalau history dibutuhkan.
20. **Completion guard** — kalau count dan update tidak berada di lock yang sama, clear bersamaan bisa membuat sesi `COMPLETED` dengan 0 produk. Mitigasi: `FindByIdForUpdate` (`FOR UPDATE`) sebelum count/update; test race/smoke.
21. **Sesi final tidak bisa dikoreksi** — `COMPLETED`/`CANCELLED` full read-only; `PATCH` sesi final juga 409, jadi tidak ada reopen. Mitigasi: pastikan count benar sebelum menyelesaikan; kalau ada kesalahan, hapus sesi dan buat ulang (history hilang) atau buka kebijakan reopen terpisah di masa depan.
22. **Readiness count tidak sepenuhnya serial** — join/create dan clear produk sama-sama memakai `FOR SHARE`, jadi count bisa lolos tepat sebelum clear. Mitigasi: dokumentasikan eventual consistency; completion/delete memakai `FOR UPDATE`; kalau butuh strict, ubah write produk ke `FOR UPDATE` atau tambah row readiness (belum dipilih).
23. **`POST /stockopname/racks` tanpa guard** — endpoint ini membuat rak langsung dari `inspector_id`, jadi sesi final bisa tetap mendapat rak. Mitigasi: resolusi sesi lewat `InspectorRepository.FindSesiIdByInspector`, lock `FOR SHARE`, final -> 409; test service + route.
24. **Import phase 2 terlalu lebar** — menghapus/meng-insert ulang barcode semua produk yang di-import membuat `barcode_id` berubah dan primary barcode bisa bergeser. Mitigasi: bandingkan set lama/baru; hanya produk dengan set berbeda yang di-`Replace`/`SaveBatch`; test re-import dengan set sama.
25. **Delete sesi `IN_PROGRESS` yang sudah punya inspector/rak** — guard delete hanya menghitung produk dan hasil stock opname. Sesi aktif dengan 0 produk tetapi sudah ada inspector/rak tetap boleh dihapus (cleanup), dan seluruh child-nya ikut terhapus. Mitigasi: konfirmasi ketik kode sesi dan jelaskan cakupan purge; kalau ingin lebih ketat, tambahkan count inspector/rak ke guard (belum dipilih).
26. **Import bersamaan pada sesi yang sama** — dua import sama-sama memakai `FOR SHARE` dan bisa lolos pre-validasi; yang kalah kena unique constraint dan mengembalikan 409. Mitigasi: dokumentasikan; jangan naikkan ke `FOR UPDATE` tanpa mengukur dampaknya.

---

## 10. Keputusan final

### D1. `/products/last-session` — selesai

Diputuskan **A: hapus endpoint**. Implementasi sudah selesai: route, handler, service, repository method, DTO, model, fake, test, dan `docs/apispec.json` sudah dibersihkan. Dashboard admin juga dihapus; statistik sesi dipindah ke Stock Opname dan `Total Produk` tidak ditampilkan lagi.

Tidak ada pekerjaan lanjutan untuk kode/spec D1. QA docs yang masih menyebut `last-session`/Dashboard/Master Data dibersihkan pada langkah 8.

### D2. Migrasi data lama — selesai diputuskan

Diputuskan **fresh schema + seed, tanpa migrasi**. DB existing boleh dihapus. `schema2_session_master.sql` tidak dibuat; tidak ada backfill, sesi legacy, atau pemetaan data lama.

Konsekuensi: semua data lama akan hilang saat DB dibuat ulang. `stock_opname` lama juga tidak dipertahankan kecuali masih ada di dump/manual.

### D3. Sesi `COMPLETED`/`CANCELLED` final — selesai

Diputuskan **final/read-only untuk semua write**.

- Sesi `COMPLETED`/`CANCELLED`: create/update/delete/import/clear produk, create inspector/rak saat join, create rak (`POST /stockopname/racks`), update status coordinator, **dan** create/update/delete hasil stock opname ditolak `409`.
- `PATCH /stockopname/sessions/:id` pada sesi final juga ditolak `409`; tidak ada reopen. Satu-satunya write adalah delete sesi (D9).
- Read produk dan read hasil stock opname tetap boleh, termasuk list, pencarian, dan export.
- `PATCH/DELETE` by ID memuat produk setelah lock sesi dan memeriksa `ProductSesiID == sessionId`; mismatch -> `404`; setelah kepemilikan cocok, cek status -> `409`.
- `Clear` dan `Import` memeriksa status sesi dari `sesiId`.
- Pengecekan status produk memakai `SesiRepository.FindByIdForShare` (`FOR SHARE`) di transaksi yang sama.
- `InspectorService` memakai `FindByIdForShare` setelah coordinator ditemukan; sesi final -> 409.
- `RackService.CreateRack` memakai `FindSesiIdByInspector` lalu `FindByIdForShare`; sesi final -> 409.
- `CoordinatorService.Update` memakai `FindByIdForShare` pada sesi milik coordinator; sesi final -> 409.
- `StockOpnameService` memakai `FindSesiIdByRak`/`FindSesiIdByStockOpnameID` lalu `FindByIdForShare` sebelum write.
- UI admin menonaktifkan/menyembunyikan tombol tambah/edit/hapus/import/clear produk dan tambah/edit/hapus hasil stock opname.
- Client menampilkan `409` saat scan/write pada sesi final; jangan retry buta.
- Opsi “hanya master produk yang dikunci” tidak dipakai lagi.

### D4. Barcode pindah PLU dalam satu import — selesai

Diputuskan **A: dua fase**.

Alur import:

1. Parse dan validasi bentuk file; kumpulkan produk dan kandidat PLU per barcode.
2. Mulai transaksi, lalu lock sesi dan load produk sesi itu (termasuk owner barcode saat ini).
3. Tentukan valid/invalid barcode dan susun `invalid_rows` tanpa bergantung urutan file.
4. Upsert field produk saja; kumpulkan `product_id`, set barcode lama/baru, dan hitung `BarcodesAdded`/`BarcodesRemoved` dari selisih set.
5. Hapus barcode lama hanya untuk produk yang set barcode barunya berbeda dari set lama.
6. Insert barcode baru hanya untuk produk yang set barcode barunya berbeda.
7. Commit; kalau ada error, rollback.

Produk dengan set barcode sama tidak di-`Replace`, sehingga `barcode_id` dan urutan primary barcode tidak berubah saat re-import. Perbandingan memakai set, bukan urutan.

Aturan multi-pengklaim final:

- Duplicate PLU di file produk mempertahankan **last occurrence** dan menandai baris sebelumnya invalid (perilaku parser sekarang); kandidat D4 memakai PLU hasil akhir.
- Duplicate barcode untuk PLU yang sama di file barcode: baris pertama tetap masuk kandidat, baris duplikat setelahnya masuk `invalid_rows` (perilaku parser sekarang via `seenBarcode`).
- Barcode hanya boleh pindah ke PLU lain kalau PLU pemilik lama ikut di-import dan barcode itu tidak ada lagi di set barunya.
- Kalau PLU pemilik lama tidak ada di file, atau masih memakai barcode itu, baris barcode tujuan masuk `invalid_rows`.
- Barcode yang sama diberikan ke beberapa PLU dalam file tetap dianggap duplikat file, bukan perpindahan. Kalau salah satunya pemilik di DB dan pemilik masih memakai barcode itu, pertahankan baris pemilik dan tolak yang lain; kalau tidak ada pemilik, tolak semua baris kode itu.
- Kalau pemilik A melepas barcode dan B/C sama-sama mengklaim, B/C invalid dan pelepasan A tetap berlaku. Barcode tidak dimiliki siapa pun setelah import.
- Barcode baru tanpa owner dengan satu pengklaim diterima.
- Jangan bergantung urutan file. Kumpulkan semua kandidat dan owner sebelum menentukan pemenang.

Catatan implementasi: `dryRun` berhenti sebelum fase hapus/insert. Komentar `updatedat` yang masih menyebut client sync harus dibersihkan; `updatedat` tetap di-bump saat field atau set barcode berubah. Fase hapus/insert hanya menyentuh produk yang set barcode-nya berubah; produk dengan set sama tidak di-`Replace`, sehingga `barcode_id`/primary barcode stabil. Opsi B (`DEFERRABLE`) dan C (tolak semua perpindahan) tidak dipakai.

### D5. Admin flow — selesai

Diputuskan **halaman Master Data global dihapus; produk pindah ke dalam sesi dan wizard pembuatan sesi**.

- `/master`, `MasterPage`, dan menu sidebar dihapus.
- Detail sesi punya section "Master Produk" dengan CRUD + import + clear.
- Create sesi memakai wizard 2 langkah: sesi + koordinator, lalu import produk.
- Import adalah langkah wajib di wizard; tidak ada tombol lewati.
- Kalau import gagal, sesi tetap ada dan bisa retry/manual atau dihapus jika masih kosong.
- Produk masih boleh diedit selama `IN_PROGRESS`; `COMPLETED`/`CANCELLED` full read-only (D3).
- Sesi final tidak punya edit/status/reopen; satu-satunya aksi tulis adalah hapus sesi (D9).
- Tombol “Selesaikan Sesi” disabled saat produk 0; backend menolak `409`.
- Detail sesi punya tombol hapus untuk `COMPLETED`/`CANCELLED` dan untuk `IN_PROGRESS` kosong (D9).
- Tidak ada session selector global; session ID dari route.

### D6. Readiness — selesai

Diputuskan **sesi tanpa produk diblokir untuk join/scan baru**.

- Readiness = `COUNT(product) > 0` untuk sesi `IN_PROGRESS`.
- `POST /stockopname/inspectors` dan join QR -> 409 kalau 0.
- Join inspector pada sesi final -> 409 (D3); readiness hanya untuk sesi `IN_PROGRESS`.
- `POST /stockopname/results` dan `POST /stockopname/results/racks` -> 409 kalau 0.
- `GET` hasil lama tetap boleh; memakai snapshot.
- `PATCH/DELETE` hasil pada sesi final ditolak 409 (D3), bukan readiness.
- Import parsial dengan minimal satu produk valid -> siap.
- Tambah produk manual -> siap.
- Clear/delete semua produk -> belum siap lagi.
- `PATCH session -> COMPLETED` dengan produk 0 -> 409.
- Tidak ada status `DRAFT`; sesi tetap `IN_PROGRESS` dan UI harus menampilkan warning.
- Sesi `IN_PROGRESS` kosong bisa dihapus untuk cleanup (D9).
- Readiness count bersifat best-effort terhadap clear/delete produk yang berjalan bersamaan: keduanya memakai `FOR SHARE`, jadi join/create bisa lolos count tepat sebelum clear, lalu scan berikutnya 404/409. Ini diterima; completion dan delete sesi memakai `FOR UPDATE` sehingga tidak bisa menghasilkan sesi final kosong.

### D7. Kontrak endpoint produk — selesai

- `GET /products`, `POST /products/import`, `POST /products/clear`: wajib `?sessionId=`.
- `PATCH/DELETE /products/:id`: wajib `?sessionId=`; mismatch -> 404.
- `POST /products`: body `sesi_id`.
- Clear hanya `POST /products/clear?sessionId=&confirm=true`; `DELETE /products` dan `DELETE /products/clear` tidak didaftarkan.
- `sessionId` hilang/invalid -> 400; sesi tidak ada -> 404; duplikat/terkunci/readiness -> 409.
- Import invalid rows -> 200 dengan `invalid_rows`; race/constraint tak terduga -> 409.

### D8. Rollout — selesai

Diputuskan **backend + admin saja; mobile out of scope**.

- Backend dan admin web dirilis bersama.
- Mobile `apispec.json` tidak diubah di plan ini; client mobile akan didesain ulang terpisah.
- Mobile lama tidak kompatibel untuk list produk (`sessionId` wajib). Jangan deploy backend ke lapangan jika mobile lama masih menjadi satu-satunya client produk.
- Kontrak backend untuk client berikutnya ada di §3.2.

### D9. Hapus sesi + completion guard — selesai

Diputuskan **sesi selesai/batal boleh dihapus permanen; sesi aktif hanya boleh dihapus jika kosong; sesi kosong tidak boleh diselesaikan**.

- `PATCH /stockopname/sessions/:id` ke `COMPLETED` ditolak `409` jika `COUNT(product) = 0`.
- Completion guard memakai `FindByIdForUpdate` (`FOR UPDATE`) sebelum count/update.
- `DELETE /stockopname/sessions/:id`:
  - `COMPLETED`/`CANCELLED` -> 200, hard purge;
  - `IN_PROGRESS` dengan 0 produk dan 0 hasil -> 200, cleanup;
  - `IN_PROGRESS` yang masih punya produk/hasil -> 409.
- Sesi `IN_PROGRESS` yang masih punya hasil tetapi produknya sudah 0 harus di-`CANCELLED` dulu, lalu dihapus.
- Delete memakai `FindByIdForUpdate` (`FOR UPDATE`).
- `PATCH /stockopname/sessions/:id` pada sesi final ditolak `409`; delete adalah satu-satunya write yang diizinkan, jadi tidak ada jalur reopen lewat update status.
- FK anak sesi dan `stock_opname` memakai `ON DELETE CASCADE`; satu `DELETE FROM sesi` menghapus semua child.
- `stock_opname` sesi tersebut ikut terhapus. Ini hard delete, bukan arsip.
- UI memakai konfirmasi ketik kode sesi; tombol delete tidak ada di list.
- Kalau history wajib disimpan, export dulu atau desain soft delete/arsip terpisah; hard delete tidak bisa di-undo.

---

## 11. Catatan keamanan

Tidak ada autentikasi. `sessionId` bisa diganti siapa saja yang bisa memanggil API. Per-sesian di sini mencegah tabrakan data dan kesalahan operasional, bukan melindungi data antar-tempat. Kalau backend dipakai bersama lewat jaringan, autentikasi/otorisasi perlu ditambahkan terpisah. Ini terutama penting untuk `DELETE /stockopname/sessions/:id` dan update status, karena keduanya destruktif/final.
