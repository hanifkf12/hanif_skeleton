# hanif_skeleton

Skeleton backend Go dengan Clean Architecture: satu codebase, tiga entry point (HTTP server, Pub/Sub worker, background worker) yang berbagi usecase dan repository yang sama.

## Requirements

- Go 1.24+
- PostgreSQL
- Redis (opsional — hanya untuk `CACHE_DRIVER=redis` / `QUEUE_DRIVER=redis`)

## Setup

```bash
git clone https://github.com/hanifkf12/hanif_skeleton.git
cd hanif_skeleton
cp .env.example .env
```

Buka `.env` lalu isi minimal tiga nilai berikut — **jangan** pakai nilai contoh:

| Variabel | Cara generate |
|---|---|
| `DB_PASSWORD` | password PostgreSQL lokal |
| `JWT_SECRET_KEY` | `openssl rand -base64 32` |
| `ENCRYPTION_KEY` | `openssl rand -base64 32` |

Sisanya sudah punya nilai default yang bisa dipakai apa adanya. Driver tiap subsistem dipilih lewat `*_DRIVER`:

| Subsistem | Driver yang tersedia | Default |
|---|---|---|
| `DB_DRIVER` | `postgres` | `postgres` |
| `CACHE_DRIVER` | `memory`, `redis` | `memory` |
| `QUEUE_DRIVER` | `asynq` (butuh Redis) | `asynq` |
| `STORAGE_DRIVER` | `local`, `gcs`, `s3` | `local` |

Migrasi lalu jalankan server:

```bash
go run main.go db:migrate
make run-http
```

Server listen di `PORT` (default `9000`). Cek dengan `curl localhost:9000/health`.

## Commands

```bash
make run-http      # HTTP server        (go run main.go http)
make run-worker    # background worker  (go run main.go worker)
make run-pubsub    # Pub/Sub consumer   (go run main.go pubsub)
make test          # go test ./...
make vet           # go vet ./...
go run main.go db:migrate           # jalankan migrasi
go run main.go db:migrate --dir=... # direktori migrasi lain
```

Flag `db:migrate` yang lain: `--table` (nama tabel migrasi, default `db`), `--verbose`, `--guide`.

`cmd/root.go` adalah satu-satunya tempat perintah-perintah ini didaftarkan. Menambah entry point baru berarti menambah satu `*cobra.Command` dan satu direktori di bawah `cmd/`.

### Menjalankan dari Zed

Konfigurasi task `.zed/tasks.json` bersifat lokal dan diabaikan Git; tidak perlu mengubah
konfigurasi global Zed. Untuk setup di mesin lain, buka `zed: open project tasks` dari
Command Palette dan definisikan task memakai perintah pada tabel di bawah. Setiap task
memakai `"cwd": "$ZED_WORKTREE_ROOT"` dan `"save": "all"`; task server/worker memakai
`"use_new_terminal": false` dan `"allow_concurrent_runs": false`.

1. Buka folder root project ini di Zed, bukan hanya `main.go`.
2. Siapkan `.env` dan PostgreSQL sesuai bagian Setup di atas.
3. Untuk HTTP server dan Pub/Sub consumer, pastikan OpenTelemetry collector menerima
   koneksi gRPC di `localhost:4317`. Startup saat ini menunggu koneksi tersebut.
   Worker membutuhkan Redis untuk queue `asynq`. Pub/Sub consumer juga membutuhkan
   `GOOGLE_CLOUD_PROJECT` di environment proses, kredensial Google Cloud, dan subscription
   `user-created-subscription`.
4. Buka Command Palette (`Cmd+Shift+P` di macOS), pilih `task: spawn`, lalu pilih task:

| Task Zed | Perintah |
|---|---|
| `Run HTTP server` | `make run-http` |
| `Run background worker` | `make run-worker` |
| `Run Pub/Sub consumer` | `make run-pubsub` |
| `Run database migrations` | `go run main.go db:migrate` |
| `Run all tests` | `make test` |
| `Run go vet` | `make vet` |
| `Show project CLI help` | `go run main.go --help` |

