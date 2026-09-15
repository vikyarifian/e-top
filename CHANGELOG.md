# Catatan Perubahan

Seluruh perubahan penting pada aplikasi etop dicatat di berkas ini.
Format mengikuti [Keep a Changelog](https://keepachangelog.com/id/1.1.0/).

## [Belum dirilis] — 2026-09-15

Tiga rumus indikator disesuaikan, kaidah pembatas pada basis aturan dihapus,
serta pencarian dan penyaring ditambahkan pada daftar pengguna, departemen, dan
tugas. Waktu muat halaman My Tasks turun sekitar sepuluh kali lipat.

### Ditambahkan

- **Pencarian pada Settings.** Tab Users mencari nama lengkap, username, dan
  email; tab Departments mencari nama dan keterangan. Pencarian berjalan
  seiring pengetikan dengan jeda 400 milidetik.
- **Pencarian dan penyaring pada My Tasks.** Kotak pencarian menelusuri judul
  dan keterangan tugas, disertai tiga penyaring untuk status, prioritas, dan
  jenis, serta tombol untuk membersihkan semuanya sekaligus. Kata kunci dan
  penyaring ikut terbawa oleh tombol halaman berikutnya maupun tombol
  pengurutan.
- **Komponen `ui.SearchBox` dan `ui.FilterSelect`.** Keduanya memakai kata kunci
  dan fungsi penyusun alamat dari cakupan Alpine di sekitarnya, sehingga seluruh
  penyaring pada satu halaman selalu terkirim bersama.
- **Singgahan tabel acuan tugas.** `task_statuses`, `task_priorities`, dan
  `task_impacts` disimpan di memori dan dibatalkan setelah Task Config disimpan.
  Ketiganya sebelumnya ditembak ulang hingga enam kali pada setiap render
  halaman My Tasks.
- **Indeks tabel `tasks`** untuk `user_id`, `created_by`, `status_id`,
  `priority_id`, dan `created_at`, satu indeks parsial untuk `completed_at`,
  serta satu indeks gabungan `(user_id, created_at DESC)`.

### Diubah

- **Penyebut OTR mencakup seluruh tugas yang nasibnya sudah pasti**, yaitu tugas
  yang sudah selesai ditambah tugas yang belum selesai padahal tenggatnya sudah
  lewat. Sebelumnya hanya tugas yang selesai, sehingga menelantarkan tugas
  justru menghasilkan OTR lebih baik daripada menyelesaikannya terlambat.
- **Rumus WER membatasi tiap tugas sebelum dirata-ratakan:**

      WER = rata-rata( min(1 ; estimasi / realisasi) ) x 100%

  Sebelumnya rasio tiap tugas dirata-ratakan dulu baru dipotong pada 100. Tugas
  yang selesai jauh lebih cepat dari perkiraan menghasilkan rasio ribuan dan
  menarik rata-ratanya melewati 100; pada data pengujian, 15 dari 20 karyawan
  tercatat sempurna. Setelah pembatasan per tugas, rentangnya menjadi 20,34
  sampai 97,29.
- **Penyebut TVS memakai cacah tugas**, bukan jumlah bobotnya.
- **Daftar tugas pada My Tasks ditukar sebagai pecahan tersendiri.** Pencarian,
  penyaring, pengurutan, dan penomoran halaman hanya menukar bagian daftar.
  Sebelumnya seluruh isi halaman dirender ulang, termasuk modal buat tugas yang
  menarik seluruh baris tabel `users` dua kali. Pada basis data berisi 349
  pengguna, dua kueri itu memakan 903 dari 1.252 milidetik waktu halaman.
- **Daftar karyawan pada halaman Achieved terurut menurut abjad.** Pengguna yang
  sedang masuk tidak lagi dipaksa ke urutan pertama, melainkan hanya ditandai
  terpilih ketika belum ada pilihan lain.

### Dihapus

- **Kaidah pembatas OTR pada basis aturan.** Kaidah itu menurunkan konsekuen
  menjadi paling tinggi "Cukup" bila OTR berada pada himpunan terendah.
  Pengukuran setelah ketiga perubahan rumus di atas menunjukkan kaidah itu tidak
  lagi mengubah kategori satu pun karyawan, baik pada data penelitian maupun
  pada seluruh riwayat tiket, karena indikator lain sudah lebih dulu menurunkan
  nilai mereka. Konsekuen kini ditentukan sepenuhnya oleh kaidah agregasi.
  Perbandingan kedua rancangan tetap tersedia pada eksperimen `kaidah` di
  `cmd/fuzzysim`.

### Kinerja

Diukur pada basis data berisi 35.004 tugas dan 349 pengguna, dalam keadaan
sambungan hangat, tiga kali tiap permintaan.

| Permintaan | Sebelum | Sesudah |
|---|---:|---:|
| Mengetik di kotak pencarian My Tasks | 1,0–2,4 s | 0,11 s |
| Mengubah penyaring My Tasks | 1,4–2,0 s | 0,11 s |
| Membuka halaman My Tasks | 1,0–1,5 s | 0,16 s |

Jumlah kueri per permintaan turun dari 11 menjadi 5 untuk pencarian dan 7 untuk
halaman penuh. Ukuran jawaban turun dari 231 KB menjadi 63 KB.

### Keamanan dan privasi

- Folder `db/migrations/` dan `cmd/simdb/` dimasukkan ke `.gitignore` atas
  permintaan pemilik proyek, dan berkas migrasi yang sebelumnya terlacak dicabut
  dari pelacakan. **Akibatnya salinan repo tidak lagi memuat berkas migrasi,
  sehingga pemasangan di mesin baru memerlukan salinan berkas itu secara
  terpisah sebelum `cmd/migrate` dapat dijalankan.**
- Luaran server berakhiran `.log` ikut diabaikan.

### Pengujian

- 75 kasus uji black box dijalankan terhadap build hasil perubahan. 74 sesuai;
  satu kasus yang memeriksa keberadaan kategori "Buruk" tidak terpenuhi karena
  basis data pengujian tidak memuat karyawan pada kategori itu.

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
