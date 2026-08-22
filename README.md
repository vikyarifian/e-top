# e-TOP

Aplikasi manajemen tugas dan proyek internal berbasis web — mencakup workspace, proyek, tugas harian/tiket, departemen, notifikasi, serta **Penilaian Kinerja** (evaluasi KPI per user) yang dihitung otomatis dari data tugas.

Dibangun dengan pendekatan *server-side rendering* penuh: Go + [templ](https://templ.guide) di sisi server, [HTMX](https://htmx.org) untuk interaksi partial-update, dan [Alpine.js](https://alpinejs.dev) untuk state ringan di browser — tanpa framework SPA.

## Fitur

- **Dashboard** — ringkasan workspace, proyek, tugas, anggota tim, KPI pribadi, distribusi status/tipe tugas, dan grafik penyelesaian bulanan.
- **Workspaces** — wadah kolaborasi dengan anggota dan role (Owner/Admin/Member), undangan via email.
- **Projects** — proyek di dalam workspace dengan kontributor, progress, status (Planning/In Progress/On Hold/Completed/Cancelled), dan tabel tugas ber-paging.
- **Tasks** — tiga tipe: `PROJECT` (terikat proyek), `DAILY` (tugas harian), `TICKET` (tugas dari user lain); dilengkapi prioritas, due date, subtask, watcher, komentar (mention & reaksi), lampiran, serta estimasi vs realisasi jam kerja.
- **My Tasks** — daftar tugas yang di-assign ke user login.
- **Achieved (Penilaian Kinerja)** — evaluasi KPI per user dengan filter tahun; admin dapat melihat evaluasi semua user, kepala departemen dapat melihat evaluasi anggotanya.
- **Departments** — struktur departemen dengan kepala departemen dan anggota.
- **Notifications** — riwayat aktivitas yang relevan untuk user (halaman ber-paging + dropdown di header), dikelompokkan per Today/Yesterday/tanggal.
- **Settings** — profil, password, preferensi notifikasi; khusus admin: kelola user dan departemen.
- **Auth** — sign-up dengan verifikasi email, sign-in JWT (cookie), lupa password, dan login Google OAuth.

## Rumus Penilaian Kinerja

| KPI | Rumus | Bobot |
|---|---|---|
| **TCR** (Task Completion Rate) | tugas selesai ÷ total tugas × 100 | 30% |
| **OTR** (On-Time Rate) | selesai tepat waktu ÷ tugas selesai × 100 | 30% |
| **TPS** (Task Priority Score) | bobot prioritas selesai ÷ total bobot prioritas × 100 | 20% |
| **WER** (Work Efficiency Rate) | rata-rata (estimasi jam ÷ realisasi jam) × 100 | 20% |

Skor akhir dikategorikan: ≥85 **Sangat Baik**, ≥70 **Baik**, ≥55 **Cukup**, ≥40 **Buruk**, sisanya **Sangat Buruk**.

## Teknologi

| Lapisan | Teknologi |
|---|---|
| Bahasa & server | Go 1.25, `net/http` (tanpa framework) |
| Template | [templ](https://github.com/a-h/templ) v0.3 |
| Interaksi UI | HTMX + Alpine.js + Tailwind CSS 3, ikon [Lucide](https://lucide.dev) |
| Database | PostgreSQL, GORM v1.31 (driver pgx) |
| Auth | JWT (`golang-jwt/v5`) via cookie `session_token`, Google OAuth2, bcrypt |
| Email | SMTP (verifikasi akun, undangan, reset password) |

## Struktur Proyek

```
├── main.go              # entry point: env, timezone Asia/Jakarta, DB init, routes
├── auth/                # JWT, middleware RequireAuth, OAuth
├── db/                  # koneksi PostgreSQL (GORM)
├── dto/                 # objek transfer (mis. UserAuth)
├── handlers/            # HTTP handler per fitur
├── models/              # model GORM + DDL tabel (komentar di tiap file)
├── routes/              # registrasi route
├── services/            # logika bisnis & query (dashboard, evaluasi, log, dsb.)
├── templates/
│   ├── layouts/         # base layout, header, sidebar
│   ├── pages/           # halaman penuh (dashboard, sign-in, dsb.)
│   ├── features/        # konten fitur (tasks, projects, achieved, settings, dsb.)
│   └── components/      # komponen reusable (+ ui/: button, table, modal, dsb.)
├── public/              # aset statis (css, js, ikon)
└── cmd/seed/            # seeder data contoh
```

## Menjalankan Secara Lokal

### Prasyarat

- Go 1.25+
- PostgreSQL 14+
- [templ CLI](https://templ.guide/quick-start/installation) (`go install github.com/a-h/templ/cmd/templ@latest`)
- (opsional, untuk ubah styling) Tailwind CSS: `bun add -D tailwindcss@3.3.1 postcss autoprefixer`

### 1. Siapkan database

Buat database PostgreSQL, lalu buat tabel-tabelnya. Skema DDL tersedia sebagai komentar di file `models/model-*.go` (tidak ada AutoMigrate — tabel dibuat manual). Data master (status tugas/proyek, prioritas, setting) akan diisi otomatis saat aplikasi pertama kali jalan (`utils.DbMutation`).

### 2. Konfigurasi environment

Buat file `.env.local` (atau `.env`) di root proyek:

```env
APP_NAME=etop
APP_ENV=dev
APP_PORT=":7878"
APP_URL=http://localhost
APP_PATH=""

JWT_KEY=ganti-dengan-secret-anda

POSTGRES_URL="postgres://user:password@localhost/etop"

# SMTP (verifikasi email, undangan, reset password)
EMAIL_HOST="smtp.gmail.com"
EMAIL_PORT=587
EMAIL_USER=you@example.com
EMAIL_PASS=app-password
EMAIL_FROM=you@example.com

# Google OAuth
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URI="/auth/google/callback"
```

### 3. Jalankan

```sh
go run .
```

Aplikasi tersedia di `http://localhost:7878`.

Untuk pengembangan dengan hot-reload template:

```sh
templ generate --watch --proxy=http://localhost:7878
```

### Seed data contoh (opsional)

```sh
go run ./cmd/seed
```

## Catatan Pengembangan

- Setiap mengubah file `.templ`, jalankan `templ generate` (atau pakai mode `--watch`) lalu build ulang — file `*_templ.go` adalah hasil generate dan ikut di-commit.
- Timezone aplikasi dipatok **Asia/Jakarta** (`time.Local` di-override di `main.go`); kolom timestamp di DB bertipe `timestamp without time zone` berisi wall-clock Jakarta. Gunakan helper `utils.TimeDiff`/`utils.TimeAgo` saat menghitung selisih waktu dari nilai DB.
- Halaman dirender penuh di server; navigasi antar halaman memakai HTMX (`hx-post` + `hx-target="#content"`), jadi handler umumnya melayani `GET` (halaman penuh) dan `POST` (partial untuk swap `#content`).