Semua task menyimpan buffer yang diubah dan berjalan dari root project, sehingga aplikasi
membaca `.env` lokal dan migrasi memakai direktori yang benar. Output tampil di terminal Zed.
Task server/worker tidak mengizinkan instance paralel dari task yang sama. Untuk restart,
hentikan proses dengan `Ctrl+C` di terminal task lalu pilih `task: rerun` dari Command Palette.
Perubahan kode tidak memicu restart otomatis.

Setelah HTTP server siap, cek `curl localhost:9000/health` (sesuaikan dengan `PORT` di `.env`).
Gunakan `Show project CLI help` untuk mengecek toolchain tanpa menjalankan service atau migrasi.

## Architecture

Arah dependensi: **`entity` ← `usecase` ← `handler` / `router`**, dengan `repository` sebagai port yang diimplementasikan di luar dan di-inject dari `router`.

```
cmd/           entry point (http, worker, pubsub, migration) — cobra
internal/
  entity/      tipe domain + aturan bisnis. Tanpa dependensi luar selain stdlib.
  usecase/     orkestrasi. Hanya tahu port appctx.Request dan interface repository.
  repository/  port (contract.go) + implementasi Postgres (sqlerr.go memetakan error driver)
  handler/     adapter HTTP → appctx.Request
  middleware/  JWT, role, HMAC, validasi content-type
  router/      wiring: menyusun repository → usecase → route
pkg/           library yang bisa dipakai ulang, tidak tahu domain
```

Beberapa hal yang ditegakkan, bukan sekadar konvensi:

- **Business layer tidak meng-import framework HTTP.** `internal/usecase`, `internal/repository`, `internal/entity`, dan `internal/appctx` bebas dari Fiber. Usecase menerima `appctx.Request` (interface), dan `handler.fiberRequest` adalah satu-satunya tempat Fiber diterjemahkan. Ganti Fiber dengan `net/http` tidak akan menyentuh satu pun file usecase.
- **Repository tidak menerima DTO transport.** `CreateUser` menerima `entity.UserInput`, `UpdateUser` menerima `entity.UserUpdate`. `UserInput` secara struktural tidak bisa merepresentasikan password plaintext — ia hanya punya `PasswordHash`.
- **Aturan otorisasi tinggal di domain.** `entity.Actor` punya method `CanAccessUser` / `CanModifyUser` (self **atau** admin). Karena aturannya butuh path param, ia tidak bisa dinyatakan sebagai daftar role statis di middleware.
- **Error terklasifikasi.** `pkg/apperror` memisahkan `Message` (aman untuk klien) dari `cause` (hanya untuk log). `internal/repository/sqlerr.go` adalah satu-satunya tempat yang tahu soal `lib/pq`, sehingga `*pq.Error` yang memuat nama tabel dan constraint tidak pernah sampai ke klien.

Verifikasi gerbang arsitektur:

```bash
grep -rn "gofiber" internal/usecase internal/repository internal/entity internal/appctx   # harus kosong
go list -deps ./internal/usecase/... | grep -c gofiber                                     # harus 0
```

## Request Lifecycle

```
HTTP → TraceMiddleware (pkg/app/app.go)
     → JWTAuth / RequireRole / ContentTypeValidator  (per-route)
     → router.handle → handler.HttpRequest
     → usecase.Serve(appctx.Data)
     → repository  →  PostgreSQL
     ← appctx.Response  →  router.response  →  JSON
```

`handler.HttpRequest` membungkus `*fiber.Ctx` menjadi `appctx.Request` dan menaruh `entity.Actor` hasil parse JWT di dalamnya. Usecase mengambilnya lewat `data.Request.Principal()`.

Semua respons (sukses maupun error) memakai envelope yang sama. Error dipetakan dengan `appctx.ResponseFromError` — satu-satunya tempat `apperror` diterjemahkan menjadi kode HTTP:

```go
return *appctx.ResponseFromError(apperror.NotFound("Campaign not found"))
```

Alasan error ditulis ke field `errors`, dan `message` dibiarkan kosong — konsisten dengan seluruh respons error yang ada.

