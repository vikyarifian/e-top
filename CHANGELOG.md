# Catatan Perubahan

Seluruh perubahan penting pada aplikasi etop dicatat di berkas ini.
Format mengikuti [Keep a Changelog](https://keepachangelog.com/id/1.1.0/).

## [Belum dirilis] — 2026-09-13

Rangkaian perubahan ini mengganti indikator ketiga penilaian kinerja dari
Task Priority Score menjadi **Task Value Score (TVS)**, menambahkan dimensi
dampak pada tugas, dan membuat kedua tabel acuan dapat diatur dari antarmuka.

### Ditambahkan

- **Dimensi dampak pada tugas.** Tabel acuan baru `task_impacts` beserta kolom
  `tasks.impact_id` yang berelasi kepadanya. Nilai kodenya HIGH, MEDIUM, dan
  LOW, diadopsi dari kolom `impact` milik iTop yang berarti dampak selingkup
  departemen, layanan, dan perorangan.
- **Pilihan Impact pada formulir tugas**, baik saat membuat maupun menyunting.
  Nilai bawaannya adalah baris acuan pertama, yaitu dampak terluas, agar
  ketiadaan data tidak menguntungkan siapa pun.
- **Batas tenggat per prioritas.** Kolom `task_priorities.max_due_minutes`
  menetapkan berapa lama sebuah tugas boleh diberi tenggat, dihitung sejak
  tanggal mulai. Nilai awal: HIGH 1.440 menit, MEDIUM 2.880, LOW 4.320. Nilai 0
  berarti tanpa batas. Validasinya berlapis: atribut `max` pada input tanggal,
  pesan galat di sisi peramban, dan pemeriksaan ulang di server.
- **Halaman Settings → Task Config** (khusus ADMIN) untuk mengatur kode, label,
  warna, value, level, bobot, dan batas tenggat pada tabel prioritas maupun
  dampak. Dilengkapi pratinjau matriks nilai tugas yang menghitung ulang secara
  langsung saat bobot diubah. Penyimpanan berjalan dalam satu transaksi lewat
  `PUT /task-config`.
- **Penjalan migrasi `cmd/migrate`.** Proyek ini tidak memakai AutoMigrate dan
  tidak mengandalkan psql, sehingga perubahan skema dijalankan lewat perintah
  ini. Setiap berkas yang berhasil dijalankan dicatat pada tabel
  `schema_migrations` agar tidak terulang.

      go run ./cmd/migrate            jalankan migrasi yang belum diterapkan
      go run ./cmd/migrate -status    tampilkan status tanpa mengubah apa pun

- **Perkakas perawatan data `cmd/dbops`.** Semua perintah yang mengubah data
  berjalan sebagai pratinjau lebih dulu; perubahan baru disimpan bila diberi
  `-apply`.

      go run ./cmd/dbops diag                      periksa keadaan data tugas
      go run ./cmd/dbops sync -peta <csv> -apply   selaraskan dengan sumber iTop
      go run ./cmd/dbops del -year 2026 -apply     hapus tugas menurut tahun

- **Tujuh eksperimen baru pada `cmd/fuzzysim`** untuk keperluan analisis tugas
  akhir: `kaidah` (peran kaidah agregasi dan pembatas), `bobot` (dapatkah bobot
  menggantikan kaidah pembatas), `kredit` (di tabel mana kredit keterlambatan
  dipasang), `merata` (bila sebaran prioritas dan dampak diseragamkan),
  `terbobot` (bila seluruh indikator dibobot), `kandidat` (mencari indikator
  yang benar-benar bergerak), dan `nilai` (cara lain menghitung nilai tugas).

### Diubah

- **Indikator ketiga berganti nama** dari Task Priority Score (TPS) menjadi
  Task Value Score (TVS) pada sebelas berkas, mencakup nama bidang struktur,
  label antarmuka, nama variabel pada premis aturan fuzzy, dan komentar.
- **Rumus TVS.** Sebelumnya mengalikan `task_priorities.level` dengan faktor
  dampak yang ditulis langsung di dalam kueri. Sekarang mengambil kedua bobot
  dari tabel acuan, sehingga nilainya dapat diubah lewat data tanpa menyentuh
  kode:

      TVS = SUM(tp.weight * ti.weight) tugas selesai
          / SUM(tp.weight * ti.weight) seluruh tugas

  Tugas yang tidak memiliki baris acuan tidak diberi bobot bawaan, melainkan
  diabaikan dari pembilang maupun penyebut.
- **Penyebut OTR.** Sebelumnya hanya tugas yang selesai, sehingga tugas yang
  ditelantarkan melewati tenggat tidak pernah menghukum ketepatan waktu —
  menelantarkan tugas justru menghasilkan OTR lebih baik daripada
  menyelesaikannya terlambat. Penyebut sekarang mencakup seluruh tugas yang
  nasibnya sudah pasti, yaitu tugas yang selesai ditambah tugas yang belum
  selesai padahal tenggatnya sudah lewat. Tugas yang belum selesai dan
  tenggatnya belum tiba tidak diikutkan karena masih mungkin tepat waktu.
- **Kode dampak diseragamkan** dari DEPARTMENT, SERVICE, dan PERSON menjadi
  HIGH, MEDIUM, dan LOW agar sejajar dengan penamaan prioritas. Makna lamanya
  tetap tercatat pada komentar migrasi.
- **Kolom warna pada Task Config berupa isian teks**, bukan daftar pilihan.
  `task_priorities.color` menyimpan kelas Tailwind utuh sedangkan
  `task_impacts.color` menyimpan nama warna sederhana, sehingga daftar pilihan
  tertutup akan menimpa nilai yang sudah ada.
- **Rumus TVS dan OTR pada `cmd/fuzzysim` disamakan dengan aplikasi.** Harness
  eksperimen sebelumnya masih memakai `SUM(tp.level)` dan penyebut OTR yang
  lama, sehingga angkanya berbeda dari yang ditampilkan aplikasi.

### Diperbaiki

- **Daftar tugas pada `/tasks` selalu kosong.** Kuerinya menyaring lewat
  `SELECT task_id FROM task_assignees`, padahal tabel itu sudah ditinggalkan
  dan tidak berisi satu baris pun; penugasan kini disimpan sebagai kolom
  `user_id` pada tabel `tasks`. Perlu dicatat bahwa templat halaman ini masih
  berupa rintisan, sehingga datanya mengalir tetapi belum ditampilkan.
- **TVS pada halaman dashboard terbaca sekitar 224 persen** bagi karyawan yang
  menuntaskan seluruh tugasnya. Penyebutnya sudah memakai skala bobot baru
  yang rata-ratanya 0,897 per tugas, tetapi pembilangnya masih memakai
  `SUM(tp.level)` yang rata-ratanya 2,011. Kedua kueri kini disatukan lewat
  satu fungsi pembantu.
- **Kegagalan perhitungan TVS tidak lagi diam-diam menjadi nol.** Galat kueri
  sebelumnya tidak diperiksa, sehingga basis data yang belum menerima migrasi
  menghasilkan TVS nol yang tampak seperti karyawan tidak menuntaskan satu
  tugas pun. Galatnya kini dilaporkan beserta petunjuk perbaikannya.
- **Atribut `stroke-dasharray` pada lingkaran KPI ternoda.** `fmt.Sprintf`
  memakai kata kunci `%s` untuk bilangan bulat, sehingga atributnya berisi
  `%!s(int=100)`. Terjadi pada halaman penilaian dan dashboard.
- **Komentar bobot dampak yang usang** pada `services/service-dashboard.go`
  masih menulis Departemen, Layanan, dan Perorangan setelah kodenya berganti.

### Dihapus

- `cmd/dbgcomment`, alat debug sekali pakai dengan id tugas yang dipatri di
  dalam kode.

### Data

Perubahan berikut diterapkan pada basis data, bukan pada kode.

- Migrasi `001_task_value_score.sql` dan `002_task_config.sql` dijalankan.
- **Kolom impact dan priority diselaraskan dengan sumber iTop.** Penjodohan
  memakai cara yang sama dengan berkas bukti penelitian: agen yang sama, waktu
  mulai yang sama, lalu judul yang sama. Seluruh 13.898 tugas berhasil
  dijodohkan. Sebanyak 600 baris dampak berubah; kolom prioritas ternyata sudah
  sesuai dan tidak diubah.
- **Tugas dengan `created_at` tahun 2026 dihapus**, yaitu 2 tugas beserta 1
  komentar terkait.

### Keamanan dan privasi

- Folder `skripsi/` seluruhnya dimasukkan ke `.gitignore`. Isinya memuat nama
  karyawan, isi tiket perusahaan, dan dokumen yang bukan bagian dari aplikasi.

### Pengujian

- 75 kasus uji black box dijalankan terhadap build hasil perubahan dan
  seluruhnya sesuai harapan: 6 autentikasi, 9 navigasi, 13 penilaian,
  24 simulasi, 14 konfigurasi tugas, dan 9 formulir tugas.

### Catatan bagi pengelola

Sebelum menjalankan versi ini, terapkan migrasi pada setiap basis data yang
dipakai:

    go run ./cmd/migrate -status
    go run ./cmd/migrate

Tanpa migrasi, tabel `task_impacts` beserta kolom `weight` dan
`max_due_minutes` tidak ada, sehingga TVS akan tercatat gagal pada log dan
formulir tugas tidak memuat pilihan Impact.
