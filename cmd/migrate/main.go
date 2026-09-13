// Perintah migrate menjalankan berkas SQL pada db/migrations secara berurutan.
// Proyek ini tidak memakai AutoMigrate dan tidak mengandalkan psql, sehingga
// perubahan skema dijalankan lewat perintah ini.
//
// Pemakaian:
//
//	go run ./cmd/migrate            menjalankan migrasi yang belum diterapkan
//	go run ./cmd/migrate -status    hanya menampilkan status, tanpa mengubah apa pun
//	go run ./cmd/migrate -dsn "postgres://..."
//
// Basis data diambil dari tanda -dsn, atau dari env POSTGRES_URL atau SEED_DSN.
// Setiap berkas yang berhasil dijalankan dicatat pada tabel schema_migrations
// sehingga tidak akan dijalankan dua kali.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const direktori = "db/migrations"

func main() {
	dsn := flag.String("dsn", "", "alamat basis data, menimpa env POSTGRES_URL")
	status := flag.Bool("status", false, "hanya tampilkan status migrasi")
	flag.Parse()

	alamat := *dsn
	if alamat == "" {
		alamat = os.Getenv("POSTGRES_URL")
	}
	if alamat == "" {
		alamat = os.Getenv("SEED_DSN")
	}
	if alamat == "" {
		alamat = bacaEnvLokal()
	}
	if alamat == "" {
		fmt.Println("Alamat basis data belum diatur. Pakai -dsn, atau setel env POSTGRES_URL.")
		fmt.Println("  go run ./cmd/migrate -dsn \"postgres://pengguna:sandi@host/etop2\"")
		os.Exit(1)
	}

	// Protokol sederhana dipakai agar satu berkas yang memuat banyak perintah,
	// termasuk blok DO, dapat dijalankan sebagai satu kesatuan.
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  alamat,
		PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("Gagal tersambung ke basis data:", err)
		os.Exit(1)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    text PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`).Error; err != nil {
		fmt.Println("Gagal menyiapkan tabel schema_migrations:", err)
		os.Exit(1)
	}

	berkas, err := filepath.Glob(filepath.Join(direktori, "*.sql"))
	if err != nil || len(berkas) == 0 {
		fmt.Printf("Tidak ada berkas migrasi pada %s\n", direktori)
		os.Exit(1)
	}
	sort.Strings(berkas)

	sudah := map[string]time.Time{}
	var baris []struct {
		Version   string
		AppliedAt time.Time
	}
	db.Raw(`SELECT version, applied_at FROM schema_migrations`).Scan(&baris)
	for _, b := range baris {
		sudah[b.Version] = b.AppliedAt
	}

	if *status {
		fmt.Printf("%-36s %s\n", "MIGRASI", "STATUS")
		for _, f := range berkas {
			v := filepath.Base(f)
			if t, ada := sudah[v]; ada {
				fmt.Printf("%-36s diterapkan %s\n", v, t.Format("2006-01-02 15:04"))
			} else {
				fmt.Printf("%-36s belum\n", v)
			}
		}
		return
	}

	dijalankan := 0
	for _, f := range berkas {
		v := filepath.Base(f)
		if _, ada := sudah[v]; ada {
			fmt.Printf("  lewati  %s (sudah diterapkan)\n", v)
			continue
		}
		isi, err := os.ReadFile(f)
		if err != nil {
			fmt.Println("Gagal membaca", f, ":", err)
			os.Exit(1)
		}
		if strings.TrimSpace(string(isi)) == "" {
			continue
		}
		fmt.Printf("  jalankan %s ...", v)
		if err := db.Exec(string(isi)).Error; err != nil {
			fmt.Println(" GAGAL")
			fmt.Println("   ", err)
			fmt.Println("Migrasi dihentikan. Berkas ini memakai BEGIN dan COMMIT sendiri,")
			fmt.Println("sehingga perubahan di dalamnya tidak tersimpan sebagian.")
			os.Exit(1)
		}
		if err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, v).Error; err != nil {
			fmt.Println(" GAGAL mencatat versi")
			fmt.Println("   ", err)
			os.Exit(1)
		}
		fmt.Println(" selesai")
		dijalankan++
	}

	if dijalankan == 0 {
		fmt.Println("Tidak ada migrasi baru. Skema sudah mutakhir.")
	} else {
		fmt.Printf("%d migrasi diterapkan.\n", dijalankan)
	}
}

// bacaEnvLokal mengambil POSTGRES_URL dari .env.local bila ada, agar perintah
// ini dapat dijalankan tanpa menyetel env terlebih dahulu.
func bacaEnvLokal() string {
	isi, err := os.ReadFile(".env.local")
	if err != nil {
		return ""
	}
	for _, baris := range strings.Split(string(isi), "\n") {
		baris = strings.TrimSpace(baris)
		if !strings.HasPrefix(baris, "POSTGRES_URL") {
			continue
		}
		bagian := strings.SplitN(baris, "=", 2)
		if len(bagian) != 2 {
			continue
		}
		return strings.Trim(strings.TrimSpace(bagian[1]), `"'`)
	}
	return ""
}
