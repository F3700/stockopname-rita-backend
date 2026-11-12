# Stock Opname RITA — Backend

API untuk aplikasi stock opname RITA: sesi hitung, koordinator dan rak, master produk
(PLU + barcode) per sesi, hasil hitung beserta snapshot harganya, riwayat perubahan,
serta ekspor laporan PDF/Excel/DBF.

Ditulis dengan Go, disajikan lewat `net/http` + `julienschmidt/httprouter`, dan
menyimpan seluruh data di PostgreSQL. Frontend web dan aplikasi mobile memakai API
yang sama.

---

## Daftar isi

1. [Arsitektur](#arsitektur)
2. [Prasyarat](#prasyarat)
3. [Menjalankan di lokal](#menjalankan-di-lokal)
4. [Konfigurasi environment](#konfigurasi-environment)
5. [Database](#database)
6. [Migrasi](#migrasi)
7. [Data contoh](#data-contoh)
8. [Autentikasi](#autentikasi)
9. [Daftar endpoint](#daftar-endpoint)
10. [Perilaku yang perlu diketahui](#perilaku-yang-perlu-diketahui)
11. [Pengujian](#pengujian)
12. [Log dan pemantauan](#log-dan-pemantauan)
13. [Build dan deploy](#build-dan-deploy)
14. [Pemecahan masalah](#pemecahan-masalah)
15. [Batasan yang perlu diketahui](#batasan-yang-perlu-diketahui)

---

## Arsitektur

```
cmd/server            entrypoint: env, pool, middleware, graceful shutdown
cmd/hashpw            pembuat hash password untuk AUTH_PASSWORD_HASH
internal/
├── router/           perakitan seluruh route + dependency (DI manual)
├── middleware/       access log, recovery, CORS, auth, body limit
├── handler/          lapisan HTTP: baca request, panggil service, tulis response
├── service/          aturan bisnis dan transaksi; satu transaksi per operasi tulis
├── repository/       query SQL; tidak tahu HTTP, tidak tahu aturan bisnis
├── model/            entitas domain + tipe error (NotFound, Conflict, Validation, …)
├── dto/              bentuk request/response JSON
├── auth/             sesi cookie, hash password, daftar putih route, rate limit
├── helper/           utilitas bersama: response JSON, parsing, aktor audit
└── database/         pembuatan pgxpool
migrations/           SQL skema, reset, dan data contoh
docs/                 apispec.json (Swagger) yang ditanam ke binary
```

Arah ketergantungan selalu satu arah: `router → handler → service → repository`.
Handler tidak menulis SQL; repository tidak memutuskan status HTTP.

Alur satu request:

```
Request → AccessLog → Recovery → CORS → RequireAuth → LimitBody → httprouter
        → handler → service → repository → PostgreSQL
```

| Middleware | Tugas |
|---|---|
| `AccessLog` | satu baris log per request + header `X-Request-Id` (probe `/health` dilewati) |
| `Recovery` | mengubah panic menjadi 500 dan tetap mencatat sebagai 500 |
| `CORS` | allowlist origin (bukan `*`), karena cookie tidak sah bersama wildcard |
| `RequireAuth` | menolak request yang bukan publik tanpa cookie sesi yang sah |
| `LimitBody` | membatasi ukuran body permintaan |

## Prasyarat

- **Go 1.26** (lihat `go.mod`; CI dan image memakai versi yang sama)
- **PostgreSQL 15+** (dipakai `GENERATED ALWAYS AS IDENTITY`, `JSONB`, dan `TIMESTAMPTZ`)
- **Podman** untuk menjalankan PostgreSQL di lokal (opsional bila sudah punya server DB)
- **psql** untuk menjalankan migrasi

## Menjalankan di lokal

### 1. Siapkan PostgreSQL

```bash
podman run -d --name postgres-rita \
  -e POSTGRES_USER=rita -e POSTGRES_PASSWORD=rita \
  -e POSTGRES_DB=stockopname -p 5432:5432 \
  docker.io/library/postgres:latest
```

### 2. Siapkan `.env`

```bash
cp .env.example .env      # Windows PowerShell: Copy-Item .env.example .env
```

Isi minimal:

```dotenv
APP_HOST=127.0.0.1
APP_PORT=8080
DATABASE_URL=postgres://rita:rita@127.0.0.1:5432/stockopname?sslmode=disable
AUTH_USERNAME=admin
AUTH_PASSWORD_HASH=pbkdf2-sha256:210000:...:...
SESSION_SECRET=<32+ karakter acak>
COOKIE_SECURE=false
CORS_ALLOWED_ORIGINS=https://rita.albertt.my.id
```

Buat hash password:

```bash
go run ./cmd/hashpw          # ketik password di stdin, salin keluarannya
```

Buat secret:

```bash
openssl rand -hex 32
```

> **Pemisah hash adalah titik dua, bukan `$`.** Nilai `AUTH_PASSWORD_HASH` disimpan di
> `.env`, dan `godotenv` mengekspansi apa pun yang menyerupai `$NAMA`. Dengan `$`
> sebagai pemisah, angka iterasi dan kedua blob ikut dimakan sehingga hash tidak
> terbaca dan server menolak start. Format yang benar:
> `pbkdf2-sha256:210000:<salt>:<key>`.

### 3. Siapkan skema dan data

```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_schema.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/002_stock_opname_audit.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/003_stock_opname_department.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/004_stock_opname_sell_factor.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/005_stock_opname_audit_session_fk.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/006_product_per_carton.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/007_stock_opname_qty_factor.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/008_stock_opname_upload_ownership.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/seed.sql
```

### 4. Jalankan

```bash
go run ./cmd/server
```

```
Server is running on 127.0.0.1:8080
```

Uji cepat:

```bash
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1:8080/health/ready
curl -s http://127.0.0.1:8080/auth/me
```

## Konfigurasi environment

Dibaca saat start; `.env` dimuat lewat `godotenv` bila ada. Variabel tanpa nilai
default **wajib** diisi — server menolak start daripada berjalan setengah benar.

| Variabel | Wajib | Default | Keterangan |
|---|---|---|---|
| `APP_HOST` | ya | — | Alamat bind, mis. `0.0.0.0` di container |
| `APP_PORT` | ya | — | Port HTTP, mis. `8080` |
| `DATABASE_URL` | ya | — | DSN PostgreSQL untuk `pgxpool` |
| `AUTH_USERNAME` | ya | — | Username login admin |
| `AUTH_PASSWORD_HASH` | ya | — | Hash `pbkdf2-sha256:<iterasi>:<salt>:<key>` |
| `SESSION_SECRET` | ya | — | Kunci HMAC cookie, minimal 32 byte |
| `SESSION_TTL` | tidak | `12h` | Masa berlaku sesi |
| `COOKIE_SECURE` | tidak | `true` | `false` hanya untuk dev lewat http |
| `COOKIE_DOMAIN` | tidak | kosong | Kosong = host-only (dipakai saat proxy same-origin) |
| `CORS_ALLOWED_ORIGINS` | tidak | `https://rita.albertt.my.id` | Daftar origin dipisah koma |
| `HTTP_MAX_BODY_BYTES` | tidak | `67108864` (64 MiB) | Batas body request |
| `HTTP_WRITE_TIMEOUT` | tidak | `3m` | Waktu tulis response |
| `REPORT_MAX_ROWS` | tidak | `20000` | Batas baris satu ekspor |
| `REPORT_MAX_CONCURRENT` | tidak | `2` | Ekspor bersamaan; `0` mematikan batas |

Nilai yang tidak valid tidak membuat server mati; ia diabaikan dan nilai default
dipakai dengan satu baris peringatan di log. Pengecualian: kredensial dan
`SESSION_SECRET` — kalau salah, server berhenti.

## Database

### Ringkasan tabel

```
sesi ──┬── product ──── barcode            (master per sesi)
       ├── coordinator ──┬── upload        (satu kiriman hasil per koordinator)
       │                 └── rak ──────────┐
       └── stock_opname_audit              │
                                           └── stock_opname (SNAPSHOT, bukan FK ke product)
```

| Tabel | Isi | Kunci |
|---|---|---|
| `sesi` | satu kegiatan stock opname: kode unik, lokasi, status, waktu mulai/selesai | `sesi_code` unik; status `IN_PROGRESS`/`COMPLETED`/`CANCELLED` dengan aturan waktu yang dijaga `CHECK` |
| `product` | master produk milik **satu sesi**: PLU, nama, departemen, harga beli/jual, isi karton | unik `(product_sesi_id, product_plu)`; `ON DELETE CASCADE` dari sesi |
| `barcode` | 0..N barcode per produk; juga unik per sesi | `UNIQUE (barcode_sesi_id, barcode_code)`; FK komposit ke `product` |
| `coordinator` | penanggung jawab sekelompok rak dalam sesi | unik `(coor_sesi_id, coor_code)` |
| `upload` | penanda satu kiriman hasil dari aplikasi mobile | milik satu koordinator |
| `rak` | unit lokasi hitung; `rak_upload_id` adalah *host* (upload terakhir yang mengklaim rak) | unik `(rak_coor_id, rak_name)` |
| `stock_opname` | baris hasil hitung: jumlah + **snapshot** produk | unik `(stock_opname_rak_id, so_product_plu)` |
| `stock_opname_audit` | riwayat perubahan baris hasil | `batch_id` mengelompokkan satu aksi massal |

### Kenapa `stock_opname` menyimpan snapshot

`stock_opname` tidak menyimpan FK ke `product`. Sebagai gantinya, saat baris dibuat, nilai
produk disalin ke kolom `so_*`:

| Kolom snapshot | Asal |
|---|---|
| `so_product_plu`, `so_product_name` | master produk saat baris dibuat |
| `so_barcode` | barcode yang dipakai (boleh kosong) |
| `so_buyprice`, `so_sellprice` | harga saat baris dibuat |
| `so_product_dept` | departemen saat baris dibuat |
| `so_qty_factor` | faktor kuantitas (`-1`, `0`, `1`) |
| `so_upload_id` | upload pemilik baris |

Konsekuensinya laporan lama tetap benar walaupun master produk diperbarui, di-*clear*,
atau dihapus. Itu sebabnya tombol "hapus semua produk" tidak mengubah riwayat.

### Trigger dan fungsi

| Objek | Perilaku |
|---|---|
| `trg_sync_sesi_endedat` | `sesi_endedat` diisi `NOW()` saat status keluar dari `IN_PROGRESS`, dan dikosongkan bila kembali |
| `update_sesi_status(id, status)` | mengubah status sesi sekaligus menutup koordinator yang masih berjalan |
| `recalc_coordinator_status(id)` | menaikkan koordinator ke `IN_REVIEW` ketika semua raknya sudah punya minimal satu baris |
| `trg_sync_coordinator_status` | memanggil fungsi di atas setiap ada baris `stock_opname` baru |

### Indeks

- `idx_product_name`, `idx_product_dept` — pencarian dan pengelompokan produk.
- `idx_barcode_product` — mengambil barcode per produk.
- `idx_upload_coor`, `idx_rak_upload` — relasi upload.
- `idx_so_audit_batch` — filter riwayat per batch.
- `idx_stock_opname_upload` — mengambil baris per upload.

Pencarian nama memakai `ILIKE`. Untuk data puluhan ribu baris ini masih wajar; kalau
nanti melambat, langkah berikutnya adalah indeks trigram (`pg_trgm`).

### Pool koneksi

`database.NewPostgresPool` memakai `pgxpool.New` dengan `DATABASE_URL` apa adanya dan
melakukan `Ping` saat start. Bila database tidak terjangkau, server berhenti dengan
pesan jelas, bukan melayani request yang pasti gagal.

## Migrasi

Migrasi berupa berkas SQL bernomor di `migrations/`, dijalankan **manual dan berurutan**.
Tidak ada tabel pelacak versi, jadi status skema harus diperiksa sendiri.

| Berkas | Isi |
|---|---|
| `001_schema.sql` | seluruh tabel awal + trigger dan fungsi status |
| `002_stock_opname_audit.sql` | tabel audit (`before_data`/`after_data` JSONB, `batch_id`) |
| `003_stock_opname_department.sql` | kolom `so_product_dept` + backfill dari master |
| `004_stock_opname_sell_factor.sql` | kolom faktor penjualan + indeks batch audit |
| `005_stock_opname_audit_session_fk.sql` | FK audit ke sesi (`ON DELETE CASCADE`), bersihkan audit yatim |
| `006_product_per_carton.sql` | `product_per_carton` untuk produk |
| `007_stock_opname_qty_factor.sql` | rename `so_sell_factor` → `so_qty_factor`, reset nilai ke `1` |
| `008_stock_opname_upload_ownership.sql` | `so_upload_id` + backfill kepemilikan baris |
| `reset.sql` | **destruktif**: `DROP SCHEMA public CASCADE`, untuk dev saja |
| `seed.sql` | data contoh: 4 sesi, 12 produk × 3 sesi, koordinator, rak, hasil |

Memeriksa skema sudah sampai mana:

```sql
SELECT column_name
FROM information_schema.columns
WHERE table_name = 'stock_opname'
  AND column_name IN ('so_upload_id', 'so_qty_factor', 'so_sell_factor');
```

- `so_upload_id` belum ada → `008` belum masuk.
- `so_qty_factor` belum ada atau `so_sell_factor` masih ada → `007` belum masuk.

Urutan 007 → 008 tidak boleh dibalik: 007 me-rename kolom, 008 menambah kolom baru
dan mem-backfill dari data yang sudah ada.

> **Pipeline rilis belum menjalankan migrasi.** Image hanya berisi binary. Setiap
> perubahan skema harus diterapkan ke database produksi **sebelum** image baru naik.

## Data contoh

`migrations/seed.sql` mengisi data development (bukan dari DBF legacy):

- 4 sesi dengan variasi status: 2 `COMPLETED`, 1 `IN_PROGRESS`, 1 `CANCELLED`.
- 12 produk untuk masing-masing dari 3 sesi pertama (sesi `CANCELLED` sengaja kosong).
- Koordinator, rak, hasil hitung, upload, dan beberapa baris audit.

Berkas `PRODUK.DBF` dan `BARCODE.DBF` **bukan** bagian dari seed; keduanya fixture
untuk menguji fitur impor.

## Autentikasi

Satu kredensial bersama untuk seluruh staf gudang. Setelah login, pengguna mengetik
nama operator yang dipakai sebagai aktor audit.

- Token sesi: cookie `rita_session`, isinya `base64url(claims) + "." + base64url(HMAC-SHA256)`.
  `claims = {op, iat, exp}`.
- `op` kosong sampai nama operator diisi; route admin menolak sesi tanpa nama.
- Semua kegagalan verifikasi dijawab satu jenis error (`ErrInvalidSession`) agar tidak
  membocorkan apakah token kedaluwarsa atau dipalsukan.
- Cookie: `HttpOnly`, `SameSite=Lax`, `Path=/`, `Max-Age` sesuai `SESSION_TTL`.
  `SameSite=Lax` + body JSON (tanpa form HTML) sudah menutup CSRF untuk mutasi.
- Percobaan login dibatasi 10 kegagalan per 5 menit per IP (in-memory, dari `X-Real-IP`).

| Endpoint | Publik | Keterangan |
|---|---|---|
| `POST /auth/login` | ya | 401 bila salah, 429 bila terlalu sering |
| `POST /auth/operator` | tidak (sesi saja) | menyimpan nama ke cookie baru |
| `POST /auth/logout` | tidak (sesi saja) | menghapus cookie |
| `GET /auth/me` | ya | selalu 200; `authenticated` dan `operator` |

Endpoint publik lain (dipakai aplikasi mobile dan probe): `GET /health`,
`GET /health/ready`, `POST /stockopname/coordinators/join-by-qr`,
`GET /stockopname/coordinators/{id}`, `GET /stockopname/sessions/{id}`,
`GET /products`, `GET /stockopname/racks`, `POST /stockopname/uploads`, dan `/docs`.

> Daftar putih itu sengaja: aplikasi mobile belum punya login. Selama itu berlaku,
> endpoint tersebut masih menerima header `X-Actor` yang bisa dipalsukan. Setelah
> login, aktor audit diambil dari cookie dan header diabaikan.

## Daftar endpoint

Dokumentasi lengkap dan selalu terbaru: `GET /docs` (Swagger UI) atau
`GET /docs/apispec.json`. Ringkasnya:

| Grup | Endpoint |
|---|---|
| Health | `GET /health`, `GET /health/ready` |
| Auth | `POST /auth/login`, `POST /auth/operator`, `POST /auth/logout`, `GET /auth/me` |
| Produk | `GET/POST /products`, `PATCH/DELETE /products/{id}`, `POST /products/import`, `POST /products/clear` |
| Sesi | `GET/POST /stockopname/sessions`, `GET/PATCH/DELETE /stockopname/sessions/{id}`, `.../qty-factor`, `.../sell-summary`, `.../pdf`, `.../excel`, `.../dbf` |
| Koordinator | `GET/POST /stockopname/coordinators`, `GET/PATCH /stockopname/coordinators/{id}`, `.../pdf`, `.../uploads/pdf`, `POST .../join-by-qr` |
| Rak | `GET /stockopname/racks`, `GET /stockopname/racks/progress`, `DELETE /stockopname/racks/{id}` |
| Upload | `GET/POST /stockopname/uploads`, `GET /stockopname/uploads/{id}/pdf` |
| Hasil SO | `GET/POST /stockopname/results`, `GET/PATCH/DELETE /stockopname/results/{id}`, `POST /stockopname/results/move` |
| Audit | `GET /stockopname/audits`, `.../batches`, `.../timeline` |

Semua response memakai bentuk yang sama:

```json
{ "message": "…", "data": … }
```

Daftar berpaginasi menambahkan `"pagination": {page, limit, total_pages, total_items}`.
Error memakai `message` yang sama dengan `data: null`, dan kode status sesuai jenis
error (`model.NotFoundError` → 404, `model.ConflictError` → 409, `model.ValidationError`
→ 400, `model.TooLargeError` → 413, `model.BusyError` → 503).

## Perilaku yang perlu diketahui

- **Impor master bersifat sinkron.** `POST /products/import` membaca kedua berkas
  sekaligus di memori, membandingkan dengan isi database, lalu menulis dalam satu
  transaksi. Puluhan ribu baris berarti puluhan detik sampai beberapa menit. Pastikan
  timeout di klien, nginx (`proxy_read_timeout`), dan `HTTP_WRITE_TIMEOUT` saling
  cocok.
- **Ekspor dibatasi.** `REPORT_MAX_ROWS` membatasi jumlah baris per laporan dan
  `REPORT_MAX_CONCURRENT` membatasi berapa laporan boleh dibuat serentak (lebih dari
  itu dijawab 503, bukan diantrikan). Laporan dirender sepenuhnya di memori.
- **Body request dibatasi** `HTTP_MAX_BODY_BYTES`. Sepasang DBF 250 ribu produk
  berukuran ±31 MiB, jadi masih lolos; jauh di atas itu akan dijawab 413.
- **Pengecekan health terpisah.** `/health` menjawab selama proses hidup; `/health/ready`
  ikut memeriksa database. Deploy sebaiknya memakai yang kedua.
- **Format waktu** dikirim sebagai RFC 3339 (`TIMESTAMPTZ`), jadi klien yang
  menampilkan jam bertanggung jawab atas zona waktunya.

## Pengujian

```bash
go test ./...            # seluruh paket
go test ./... -count=1   # paksa jalan ulang, tanpa cache
go vet ./...
gofmt -l .               # CI gagal bila ada berkas yang belum diformat
```

- Test ditulis *table-driven* dan tidak membutuhkan database: repository diuji dengan
  `pgxmock`, service dengan repository palsu, handler dengan service palsu.
- Test pada `internal/service` juga membuat berkas DBF asli di memori (lewat
  `go-dbase`) untuk menguji jalur impor dan ekspor — jadi tidak perlu fixture biner.
- CI (`.github/workflows/ci.yml`) menjalankan pemeriksaan format, `go vet`, dan
  `go test ./... -count=1`.

## Log dan pemantauan

Satu baris per request, ditulis middleware `AccessLog`:

```
request id=6f1c… method=GET uri=/stockopname/sessions?page=1 status=200 dur=12.40ms bytes=1832 remote=10.0.0.5:52311
```

- ID request bisa ditentukan klien lewat header `X-Request-Id` dan selalu dikembalikan
  pada response — berguna untuk mencocokkan laporan pengguna dengan log.
- Probe `/health` tidak dicatat agar tidak menenggelamkan baris lain.
- Error tak terduga dicatat oleh `helper.WriteError` beserta pesan aslinya; pesan ke
  klien tetap generik.

## Build dan deploy

```bash
go build -o bin/server ./cmd/server
```

Image (`Containerfile`): stage `golang:1.26.5` membangun binary statis
(`CGO_ENABLED=0`), lalu disalin ke `gcr.io/distroless/static-debian12` bersama
`docs/` (karena `apispec.json` ditanam lewat `//go:embed`).

Alur rilis:

1. Push ke `master` memicu `.github/workflows/release.yml`.
2. Job `test` menjalankan CI, lalu image dibangun dan didorong ke
   `ghcr.io/<owner>/<repo>` (tag `sha-xxxxxxx` dan `latest`).
3. VPS menjalankan `deploy-backend.sh <tag>` lewat SSH.

Di VPS, nginx memproxy `rita.albertt.my.id/api/` ke container ini dengan membuang
prefix `/api`. **Migrasi tidak dijalankan otomatis** oleh pipeline; terapkan manual
sebelum deploy.

## Pemecahan masalah

| Gejala | Penyebab yang paling sering |
|---|---|
| Server langsung mati saat start | `AUTH_USERNAME`, `AUTH_PASSWORD_HASH`, atau `SESSION_SECRET` kosong/terlalu pendek |
| `AUTH_PASSWORD_HASH is unusable` | hash ditulis dengan pemisah `$`, atau terpotong saat ditempel |
| `Failed to connect to the database` | `DATABASE_URL` salah, container DB mati, atau host tidak terjangkau |
| Semua endpoint admin menjawab 401 | cookie tidak terkirim (host/`SameSite`), atau nama operator belum diisi |
| Login menjawab 429 | batas 10 kegagalan per 5 menit per IP; tunggu atau periksa IP di belakang proxy |
| `401` pada endpoint yang seharusnya publik | route tidak ada di daftar putih `internal/auth/policy.go`, atau method-nya salah |
| Impor balik 413 | body melebihi `HTTP_MAX_BODY_BYTES` atau `client_max_body_size` nginx |
| Impor berhenti di tengah | timeout nginx (180s) atau `HTTP_WRITE_TIMEOUT` lebih pendek daripada waktu impor |
| Ekspor menjawab 503 | `REPORT_MAX_CONCURRENT` terlampaui; coba lagi setelah laporan lain selesai |
| Cascade menghapus data tak terduga | menghapus sesi ikut menghapus produk, koordinator, rak, hasil, dan audit — memang disengaja |

## Batasan yang perlu diketahui

- **Nama operator bukan bukti identitas.** Kredensial dipakai bersama dan nama
  diketik sendiri. Audit berguna menelusuri kebiasaan, bukan membuktikan pelaku.
- **Mobile belum punya login.** Enam endpoint tetap terbuka dan masih menerima
  `X-Actor` yang bisa dipalsukan sampai token join dikerjakan.
- **Rate limit login disimpan di memori.** Restart container mengosongkannya, dan
  beberapa replika tidak berbagi hitungan.
- **Tidak ada pelacak versi migrasi.** Urutan penerapan harus dijaga manusia.
- **Laporan dirender di memori.** Satu laporan besar menahan memori sebesar hasil
  akhirnya; itulah alasan ada batas baris dan batas konkurensi.
