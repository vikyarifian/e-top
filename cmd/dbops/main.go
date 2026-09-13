// Perintah dbops adalah perkakas perawatan data tugas.
//
//	go run ./cmd/dbops diag              memeriksa keadaan data tugas
//	go run ./cmd/dbops sync -peta x.csv  menyamakan impact dan priority dengan asal iTop
//	go run ./cmd/dbops del -year 2026    menghapus tugas menurut tahun dibuatnya
//
// Semua perintah yang mengubah data berjalan sebagai pratinjau lebih dulu.
// Tambahkan -apply agar perubahan benar-benar disimpan.
//
// Basis data diambil dari tanda -dsn, env POSTGRES_URL atau SEED_DSN, atau
// berkas .env.local.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Perintah: diag | sync | del")
		os.Exit(1)
	}
	perintah := os.Args[1]
	fs := flag.NewFlagSet(perintah, flag.ExitOnError)
	dsn := fs.String("dsn", "", "alamat basis data")
	apply := fs.Bool("apply", false, "simpan perubahan, bukan sekadar pratinjau")
	peta := fs.String("peta", "", "berkas CSV pemetaan asal iTop")
	tahun := fs.Int("year", 0, "tahun created_at yang dituju")
	user := fs.String("user", "", "id pengguna untuk pemeriksaan daftar tugas")
	fs.Parse(os.Args[2:])

	db := sambung(*dsn)

	switch perintah {
	case "diag":
		diag(db, *user)
	case "sync":
		if *peta == "" {
			fmt.Println("Tanda -peta wajib diisi.")
			os.Exit(1)
		}
		sync(db, *peta, *apply)
	case "del":
		if *tahun == 0 {
			fmt.Println("Tanda -year wajib diisi.")
			os.Exit(1)
		}
		hapusTahun(db, *tahun, *apply)
	default:
		fmt.Println("Perintah tidak dikenal:", perintah)
		os.Exit(1)
	}
}

func sambung(dsn string) *gorm.DB {
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_URL")
	}
	if dsn == "" {
		dsn = os.Getenv("SEED_DSN")
	}
	if dsn == "" {
		if isi, err := os.ReadFile(".env.local"); err == nil {
			for _, b := range strings.Split(string(isi), "\n") {
				b = strings.TrimSpace(b)
				if strings.HasPrefix(b, "POSTGRES_URL") {
					if bagian := strings.SplitN(b, "=", 2); len(bagian) == 2 {
						dsn = strings.Trim(strings.TrimSpace(bagian[1]), `"'`)
					}
					break
				}
			}
		}
	}
	if dsn == "" {
		fmt.Println("Alamat basis data belum diatur. Pakai -dsn atau setel POSTGRES_URL.")
		os.Exit(1)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("Gagal tersambung:", err)
		os.Exit(1)
	}
	// tampilkan tujuan tanpa membocorkan sandi
	tampil := dsn
	if i := strings.Index(tampil, "://"); i >= 0 {
		if j := strings.Index(tampil[i+3:], "@"); j >= 0 {
			tampil = tampil[:i+3] + "***" + tampil[i+3+j:]
		}
	}
	fmt.Println("basis data:", tampil)
	return db
}

// ------------------------------------------------------------------ diag

