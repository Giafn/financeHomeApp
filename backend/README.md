# Family Finance API

Backend aplikasi keuangan keluarga — Go + Fiber + GORM + PostgreSQL + JWT, dengan Clean Architecture.
Skeleton ini adalah lanjutan dari dokumen spesifikasi produk (`specs-aplikasi-keuangan-keluarga.md`).

## Struktur Folder (Clean Architecture)

```
cmd/
  api/main.go        -> entrypoint aplikasi (HTTP server), TIDAK menjalankan migrasi
  migrate/main.go     -> binary migrasi TERPISAH, dijalankan manual

internal/
  config/              -> load & validasi environment variable
  entity/              -> model domain (GORM struct) — semua tabel dari spesifikasi produk
  database/            -> koneksi Postgres
  database/migration/  -> daftar migrasi gormigrate (versioned, manual)
  repository/          -> interface akses data (kontrak)
  repository/postgres/ -> implementasi repository pakai GORM
  usecase/             -> business logic (tidak tahu HTTP/DB detail)
  delivery/http/       -> handler, middleware, dto, router (Fiber)
  pkg/                 -> helper lintas layer: jwt, hash, response, apperror
```

Alur dependency mengikuti Clean Architecture:
`handler -> usecase -> repository (interface) <- repository/postgres (implementasi)`
Layer `usecase` tidak pernah import Fiber atau GORM secara langsung — hanya bergantung pada interface repository.

## Modul yang Sudah Diimplementasi Penuh

- **Auth**: register & login (email/password + JWT). `GoogleID` sudah disiapkan di entity `User` untuk login Google di kemudian hari.
- **Household**: buat rumah tangga, join via kode undangan (expired 7 hari, sekali pakai), generate kode undangan (khusus owner).

## Modul yang Baru Berupa Entity (siap dipakai, tinggal dibuatkan repository/usecase/handler)

`Account`, `Category`, `Transaction`, `Budget`, `Goal`, `Bill`, `BillPeriod`, `NotificationLog` — semua sudah ada di `internal/entity/` dan sudah masuk daftar migrasi. Untuk menambah modul baru, ikuti pola yang sama seperti modul `household`:

1. Buat interface di `internal/repository/xxx_repository.go`
2. Buat implementasi di `internal/repository/postgres/xxx_repository.go`
3. Buat business logic di `internal/usecase/xxx_usecase.go`
4. Buat DTO di `internal/delivery/http/dto/xxx_dto.go`
5. Buat handler di `internal/delivery/http/handler/xxx_handler.go`
6. Daftarkan route di `internal/delivery/http/router.go`

Urutan pengerjaan yang disarankan: `accounts -> categories -> transactions -> budgets -> goals -> bills`.

## Setup

### 1. Siapkan `.env`

```bash
cp .env.example .env
# lalu isi JWT_SECRET dan kredensial database sesuai environment kamu
```

### 2. Install dependency

```bash
go mod tidy
```

### 3. Jalankan migrasi (MANUAL, tidak auto-run)

Migrasi memakai [gormigrate](https://github.com/go-gormigrate/gormigrate) — versioned, dan tidak pernah dijalankan otomatis saat `cmd/api` start. Pastikan Postgres sudah aktif, lalu:

```bash
make migrate-up
# atau
go run ./cmd/migrate up
```

Rollback migrasi terakhir kalau perlu:

```bash
make migrate-down
```

### 4. Jalankan server

```bash
make run
# atau
go run ./cmd/api
```

Server default jalan di `http://localhost:8080`. Cek kesehatan server:

```bash
curl http://localhost:8080/health
```

## Menambah Migrasi Baru

Jangan pernah mengedit migrasi yang sudah pernah dijalankan di environment manapun. Tambahkan entri baru di `internal/database/migration/migrations.go`, di bagian bawah list `Migrations`, dengan ID unik (disarankan format timestamp: `YYYYMMDDHHMMSS_deskripsi`).

## Contoh Request

**Register**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi","email":"budi@example.com","password":"password123"}'
```

**Login**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"password123"}'
```

**Buat Rumah Tangga** (pakai token dari hasil login/register)
```bash
curl -X POST http://localhost:8080/api/v1/households \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Keluarga Budi"}'
```

**Generate Kode Undangan** (hanya owner)
```bash
curl -X POST http://localhost:8080/api/v1/households/invitations \
  -H "Authorization: Bearer <TOKEN>"
```

**Join Rumah Tangga**
```bash
curl -X POST http://localhost:8080/api/v1/households/join \
  -H "Authorization: Bearer <TOKEN_ANGGOTA_LAIN>" \
  -H "Content-Type: application/json" \
  -d '{"code":"ABCD1234"}'
```

## Catatan

- `go.mod` berisi versi dependency indikatif — jalankan `go mod tidy` di environment kamu (dengan akses internet penuh ke proxy Go) untuk resolve versi final & `go.sum`.
- Belum ada rate limiting / refresh token — pertimbangkan menambah kalau butuh keamanan lebih untuk production.
- Saldo akun (`current_balance`) sengaja tidak disimpan sebagai kolom di entity `Account` — dihitung dari `initial_balance` + agregasi `transactions` di layer usecase saat modul Account/Transaction dibuat, supaya selalu konsisten.