## API

| Method | Path | Auth | Middleware tambahan |
|---|---|---|---|
| GET | `/health` | — | — |
| POST | `/auth/login` | — | — |
| POST | `/auth/refresh` | — | — |
| GET | `/users` | JWT | — |
| POST | `/users` | JWT | `RequireRole(["admin"])`, content-type JSON |
| GET | `/users/:id` | JWT | self **atau** admin (aturan domain) |
| PUT | `/users/:id` | JWT | content-type JSON, self **atau** admin (aturan domain) |
| DELETE | `/users/:id` | JWT | `RequireRole(["admin"])` |
| GET | `/campaigns` | JWT | — |
| GET | `/campaigns/:id` | JWT | — |
| POST | `/campaigns` | JWT | content-type JSON |
| PUT | `/campaigns` | JWT | content-type JSON |
| DELETE | `/campaigns/:id` | JWT | — |

`GET` dan `PUT /users/:id` sengaja **tidak** memakai `RequireRole(["admin"])`: endpoint-nya ada supaya user bisa mengubah profilnya sendiri, jadi aturannya adalah "self atau admin". Aturan itu hidup di `entity.Actor` dan diperiksa **sebelum** query dijalankan — request yang ditolak tidak menyentuh database sama sekali.

Contoh:

```bash
# login
curl -s -X POST localhost:9000/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# ambil token dari respons, lalu:
curl -s localhost:9000/users -H "Authorization: Bearer $TOKEN"
```

## Testing

```bash
make test
```

Test usecase adalah unit test murni — tanpa HTTP, tanpa database, tanpa Fiber. `internal/appctx/appctxtest` menyediakan `Request` in-memory (pola yang sama dengan `net/http/httptest`):

```go
resp := NewHealth().Serve(appctx.Data{Request: appctxtest.New()})
require.Equal(t, appctx.StatusOK, resp.Code)
```

Repository palsu (`fake_user_repository_test.go`) mencatat jumlah pemanggilan tulis. Test otorisasi meng-assert angka itu, bukan hanya kode status — assert `403` saja akan tetap lolos walau usecase menulis dulu lalu melaporkan 403:

```go
require.Equal(t, 0, repo.updateCalls, "percobaan yang ditolak tidak boleh menyentuh database")
```

`appctxtest` tidak ikut ke binary produksi: linker Go hanya menyertakan paket yang reachable dari `main`.

## Security Notes

- **`.env` tidak dilacak.** Nilai asli hanya ada di file lokal. `.env.example` berisi daftar variabel dengan nilai yang sudah disanitasi.
- **Ganti `JWT_SECRET_KEY` dan `ENCRYPTION_KEY` sebelum deploy.** Nilai di `.env.example` adalah placeholder literal. `JWT_SECRET_KEY` yang diketahui penyerang berarti siapa pun bisa menerbitkan token yang valid untuk aplikasi ini.
- `ENCRYPTION_KEY` hanya dibutuhkan kalau `pkg/crypto` dipakai. Library-nya tersedia, tapi belum ada usecase yang memanggilnya.
- Password di-hash dengan bcrypt. Karena bcrypt memotong input di 72 byte dan mengembalikan `ErrPasswordTooLong` untuk yang lebih panjang, validator membatasi `max=72` sehingga password kepanjangan menjadi `400`, bukan `500`.
- Login yang gagal selalu menjawab `401` dengan pesan yang sama, baik username tidak dikenal maupun password salah — klien tidak bisa membedakan keberadaan akun.

## Menambah Endpoint Baru

1. `internal/entity/` — tipe request/response, plus aturan bisnisnya kalau ada.
2. `internal/repository/contract.go` — tambahkan method ke interface.
3. `internal/repository/<domain>/` — implementasinya. Bungkus error driver dengan `MapSQLError`.
4. `internal/usecase/` — satu file satu usecase, implementasi `contract.UseCase`. Akses request lewat `data.Request`, jangan lewat Fiber.
5. `internal/router/router.go` — daftarkan route beserta middleware-nya.

Setelah semua langkah: `make vet && make test`.