func diag(db *gorm.DB, userID string) {
	tabel := func(judul, kueri string, arg ...any) {
		fmt.Println("\n---", judul)
		rows, err := db.Raw(kueri, arg...).Rows()
		if err != nil {
			fmt.Println("   GALAT:", err)
			return
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		n := 0
		for rows.Next() {
			vals := make([]any, len(cols))
			ptr := make([]any, len(cols))
			for i := range vals {
				ptr[i] = &vals[i]
			}
			rows.Scan(ptr...)
			for i, c := range cols {
				fmt.Printf("   %s=%v", c, vals[i])
			}
			fmt.Println()
			n++
		}
		if n == 0 {
			fmt.Println("   (tidak ada baris)")
		}
	}

	tabel("jumlah tugas", `SELECT COUNT(*) total,
		COUNT(*) FILTER (WHERE completed_at IS NOT NULL) selesai,
		COUNT(*) FILTER (WHERE is_archived) diarsipkan,
		COUNT(*) FILTER (WHERE user_id IS NULL) tanpa_pemilik FROM tasks`)
	tabel("tugas per tahun created_at", `SELECT EXTRACT(YEAR FROM created_at)::int tahun, COUNT(*) n
		FROM tasks GROUP BY 1 ORDER BY 1`)
	tabel("sebaran prioritas", `SELECT t.priority_id, t.priority, COUNT(*) n FROM tasks t GROUP BY 1,2 ORDER BY 3 DESC`)
	tabel("sebaran dampak", `SELECT t.impact_id, t.impact, COUNT(*) n FROM tasks t GROUP BY 1,2 ORDER BY 3 DESC`)
	tabel("acuan prioritas", `SELECT no, priority, label, weight, max_due_minutes FROM task_priorities ORDER BY no`)
	tabel("acuan dampak", `SELECT no, impact, label, weight FROM task_impacts ORDER BY no`)
	tabel("tugas yang acuannya tidak ketemu", `SELECT
		COUNT(*) FILTER (WHERE tp.no IS NULL) prioritas_yatim,
		COUNT(*) FILTER (WHERE ti.no IS NULL) dampak_yatim
		FROM tasks t
		LEFT JOIN task_priorities tp ON tp.no = t.priority_id
		LEFT JOIN task_impacts ti ON ti.no = t.impact_id`)
	tabel("10 pemilik tugas terbanyak", `SELECT u.id, u.full_name, u.level, COUNT(*) n
		FROM tasks t LEFT JOIN users u ON u.id = t.user_id
		GROUP BY 1,2,3 ORDER BY 4 DESC LIMIT 10`)
	// meniru kueri yang dipakai tiap halaman daftar tugas
	tabel("jumlah baris menurut kueri tiap halaman", `SELECT u.full_name,
		(SELECT COUNT(*) FROM tasks t WHERE t.user_id = u.id OR t.created_by = u.id) my_tasks,
		(SELECT COUNT(*) FROM tasks t WHERE t.id IN (SELECT task_id FROM task_assignees WHERE user_id = u.id)) halaman_all,
		(SELECT COUNT(*) FROM tasks t WHERE t.user_id = u.id
			AND t.status_id NOT IN (SELECT no FROM task_statuses WHERE status = 'CANCELLED')) achieved
		FROM users u WHERE u.id IN (SELECT DISTINCT user_id FROM tasks) ORDER BY 2 DESC`)
	tabel("acuan status", `SELECT no, status, label, form, value, level FROM task_statuses ORDER BY no`)
	tabel("tabel task_assignees", `SELECT COUNT(*) n FROM task_assignees`)

	if userID != "" {
		tabel("kueri halaman My Tasks untuk pengguna ini",
			`SELECT COUNT(*) n FROM tasks WHERE user_id = ? OR created_by = ?`, userID, userID)
		tabel("kueri halaman Tasks untuk pengguna ini",
			`SELECT COUNT(*) n FROM tasks WHERE id IN (SELECT task_id FROM task_assignees WHERE user_id = ?)`, userID)
	}
}

// ------------------------------------------------------------------ sync

type barisPeta struct {
	TaskID    string
	ImpactID  int
	ImpactKod string
	PrioKod   string
}

func sync(db *gorm.DB, berkas string, apply bool) {
	f, err := os.Open(berkas)
	if err != nil {
		fmt.Println("Gagal membuka berkas peta:", err)
		os.Exit(1)
	}
	defer f.Close()
	rd := csv.NewReader(f)
	rekaman, err := rd.ReadAll()
	if err != nil || len(rekaman) < 2 {
		fmt.Println("Berkas peta tidak terbaca:", err)
		os.Exit(1)
	}

	var peta []barisPeta
	for _, r := range rekaman[1:] {
		if len(r) < 4 {
			continue
		}
		var imp int
		fmt.Sscanf(r[1], "%d", &imp)
		peta = append(peta, barisPeta{TaskID: r[0], ImpactID: imp, ImpactKod: r[2], PrioKod: r[3]})
	}
	fmt.Printf("\nbaris pemetaan terbaca: %d\n", len(peta))

	// kode prioritas dipetakan ke nomor acuan yang berlaku di basis data ini
	type acuan struct {
		No   int
		Kode string
	}
	var prio []acuan
	db.Raw(`SELECT no, upper(priority) kode FROM task_priorities`).Scan(&prio)
	prioNo := map[string]int{}
	for _, p := range prio {
		prioNo[p.Kode] = p.No
	}
	var damp []acuan
	db.Raw(`SELECT no, upper(impact) kode FROM task_impacts`).Scan(&damp)
	fmt.Print("acuan prioritas:")
	for _, p := range prio {
		fmt.Printf(" %s=%d", p.Kode, p.No)
	}
	fmt.Print("\nacuan dampak   :")
	for _, d := range damp {
		fmt.Printf(" %s=%d", d.Kode, d.No)
	}
	fmt.Println()

	kurang := map[string]bool{}
	for _, b := range peta {
		if _, ada := prioNo[b.PrioKod]; !ada {
			kurang[b.PrioKod] = true
		}
	}
	if len(kurang) > 0 {
		fmt.Println("\nKode prioritas berikut tidak ada pada tabel acuan basis data ini:")
		for k := range kurang {
			fmt.Println("  ", k)
		}
		os.Exit(1)
	}

	// hitung berapa baris yang benar-benar akan berubah
	var adaTugas, akanUbahImpact, akanUbahPrio int64
	for _, b := range peta {
		_ = b
		break
	}
	type hitung struct {
		Ada, BedaImpact, BedaPrio int64
	}
	var h hitung
	for i := 0; i < len(peta); i += 500 {
		j := i + 500
		if j > len(peta) {
			j = len(peta)
		}
		potong := peta[i:j]
		ids := make([]string, 0, len(potong))
		impById := map[string]int{}
		prioById := map[string]int{}
		for _, b := range potong {
			ids = append(ids, b.TaskID)
			impById[b.TaskID] = b.ImpactID
			prioById[b.TaskID] = prioNo[b.PrioKod]
		}
		var baris []struct {
			ID         string
			ImpactID   int
			PriorityID int
		}
		db.Raw(`SELECT id, impact_id, priority_id FROM tasks WHERE id IN (?)`, ids).Scan(&baris)
		for _, r := range baris {
			h.Ada++
			if r.ImpactID != impById[r.ID] {
				h.BedaImpact++
			}
			if r.PriorityID != prioById[r.ID] {
				h.BedaPrio++
			}
		}
	}
	adaTugas, akanUbahImpact, akanUbahPrio = h.Ada, h.BedaImpact, h.BedaPrio

	fmt.Printf("\ntugas yang ditemukan di basis data ini : %d dari %d baris peta\n", adaTugas, len(peta))
	fmt.Printf("dampak yang akan berubah              : %d\n", akanUbahImpact)
	fmt.Printf("prioritas yang akan berubah           : %d\n", akanUbahPrio)

	if !apply {
		fmt.Println("\nPratinjau saja. Tambahkan -apply untuk menyimpan.")
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for i, b := range peta {
			if err := tx.Exec(`UPDATE tasks t SET impact_id = ?, impact = ?, priority_id = ?, priority = ?
				FROM task_priorities tp WHERE t.id = ? AND tp.no = ?`,
				b.ImpactID, b.ImpactKod, prioNo[b.PrioKod], b.PrioKod, b.TaskID, prioNo[b.PrioKod]).Error; err != nil {
				return err
			}
			if (i+1)%2000 == 0 {
				fmt.Printf("  %d baris diproses\n", i+1)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("GAGAL:", err)
		os.Exit(1)
	}
	fmt.Println("\nselesai. Sebaran setelah penyelarasan:")
	diagSingkat(db)
}

func diagSingkat(db *gorm.DB) {
	var baris []struct {
		Priority string
		Impact   string
		N        int64
	}
	db.Raw(`SELECT t.priority, t.impact, COUNT(*) n FROM tasks t GROUP BY 1,2 ORDER BY 3 DESC`).Scan(&baris)
	for _, b := range baris {
		fmt.Printf("  prioritas %-8s dampak %-8s %6d\n", b.Priority, b.Impact, b.N)
	}
}

// ------------------------------------------------------------------ del

// tabel anak yang menunjuk tasks.id, diurutkan agar penghapusan aman
var tabelAnak = []string{
	"task_watchers", "task_tags", "subtasks", "comments", "attachments",
	"activities", "notifications", "task_assignees",
}

func hapusTahun(db *gorm.DB, tahun int, apply bool) {
	var ids []string
	db.Raw(`SELECT id FROM tasks WHERE EXTRACT(YEAR FROM created_at) = ?`, tahun).Scan(&ids)
	fmt.Printf("\ntugas dengan created_at tahun %d: %d\n", tahun, len(ids))
	if len(ids) == 0 {
		return
	}

	var rinci []struct {
		FullName string
		Status   string
		N        int64
	}
	db.Raw(`SELECT COALESCE(u.full_name,'(tanpa pemilik)') full_name, COALESCE(t.status,'') status, COUNT(*) n
		FROM tasks t LEFT JOIN users u ON u.id = t.user_id
		WHERE EXTRACT(YEAR FROM t.created_at) = ? GROUP BY 1,2 ORDER BY 3 DESC`, tahun).Scan(&rinci)
	for _, r := range rinci {
		fmt.Printf("  %-24s %-12s %4d\n", r.FullName, r.Status, r.N)
	}

	// Tabel anak ditentukan dari katalog, bukan dicoba satu per satu. Di
	// Postgres satu perintah yang gagal membatalkan seluruh transaksi, jadi
	// tabel yang tidak ada tidak boleh sampai ikut dieksekusi.
	var adaTabel []string
	db.Raw(`SELECT table_name FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name = 'task_id'
		  AND table_name IN (?)`, tabelAnak).Scan(&adaTabel)

	// baris anak yang ikut terhapus
	for _, t := range adaTabel {
		var n int64
		if err := db.Raw(fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE task_id IN (?)`, t), ids).Scan(&n).Error; err != nil {
			fmt.Printf("  gagal menghitung %s: %v\n", t, err)
			continue
		}
		if n > 0 {
			fmt.Printf("  baris terkait di %-16s %4d\n", t, n)
		}
	}

	if !apply {
		fmt.Println("\nPratinjau saja. Tambahkan -apply untuk menghapus.")
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, t := range adaTabel {
			if err := tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE task_id IN (?)`, t), ids).Error; err != nil {
				return fmt.Errorf("menghapus dari %s: %w", t, err)
			}
		}
		return tx.Exec(`DELETE FROM tasks WHERE id IN (?)`, ids).Error
	})
	if err != nil {
		fmt.Println("GAGAL:", err)
		os.Exit(1)
	}

	var sisa int64
	db.Raw(`SELECT COUNT(*) FROM tasks WHERE EXTRACT(YEAR FROM created_at) = ?`, tahun).Scan(&sisa)
	fmt.Printf("\n%d tugas dihapus, tersisa %d untuk tahun %d.\n", len(ids), sisa, tahun)
}
