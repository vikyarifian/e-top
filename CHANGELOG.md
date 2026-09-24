# Catatan Perubahan

Seluruh perubahan penting pada aplikasi etop dicatat di berkas ini.
Format mengikuti [Keep a Changelog](https://keepachangelog.com/id/1.1.0/).

## [Belum dirilis] — 2026-09-25

Administrator kini dapat menyetel ulang sandi pengguna lain dari Settings.

### Ditambahkan

- **Setel ulang sandi pengguna dari Settings > Users.** Kolom Actions mendapat
  tombol bergambar kunci di sebelah tombol sunting, beserta modal berisi sandi
  baru dan pengulangannya. Tombolnya hanya tampil bagi administrator.

  Berbeda dari `/update-password` yang sudah ada, sandi lama tidak ditanyakan:
  yang itu dipakai seseorang untuk sandinya sendiri, sedangkan di sini
  administrator memang tidak mengetahui sandi lama orang yang diurusnya.
  Keduanya karena itu dipisah, bukan digabung.

  Wewenangnya diperiksa di peladen pada kedua rute baru, `/reset-password-form`
  dan `/reset-password`, bukan sekadar disembunyikan tombolnya. Menyembunyikan
  tombol hanya menyembunyikan, tidak menghalangi siapa pun mengirim
  permintaannya langsung.

  Perubahannya dicatat atas nama administrator yang melakukan, bukan atas nama
  pemilik akun, sehingga jejaknya menunjukkan siapa yang benar-benar bertindak.
  Pemilik akun tidak diberi tahu lewat surel, dan hal itu dinyatakan apa adanya
  pada formulirnya.

### Pengujian

- Tampilan: administrator melihat 20 tombol untuk 20 pengguna, anggota biasa
  tidak melihat satu pun.
- Wewenang: anggota biasa menerima 401 pada kedua rute, baik ketika membuka
  formulir maupun ketika mengirim perubahannya.
- Penyahihan: sandi kurang dari enam huruf, pengulangan yang tidak sama, dan
  pengenal pengguna yang tidak ada, ketiganya ditolak dengan 400 beserta
  keterangannya masing-masing.
- Perubahan benar-benar berlaku: setelah disetel ulang, masuk dengan sandi lama
  ditolak dan masuk dengan sandi baru berhasil. Sandi pengguna uji dikembalikan
  ke semula setelahnya, dan jejak pengujiannya dihapus dari tabel logs.
- 75 kasus uji black box dijalankan, 74 sesuai; satu kasus yang memeriksa
  keberadaan kategori "Buruk" tetap tidak terpenuhi karena basis data pengujian
  tidak memuat karyawan pada kategori itu.

## [Belum dirilis] — 2026-09-24

Seluruh label antarmuka kini berbahasa Inggris, skor dan indikator ditampilkan
dengan dua angka di belakang koma, catatan masuk tidak lagi muncul sebagai
pemberitahuan, dan daftar tugas mendapat penyaring luas dampak.

### Ditambahkan

- **Penyaring Impact** pada My Tasks dan pada daftar tugas anggota di halaman
  Department. Pilihannya dibaca dari tabel acuan `task_impacts`, sehingga ikut
  berubah bila acuannya disunting lewat Task Config, sama seperti penyaring
  status dan prioritas. Penyaring baru ini ikut terbawa tombol halaman
  berikutnya maupun tombol pengurutan tanpa perubahan tambahan, sebab tautannya
  disusun `PageInfo.Params()` dari peta penyaring.

### Diubah

- **Label dan keterangan antarmuka diterjemahkan ke bahasa Inggris**, 99 label
  pada sembilan berkas: halaman Simulasi, Department, My Tasks, Settings,
  Penilaian, kotak pencarian, judul halaman Simulasi, nama bentuk fungsi
  keanggotaan beserta parameternya, dan nama serta keterangan prasetel. Kode
  nilai isian formulir seperti `linear-turun` dan `trapesium` tidak diubah
  karena dipakai sebagai kunci, bukan sebagai teks yang dibaca.

  Istilah linguistik fuzzy sengaja tetap berbahasa Indonesia atas keputusan
  pemilik proyek, yaitu himpunan Rendah, Sedang, dan Tinggi beserta kategori
  Sangat Buruk sampai Sangat Baik, supaya cocok dengan naskah penelitian.
  Keterangan kode, pesan log, dan luaran perkakas `cmd/*` juga tetap berbahasa
  Indonesia karena ditujukan kepada peneliti, bukan kepada pengguna aplikasi.
- **Skor dan keempat indikator ditampilkan dengan dua angka di belakang koma.**
  Sebelumnya Dashboard dan halaman Penilaian memakai satu angka sedangkan
  keterangan di bawahnya membulatkan ke bilangan bulat, sehingga nilai yang
  sama tampil sebagai 93,7 persen dan 94 persen pada satu layar. Derajat
  keanggotaan dan alpha tetap empat angka, sebab keduanya berkisar nol sampai
  satu dan dua angka akan menghapus perbedaan yang berarti.
- **Pemberitahuan tidak lagi memuat catatan akun.** Aksi `login_user`,
  `verified_user`, `registered_user`, `forgot_password_user`, dan
  `resend_email_user` disaring keluar; pemilik akunnya sendiri yang melakukan
  semuanya, jadi tidak ada gunanya diberitahukan kembali kepadanya.

### Pengujian

- Lima halaman dirender lalu disisir dengan daftar tiga puluh kata Indonesia;
  seluruhnya bersih. Yang tersisa hanya istilah linguistik fuzzy yang memang
  dipertahankan.
- Penyaring Impact diuji bersama penyaring lain, bukan sendiri-sendiri. Pada
  My Tasks: High 1.934, Medium 61, Low 415, berjumlah 2.410 sama dengan cacah
  tugas pemiliknya; High bersama priority Low menyisakan 1.889; kata kunci
  "email" bersama High menyisakan 9. Pada daftar departemen: High 27.695,
  Medium 1.572, Low 1.549, berjumlah 30.816.
- Halaman pemberitahuan diperiksa: nol penyebutan login maupun verifikasi
  surel, sementara peristiwa tugas tetap tampil.
- Terjemahan sempat menjatuhkan 19 kasus uji black box karena penandanya
  memang caption yang diubah. Penandanya disesuaikan tanpa mengubah apa yang
  diperiksa. Hasil akhir 75 kasus dijalankan, 74 sesuai; satu kasus yang
  memeriksa keberadaan kategori "Buruk" tetap tidak terpenuhi karena basis data
  pengujian tidak memuat karyawan pada kategori itu.

## [Belum dirilis] — 2026-09-22

Empat indikator pada Dashboard selama ini dihitung dengan cara yang berbeda
dari halaman Penilaian, sehingga angkanya bisa tidak cocok untuk orang yang
sama; keduanya kini memakai definisi yang sama. Selain itu menu Simulasi
disembunyikan, dan sisa keterangan kaidah pembatas yang sudah tidak berlaku
dibersihkan dari halaman Simulasi.

### Diubah

- **Menu Simulasi disembunyikan** dari bilah sisi atas permintaan pemilik
  proyek. Rutenya sengaja dibiarkan hidup, sehingga halamannya tetap dapat
  dibuka langsung lewat `/simulation` ketika diperlukan untuk peragaan.
- **Kasus uji black box nomor 15 disesuaikan.** Sebelumnya memeriksa bahwa menu
  Simulasi ada pada bilah sisi; sekarang memeriksa bahwa menu itu justru tidak
  ada, sementara halamannya tetap terbuka. Pembantu uji baru `ta` ditambahkan
  untuk penanda yang memang harus absen, karena pembantu `t` yang sudah ada
  hanya dapat memeriksa penanda yang harus hadir.

### Diperbaiki

- **KPI pada Dashboard memakai definisi yang berbeda dari halaman Penilaian.**
  Keduanya menampilkan empat indikator dengan nama yang sama untuk orang yang
  sama, tetapi menghitungnya dengan cara berbeda, sehingga angkanya bisa tidak
  cocok. Dua sebabnya:

  *Cakupan tugas.* Dashboard menyaring dengan `user_id = ? OR created_by = ?`,
  sedangkan halaman Penilaian hanya `user_id = ?`. Akibatnya tiket yang
  dilaporkan seorang karyawan tetapi dikerjakan orang lain ikut terhitung
  sebagai tugasnya sendiri.

  *Penyebut OTR.* Dashboard hanya menghitung tugas yang sudah selesai,
  sedangkan halaman Penilaian juga memasukkan tugas yang belum selesai padahal
  tenggatnya sudah lewat. Tugas terlambat yang dibiarkan menggantung karena itu
  lenyap dari penyebut, dan ketepatan waktu terbaca lebih tinggi daripada yang
  sebenarnya. Pada basis data penelitian, Sutedy yang baru menuntaskan 52,38
  persen tugasnya terbaca **20,00 persen pada halaman Penilaian tetapi 38,18
  persen pada Dashboard**, selisih 18,18 angka.

  Dashboard kini memakai definisi halaman Penilaian untuk keempat indikator,
  termasuk kartu jumlah tugas, sebaran status, sebaran jenis, dan grafik
  penyelesaian bulanan, supaya seluruh angka pada satu layar berasal dari
  cakupan yang sama.

- **Keterangan kaidah pembatas pada halaman Simulasi.** Panel "Basis Aturan"
  yang seluruhnya dikomentari masih menyatakan bahwa OTR pada himpunan terendah
  membatasi konsekuen paling tinggi Cukup. Kaidah itu sudah dihapus dari mesin
  inferensi, sehingga keterangannya diganti dengan kaidah agregasi yang berlaku
  sekarang. Panelnya sendiri dibiarkan utuh agar masih dapat dinyalakan kembali.

### Pengujian

- Basis aturan kedua halaman diperiksa ulang untuk memastikan kaidah pembatas
  benar-benar tidak ada. Pada halaman Penilaian terdapat empat aturan berpremis
  OTR Rendah yang berkonsekuen Baik, yaitu R19, R20, R22, dan R46; keempatnya
  mustahil ada bila pembatas masih terpasang, sebab dahulu semuanya dipaksa
  turun ke Cukup. Kelima prasetel halaman Simulasi menunjukkan pola yang sama.
- Keluaran `GetDashboardData` dan `GetAchievedEvaluation` dibandingkan lewat
  jalur kode sungguhan untuk 18 karyawan pada kedua basis data. Sebelum
  perbaikan angkanya berbeda; sesudahnya seluruh 18 cocok sampai dua angka
  di belakang koma, pada keempat indikator maupun cacah tugasnya.
- 75 kasus uji black box dijalankan terhadap build hasil perubahan. 74 sesuai;
  satu kasus yang memeriksa keberadaan kategori "Buruk" tidak terpenuhi karena
  basis data pengujian tidak memuat karyawan pada kategori itu.

## [Belum dirilis] — 2026-09-16

Riwayat tiap tugas pada basis data simulasi dilengkapi sehingga panel Activity
dan Comments tidak lagi kosong. Selain itu tiga kendali yang selama ini tidak
tampil diperbaiki, seluruhnya berpangkal pada perbandingan yang terlalu ketat:
tombol Add Task pada halaman My Tasks dan pada halaman project, serta daftar
peran pada Settings dan Workspace. Pengalih workspace disembunyikan.

### Ditambahkan

- **Riwayat dan komentar tugas pada basis data simulasi.** Perintah baru
  `go run ./cmd/simdb jejak` menyusun ulang riwayat tiap tugas langsung dari
  berkas iTop: tugas dibuka oleh pelapornya pada `start_date`, diserahkan
  kepada agen pada `assignment_date`, dinyatakan selesai lalu ditandai rampung
  pada `resolution_date`, dan ditutup pada tanggal yang sama. Tanggal penutup
  sengaja diambil dari `resolution_date`, bukan `close_date`, mengikuti aturan
  yang sudah dipakai kolom `completed_at`; `close_date` pada iTop kerap
  tercatat massal sehingga berselang jauh dari penyelesaian sebenarnya.
  Komentarnya diambil dari kolom `solution`, yaitu keterangan penyelesaian yang
  ditulis agen ketika menutup tiket. Hasilnya 174.828 baris riwayat dan 34.940
  komentar untuk 35.004 tugas, seluruhnya terhubung, tanpa satu pun baris yatim.
- **Pelapor tiket sebagai pengguna.** Pelapor dicocokkan dengan pengguna yang
  sudah ada lewat surel, lalu lewat nama lengkapnya, lalu lewat nama
  belakangnya karena agen yang dibuat `cmd/simdb` hanya memakai nama belakang.
  Yang belum ada dibuatkan pengguna baru; tanpa itu barisnya tertolak oleh
  kunci asing `logs.user_id`. Pada basis data simulasi 121 pengguna baru
  dibuat, dan 587 tiket tanpa pelapor yang dikenali dicatat atas nama agennya.
- **Indeks tabel `logs` dan `comments`** (`db/migrations/004_log_indexes.sql`):
  gabungan `(resource_type, resource_id)` untuk panel Activity, gabungan
  `(user_id, created_at DESC)` untuk daftar pemberitahuan, dan gabungan
  `(task_id, created_at)` untuk komentar. Tabel `logs` sebelumnya sama sekali
  tanpa indeks; setelah riwayatnya lengkap, setiap kali panel Activity dibuka
  seluruh tabel dipindai. Waktu kuerinya turun dari 97,2 milidetik menjadi
  0,126 milidetik.

### Diubah

- **Pengalih workspace disembunyikan dan dimatikan** atas permintaan pemilik
  proyek. Kotaknya dilepas dari bilah atas dan dari bilah sisi versi telepon,
  pemanggilan `htmx.ajax` yang mengisinya dilepas dari tata letak dasar dan dari
  formulir workspace, dan rutenya dinonaktifkan. Penangan beserta komponennya
  sengaja dibiarkan utuh supaya mudah dihidupkan kembali. Perpindahan antar
  workspace dilakukan lewat menu Workspaces.

### Diperbaiki

- **Peran keanggotaan dibandingkan peka huruf.** Aplikasi selalu menulis peran
  dengan huruf besar, tetapi data hasil impor menyimpannya dengan huruf kecil:
  seluruh 70 baris `project_members` dan 294 baris `workspace_members`
  berperan `member`, bukan `MEMBER`. Akibatnya tombol Add Task pada halaman
  project tidak pernah tampil bagi anggota biasa, dan beberapa kendali lain
  pada halaman workspace serta daftar peran pada Settings ikut salah. Seluruh
  perbandingan peran kini melewati `utils.Peran` yang menyeragamkan
  penulisannya lebih dulu. Datanya sengaja tidak diubah: yang perlu bertoleransi
  adalah aplikasinya, bukan hasil impornya.
- **Riwayat tugas mencatat penyelesaian dua kali.** Basis data penelitian
  mencatatnya sebagai `updated_task` dan `completed_task` dengan keterangan yang
  sama persis, sehingga panel Activity menampilkan baris kembar. Kini hanya
  `completed_task` yang ditulis, dan jumlah barisnya turun dari 174.828 menjadi
  139.888.
- **Urutan riwayat tugas terbaca terbalik.** Baris "selesai" dan "ditutup"
  berbagi cap waktu yang sama karena iTop hanya mencatat satu `resolution_date`.
  Tanpa pemecah seri, keduanya tampil dengan urutan sembarang. Panel Activity
  kini mengurutkan dengan `created_at DESC, id DESC NULLS LAST`; nomor baris
  ditulis menurut urutan kejadian, sedangkan baris buatan aplikasi yang belum
  bernomor ditempatkan paling belakang di dalam cap waktu yang sama.
- **Tombol bervarian `primary` tampil tanpa warna latar.** Tombol "Add Task"
  pada halaman My Tasks adalah satu-satunya pemakainya, dan selama ini tampil
  polos sehingga sukar dikenali sebagai tombol. templ menyusun daftar kelas
  memakai peta yang berkunci teks kelasnya sendiri, sehingga dua baris
  `templ.KV` dengan teks kelas persis sama saling menimpa dan yang terakhir
  menentukan. Baris varian `primary` dan varian kosong kebetulan memakai teks
  kelas yang sama, sehingga ketika varian `primary` diminta, barisnya justru
  dimatikan oleh baris varian kosong di bawahnya. Kedua syarat kini disatukan
  dalam satu baris.
- **Kekeliruan serupa pada tiga komponen lain** turut diperbaiki meski belum
  terpakai: `ui.Badge` varian `default` dan `secondary`, `ui.Avatar` ukuran
  `md`, serta `ui.Toast` jenis `default`.

### Pengujian

- 75 kasus uji black box dijalankan ulang. 74 sesuai; satu kasus yang memeriksa
  keberadaan kategori "Buruk" tetap tidak terpenuhi karena basis data pengujian
  tidak memuat karyawan pada kategori itu.
- Panel Activity dan Comments diperiksa lewat `/task-activities` dan
  `/task-comments` pada satu tugas contoh: kelimanya tampil berurutan, dengan
  pelapor sebagai pembuka dan penutup, agen sebagai pengerja.

## [Belum dirilis] — 2026-09-15

Tiga rumus indikator disesuaikan, kaidah pembatas pada basis aturan dihapus,
serta pencarian dan penyaring ditambahkan pada daftar pengguna, departemen, dan
tugas. Waktu muat halaman My Tasks turun sekitar sepuluh kali lipat.

### Ditambahkan

- **Daftar tugas anggota pada halaman Department.** Halaman itu sebelumnya
  hanya menampilkan daftar anggota, sehingga kepala departemen dapat melihat
  nilai kinerja bawahannya lewat halaman penilaian tetapi tidak dapat melihat
  pekerjaan apa yang sedang mereka kerjakan. Daftar ini dilengkapi pencarian
  judul dan keterangan serta penyaring anggota, status, dan prioritas.
  Cakupannya ditentukan dari keanggotaan departemen, bukan dari peran
  pengguna, sehingga tidak pernah memperlihatkan tugas di luar departemen
  yang sedang dibuka.
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
- **Penyaring karyawan tidak lagi membuka data pengguna yang sedang masuk.**
  Yang terpilih pada muatan awal adalah nama teratas menurut abjad. Bagi
  atasan yang menilai banyak orang, sebelumnya dirinya sendiri yang selalu
  muncul lebih dulu. Karyawan yang tidak punya anggota tetap melihat datanya
  sendiri.
- **Halaman Simulasi hanya menghitung ulang lewat tombol.** Menggeser atau
  mengetik nilai indikator tidak lagi mengirim apa pun; hasilnya diperbarui
  ketika tombol Hitung Ulang ditekan. Sebelumnya setiap geseran memicu
  pengiriman, dan nilai yang baru saja diubah justru kembali ke angka basis
  data. Penyebabnya htmx menyusun data permintaan dari form.elements yang
  juga memuat tombol, sehingga penanda action=reload pada tombol "Muat ulang
  nilai dari basis data" ikut terkirim pada setiap perubahan. Penanda itu
  kini disimpan pada isian tersembunyi, dan tiap kendali menentukan sendiri
  apakah pengirimannya membaca basis data atau tidak.
- **Halaman Simulasi mendapat tombol Hitung Ulang.** Tombol itu menghitung
  ulang memakai nilai yang tampil di layar tanpa membaca basis data lagi,
  berguna ketika angka diketik langsung pada kotaknya sebab peristiwa change
  baru terpicu saat kotak ditinggalkan. Tombol memuat ulang dari basis data
  tetap tersedia terpisah.
- **Daftar karyawan pada halaman Achieved dan Simulasi terurut menurut abjad.**
  Pengguna yang sedang masuk tidak lagi dipaksa ke urutan pertama, melainkan
  hanya ditandai terpilih ketika belum ada pilihan lain. Berlaku pula bagi
  kepala departemen yang melihat daftar anggotanya.

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

- **Halaman rincian workspace tidak lagi memuat seluruh tugas.** Daftar tugas
  tiap project dipakai hanya untuk menampilkan jumlahnya, tetapi dimuat penuh
  lewat Preload. Bagi pengguna yang menjadi anggota project besar, membuka
  satu workspace memerlukan 4,24 detik; kini 0,66 detik karena jumlahnya
  diambil lewat satu kueri agregat.

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
