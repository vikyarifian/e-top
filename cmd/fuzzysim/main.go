package main

// Perangkat eksperimen untuk Bab 4 Tugas Akhir.
//
//	go run ./cmd/fuzzysim rules     -> validasi pembangkit aturan terhadap services.FuzzyTsukamoto
//	go run ./cmd/fuzzysim manual    -> rincian fuzzifikasi/inferensi/defuzzifikasi kasus uji
//	go run ./cmd/fuzzysim boundary  -> uji nilai batas dan monotonisitas
//	go run ./cmd/fuzzysim data      -> KPI dan skor seluruh agen dari basis data etop2
//	go run ./cmd/fuzzysim mf        -> perbandingan bentuk fungsi keanggotaan
//	go run ./cmd/fuzzysim inputs    -> dampak penambahan jumlah variabel input
//	go run ./cmd/fuzzysim sens      -> analisis sensitivitas tiap variabel
//	go run ./cmd/fuzzysim all       -> seluruh eksperimen, keluaran ke berkas

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"etop/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var out = os.Stdout

func p(format string, a ...any) { fmt.Fprintf(out, format, a...) }

func main() {
	cmd := "all"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	if len(os.Args) > 2 {
		f, err := os.Create(os.Args[2])
		if err != nil {
			fmt.Println("create error:", err)
			return
		}
		defer f.Close()
		out = f
	}

	switch cmd {
	case "rules":
		expRules()
	case "manual":
		expManual()
	case "boundary":
		expBoundary()
	case "data":
		expData()
	case "mf":
		expMF()
	case "inputs":
		expInputs()
	case "sens":
		expSensitivity()
	case "linear":
		expLinear()
	case "tune":
		expTune()
	case "mono":
		expMono()
	case "calib":
		expCalibrated()
	case "all":
		expRules()
		expManual()
		expBoundary()
		expData()
		expMF()
		expInputs()
		expSensitivity()
		expLinear()
		expCalibrated()
	default:
		fmt.Println("perintah tidak dikenal:", cmd)
	}
}

func header(s string) {
	p("\n\n================================================================\n")
	p("%s\n", s)
	p("================================================================\n")
}

// ============================================================
// 1. Validasi pembangkit basis aturan
// ============================================================

func expRules() {
	header("EKSPERIMEN 1 - VALIDASI BASIS ATURAN")
	a := baseline()
	p("Konfigurasi B (rancangan sistem): %d variabel, %d himpunan, %d aturan\n\n", len(a.VarNames), len(a.Sets), len(a.Rules))
	p("%s\n", a.RuleTable())

	p("Sebaran konsekuen pada basis aturan:\n")
	dist := map[string]int{}
	for _, r := range a.Rules {
		dist[r.Output]++
	}
	for _, k := range outLabels {
		p("  %-13s : %2d aturan (%.2f%%)\n", k, dist[k], float64(dist[k])/float64(len(a.Rules))*100)
	}
	p("\n")

	// bandingkan keluaran mesin simulasi terhadap implementasi aplikasi
	maxDiff := 0.0
	beda := 0
	n := 0
	for tcr := 0.0; tcr <= 100; tcr += 5 {
		for otr := 0.0; otr <= 100; otr += 5 {
			for tps := 0.0; tps <= 100; tps += 5 {
				for wer := 0.0; wer <= 100; wer += 5 {
					zApp, cApp, _ := services.FuzzyTsukamoto(tcr, otr, tps, wer)
					zSim, cSim, _, _ := a.Infer([]float64{tcr, otr, tps, wer})
					d := math.Abs(zApp - zSim)
					if d > maxDiff {
						maxDiff = d
					}
					if cApp != cSim {
						beda++
					}
					n++
				}
			}
		}
	}
	p("Verifikasi silang terhadap services.FuzzyTsukamoto pada %d titik uji (langkah 5):\n", n)
	p("  selisih maksimum nilai Z  : %.12f\n", maxDiff)
	p("  jumlah kategori berbeda   : %d\n", beda)
	p("  kesimpulan                : mesin simulasi setara dengan implementasi aplikasi\n")

	p("\nJumlah aturan untuk tiap konfigurasi:\n")
	for _, c := range mfConfigs() {
		p("  %-3s %-55s %d himpunan -> %d aturan\n", c.Name, c.Desc, len(c.Sets), len(c.Rules))
	}
	for _, c := range inputConfigs() {
		p("  %-3s %-55s %d variabel -> %d aturan\n", c.Name, c.Desc, len(c.VarNames), len(c.Rules))
	}
}

// ============================================================
// 2. Rincian perhitungan manual
// ============================================================

type kasus struct {
	Nama string
	X    []float64
}

var kasusUji = []kasus{
	{"U-01 kinerja tinggi seragam", []float64{100, 90, 100, 88}},
	{"U-02 kinerja rendah seragam", []float64{30, 25, 30, 20}},
	{"U-03 seluruh variabel di titik tengah", []float64{50, 50, 50, 50}},
	{"U-04 tuntas namun tidak disiplin tenggat", []float64{100, 40, 100, 51}},
	{"U-05 zona transisi campuran", []float64{100, 52, 100, 57}},
	{"U-06 disiplin namun produktivitas rendah", []float64{35, 95, 40, 90}},
	{"U-07 nilai batas bawah semesta", []float64{0, 0, 0, 0}},
	{"U-08 nilai batas atas semesta", []float64{100, 100, 100, 100}},
}

func expManual() {
	header("EKSPERIMEN 2 - RINCIAN FUZZIFIKASI, INFERENSI, DAN DEFUZZIFIKASI")
	a := baseline()
	for _, k := range kasusUji {
		z, cat, fired, _ := a.Infer(k.X)
		zApp, catApp, _ := services.FuzzyTsukamoto(k.X[0], k.X[1], k.X[2], k.X[3])
		p("\n--- %s ---\n", k.Nama)
		p("Input : TCR=%.2f OTR=%.2f TPS=%.2f WER=%.2f\n", k.X[0], k.X[1], k.X[2], k.X[3])
		p("Fuzzifikasi:\n")
		for i, v := range a.VarNames {
			p("  %-4s = %6.2f  ->  mu_Rendah=%.4f  mu_Sedang=%.4f  mu_Tinggi=%.4f\n",
				v, k.X[i], a.Sets[0].Mu(k.X[i]), a.Sets[1].Mu(k.X[i]), a.Sets[2].Mu(k.X[i]))
		}
		p("Aturan aktif (alpha > 0):\n")
		sumA, sumAZ := 0.0, 0.0
		for _, f := range fired {
			p("  %-4s %-12s alpha=%.4f  z=%7.2f  alpha*z=%8.2f\n", f.Code, f.Output, f.Alpha, f.Z, f.Alpha*f.Z)
			sumA += f.Alpha
			sumAZ += f.Alpha * f.Z
		}
		p("Defuzzifikasi: Z = %.4f / %.4f = %.4f  -> kategori %s\n", sumAZ, sumA, z, cat)
		p("Keluaran aplikasi (services.FuzzyTsukamoto): Z = %.4f -> %s  [selisih %.2e]\n",
			zApp, catApp, math.Abs(z-zApp))
	}
}

// ============================================================
// 3. Uji nilai batas dan monotonisitas
// ============================================================

func expBoundary() {
	header("EKSPERIMEN 3 - UJI NILAI BATAS, MONOTONISITAS, DAN KESINAMBUNGAN")
	a := baseline()

	p("\n3.1 Nilai batas pada input seragam x untuk keempat variabel\n")
	p("%-8s %-10s %-14s %s\n", "x", "Z", "Kategori", "aturan aktif")
	for _, x := range []float64{0, 20, 39.99, 40, 45, 50, 55, 59.99, 60, 65, 70, 75, 80, 85, 89.99, 90, 95, 100} {
		z, c, fired, _ := a.Infer([]float64{x, x, x, x})
		p("%-8.2f %-10.4f %-14s %d\n", x, z, c, len(fired))
	}

	p("\n3.2 Uji monotonisitas: menaikkan satu variabel tidak boleh menurunkan Z\n")
	viol := 0
	tot := 0
	step := 5.0
	for tcr := 0.0; tcr <= 100; tcr += 5 {
		for otr := 0.0; otr <= 100; otr += 5 {
			for tps := 0.0; tps <= 100; tps += 5 {
				for wer := 0.0; wer <= 100; wer += 5 {
					base := []float64{tcr, otr, tps, wer}
					z0 := a.Score(base)
					for i := 0; i < 4; i++ {
						if base[i]+step > 100 {
							continue
						}
						nx := append([]float64{}, base...)
						nx[i] += step
						z1 := a.Score(nx)
						tot++
						if z1 < z0-1e-9 {
							viol++
							if viol <= 5 {
								p("  pelanggaran: %v var=%s  Z %.4f -> %.4f\n", base, a.VarNames[i], z0, z1)
							}
						}
					}
				}
			}
		}
	}
	p("  pasangan diuji = %d, pelanggaran = %d (%.2f%%)\n", tot, viol, float64(viol)/float64(tot)*100)

	p("\n3.3 Uji kesinambungan: lompatan Z maksimum untuk perubahan input 0,01\n")
	maxJump, at := 0.0, []float64{}
	for tcr := 30.0; tcr <= 70; tcr += 2 {
		for otr := 30.0; otr <= 70; otr += 2 {
			for tps := 30.0; tps <= 70; tps += 2 {
				for wer := 30.0; wer <= 70; wer += 2 {
					b := []float64{tcr, otr, tps, wer}
					z0 := a.Score(b)
					for i := 0; i < 4; i++ {
						nx := append([]float64{}, b...)
						nx[i] += 0.01
						d := math.Abs(a.Score(nx) - z0)
						if d > maxJump {
							maxJump = d
							at = append([]float64{}, b...)
						}
					}
				}
			}
		}
	}
	p("  lompatan maksimum = %.6f pada titik %v\n", maxJump, at)
	p("  (pada penilaian tegas berbasis ambang, lompatan pada titik ambang mencapai lebar satu pita kategori, yaitu 20 poin)\n")
}

// ============================================================
// data dari basis data
// ============================================================

type Row struct {
	User   string
	Level  string
	Period string
	All    int64
	Done   int64
	OnTime int64
	TCR    float64
	OTR    float64
	TPS    float64
	WER    float64
	WLR    float64
}

func openDB() *gorm.DB {
	dsn := dsnDariEnv()
	if dsn == "" {
		fmt.Println("DSN basis data belum diatur. Setel env SEED_DSN atau POSTGRES_URL, contoh:")
		fmt.Println("  SEED_DSN=postgres://user:sandi@localhost/etop2")
		os.Exit(1)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		fmt.Println("connect error:", err)
		os.Exit(1)
	}
	return db
}

func loadRows(db *gorm.DB) []Row {
	type UR struct {
		ID    string
		Name  string
		Level string
	}
	var us []UR
	db.Raw(`SELECT u.id, u.full_name AS name, u.level FROM users u
		WHERE EXISTS (SELECT 1 FROM tasks t WHERE t.user_id = u.id)
		ORDER BY u.full_name`).Scan(&us)

	// rata-rata jumlah tugas per karyawan aktif, dipakai sebagai pembagi WLR
	avgAll := map[string]float64{}
	type YA struct {
		Y   string
		Avg float64
	}
	var yas []YA
	db.Raw(`SELECT y::text AS y, AVG(c) AS avg FROM (
		SELECT EXTRACT(YEAR FROM created_at)::int AS y, user_id, COUNT(*) AS c
		FROM tasks GROUP BY 1,2) s GROUP BY y`).Scan(&yas)
	for _, v := range yas {
		avgAll[v.Y] = v.Avg
	}
	var avgTot float64
	db.Raw(`SELECT AVG(c) FROM (SELECT user_id, COUNT(*) AS c FROM tasks GROUP BY 1) s`).Scan(&avgTot)
	avgAll[""] = avgTot

	rows := []Row{}
	for _, u := range us {
		var ys []int
		db.Raw(`SELECT DISTINCT EXTRACT(YEAR FROM created_at)::int FROM tasks WHERE user_id = ? ORDER BY 1`, u.ID).Scan(&ys)
		periods := []string{}
		for _, y := range ys {
			periods = append(periods, fmt.Sprintf("%d", y))
		}
		periods = append(periods, "")
		for _, y := range periods {
			r := evalKPI(db, u.ID, y)
			r.User = u.Name
			r.Level = u.Level
			r.Period = y
			if r.Period == "" {
				r.Period = "GAB"
			}
			if a := avgAll[y]; a > 0 {
				r.WLR = float64(r.All) / a * 100
				if r.WLR > 100 {
					r.WLR = 100
				}
			}
			rows = append(rows, r)
		}
	}
	return rows
}

// replika perhitungan KPI pada services.GetAchievedEvaluation
func evalKPI(db *gorm.DB, userID string, year string) Row {
	ft, fd := "", ""
	if year != "" {
		// periode ditentukan oleh tahun penugasan; kohor yang sama dipakai
		// untuk seluruh indikator agar pembilang tidak melebihi penyebut
		ft = " AND EXTRACT(YEAR FROM t.created_at) = " + year
		fd = ft
	}
	var r Row
	db.Raw(`SELECT COUNT(*) FROM tasks t WHERE t.user_id = ?`+ft, userID).Scan(&r.All)
	db.Raw(`SELECT COUNT(*) FROM tasks t WHERE t.user_id = ? AND t.completed_at IS NOT NULL`+fd, userID).Scan(&r.Done)
	db.Raw(`SELECT COUNT(*) FROM tasks t WHERE t.user_id = ? AND t.completed_at IS NOT NULL AND t.completed_at <= t.due_date`+fd, userID).Scan(&r.OnTime)
	if r.All > 0 {
		r.TCR = float64(r.Done) / float64(r.All) * 100
	}
	if r.Done > 0 {
		r.OTR = float64(r.OnTime) / float64(r.Done) * 100
	}
	var wAll, wDone float64
	db.Raw(`SELECT COALESCE(SUM(tp.level),0) FROM tasks t JOIN task_priorities tp ON tp.no=t.priority_id
		WHERE t.user_id = ?`+ft, userID).Scan(&wAll)
	db.Raw(`SELECT COALESCE(SUM(tp.level),0) FROM tasks t JOIN task_priorities tp ON tp.no=t.priority_id
		WHERE t.user_id = ? AND t.completed_at IS NOT NULL`+fd, userID).Scan(&wDone)
	if wAll > 0 {
		r.TPS = wDone / wAll * 100
	}
	db.Raw(`SELECT COALESCE(AVG((t.estimated_hours/NULLIF(t.actual_hours,0))*100),0) FROM tasks t
		WHERE t.user_id = ? AND t.completed_at IS NOT NULL AND t.estimated_hours>0 AND t.actual_hours>0`+fd, userID).Scan(&r.WER)
	if r.WER > 100 {
		r.WER = 100
	}
	return r
}

var cachedRows []Row

func rows() []Row {
	if cachedRows == nil {
		cachedRows = loadRows(openDB())
	}
	return cachedRows
}

// ============================================================
// 4. Data operasional
// ============================================================

func expData() {
	header("EKSPERIMEN 4 - DATA OPERASIONAL DARI BASIS DATA etop2")
	a := baseline()
	p("%-24s %-6s %6s %6s %7s %7s %7s %7s %7s %8s %-13s %s\n",
		"Karyawan", "Per", "Tugas", "Sels", "TCR", "OTR", "TPS", "WER", "WLR", "Z", "Kategori", "n_aturan")
	for _, r := range rows() {
		z, c, fired, _ := a.Infer([]float64{r.TCR, r.OTR, r.TPS, r.WER})
		p("%-24s %-6s %6d %6d %7.2f %7.2f %7.2f %7.2f %7.2f %8.2f %-13s %d\n",
			trunc(r.User, 24), r.Period, r.All, r.Done, r.TCR, r.OTR, r.TPS, r.WER, r.WLR, z, c, len(fired))
	}

	p("\n4.1 Anomali data yang terdeteksi (TCR atau TPS melebihi 100 persen)\n")
	an := 0
	for _, r := range rows() {
		if r.TCR > 100.0001 || r.TPS > 100.0001 {
			p("  %-24s %-6s TCR=%.2f TPS=%.2f (tugas=%d, selesai=%d)\n", trunc(r.User, 24), r.Period, r.TCR, r.TPS, r.All, r.Done)
			an++
		}
	}
	p("  jumlah baris anomali = %d dari %d baris\n", an, len(rows()))

	p("\n4.2 Sebaran kategori pada periode gabungan\n")
	dist := map[string]int{}
	uniq := map[float64]bool{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		z, c, _, _ := a.Infer([]float64{r.TCR, r.OTR, r.TPS, r.WER})
		dist[c]++
		uniq[math.Round(z*100)/100] = true
	}
	for _, k := range outLabels {
		p("  %-13s : %d\n", k, dist[k])
	}
	p("  nilai Z berbeda = %d\n", len(uniq))
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// ============================================================
// 5. Perbandingan bentuk fungsi keanggotaan
// ============================================================

func expMF() {
	header("EKSPERIMEN 5 - PERBANDINGAN BENTUK FUNGSI KEANGGOTAAN")
	cfgs := mfConfigs()

	p("\n5.1 Nilai Z data operasional pada tiap konfigurasi (periode gabungan)\n")
	p("%-24s", "Karyawan")
	for _, c := range cfgs {
		p(" %10s", c.Name)
	}
	p("   | kategori per konfigurasi\n")
	series := map[string][]float64{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		p("%-24s", trunc(r.User, 24))
		cats := []string{}
		for _, c := range cfgs {
			z, cat, _, _ := c.Infer([]float64{r.TCR, r.OTR, r.TPS, r.WER})
			p(" %10.2f", z)
			cats = append(cats, cat)
			series[c.Name] = append(series[c.Name], z)
		}
		p("   | %s\n", strings.Join(cats, ", "))
	}

	p("\n5.2 Daya pembeda pada data operasional gabungan\n")
	p("%-4s %-58s %8s %8s %8s %8s %6s\n", "Cfg", "Deskripsi", "min", "maks", "rentang", "simpbaku", "unik")
	for _, c := range cfgs {
		v := series[c.Name]
		mn, mx, sum := math.Inf(1), math.Inf(-1), 0.0
		u := map[float64]bool{}
		for _, x := range v {
			mn = math.Min(mn, x)
			mx = math.Max(mx, x)
			sum += x
			u[math.Round(x*100)/100] = true
		}
		mean := sum / float64(len(v))
		vr := 0.0
		for _, x := range v {
			vr += (x - mean) * (x - mean)
		}
		sd := math.Sqrt(vr / float64(len(v)))
		p("%-4s %-58s %8.2f %8.2f %8.2f %8.2f %6d\n", c.Name, trunc(c.Desc, 58), mn, mx, mx-mn, sd, len(u))
	}

	p("\n5.3 Sapuan grid seluruh semesta (0..100 langkah 10, 14641 titik) terhadap baseline A\n")
	grid := [][]float64{}
	for a1 := 0.0; a1 <= 100; a1 += 10 {
		for a2 := 0.0; a2 <= 100; a2 += 10 {
			for a3 := 0.0; a3 <= 100; a3 += 10 {
				for a4 := 0.0; a4 <= 100; a4 += 10 {
					grid = append(grid, []float64{a1, a2, a3, a4})
				}
			}
		}
	}
	base := make([]float64, len(grid))
	for i, g := range grid {
		base[i] = cfgs[0].Score(g)
	}
	p("%-4s %10s %10s %10s %10s %12s %10s %10s\n", "Cfg", "MAE", "RMSE", "maks|d|", "Spearman", "kategori sama", "rerata z", "sd z")
	for _, c := range cfgs {
		var sae, sse, mx float64
		sameCat := 0
		vals := make([]float64, len(grid))
		for i, g := range grid {
			z, cat, _, _ := c.Infer(g)
			vals[i] = z
			d := math.Abs(z - base[i])
			sae += d
			sse += d * d
			mx = math.Max(mx, d)
			if cat == Category(base[i]) {
				sameCat++
			}
		}
		n := float64(len(grid))
		mean := 0.0
		for _, x := range vals {
			mean += x
		}
		mean /= n
		vr := 0.0
		for _, x := range vals {
			vr += (x - mean) * (x - mean)
		}
		p("%-4s %10.3f %10.3f %10.3f %10.4f %11.2f%% %10.2f %10.2f\n",
			c.Name, sae/n, math.Sqrt(sse/n), mx, spearman(vals, base),
			float64(sameCat)/n*100, mean, math.Sqrt(vr/n))
	}

	p("\n5.4 Beban komputasi: jumlah aturan menyala dan waktu inferensi\n")
	p("%-4s %8s %14s %16s %14s\n", "Cfg", "aturan", "rerata aktif", "rerata efektif", "ns/inferensi")
	for _, c := range cfgs {
		totAct, totEff := 0, 0
		for _, g := range grid {
			_, _, fired, eff := c.Infer(g)
			totAct += len(fired)
			totEff += eff
		}
		t0 := time.Now()
		iter := 20000
		for i := 0; i < iter; i++ {
			c.Score(grid[i%len(grid)])
		}
		el := time.Since(t0)
		p("%-4s %8d %14.2f %16.2f %14.0f\n", c.Name, len(c.Rules),
			float64(totAct)/float64(len(grid)), float64(totEff)/float64(len(grid)),
			float64(el.Nanoseconds())/float64(iter))
	}

	p("\n5.5 Profil Z pada input seragam x (memperlihatkan bentuk kurva keluaran)\n")
	p("%-8s", "x")
	for _, c := range cfgs {
		p(" %10s", c.Name)
	}
	p("\n")
	for x := 0.0; x <= 100; x += 5 {
		p("%-8.0f", x)
		for _, c := range cfgs {
			p(" %10.2f", c.Score([]float64{x, x, x, x}))
		}
		p("\n")
	}

	p("\n5.6 Perilaku pada zona jenuh: variasi OTR dan WER saat TCR=TPS=100\n")
	p("%-6s %-6s", "OTR", "WER")
	for _, c := range cfgs {
		p(" %10s", c.Name)
	}
	p("\n")
	for _, v := range [][2]float64{{60, 60}, {70, 70}, {80, 80}, {85, 85}, {88, 87}, {89, 88}, {90, 88}, {92, 90}, {95, 95}, {100, 100}} {
		p("%-6.0f %-6.0f", v[0], v[1])
		for _, c := range cfgs {
			p(" %10.2f", c.Score([]float64{100, v[0], 100, v[1]}))
		}
		p("\n")
	}
}

// ============================================================
// 6. Dampak penambahan jumlah variabel input
// ============================================================

func expInputs() {
	header("EKSPERIMEN 6 - DAMPAK PENAMBAHAN JUMLAH VARIABEL INPUT")
	cfgs := inputConfigs()

	p("\n6.1 Pertumbuhan basis aturan dan beban komputasi\n")
	p("%-4s %-48s %6s %10s %14s %14s\n", "Cfg", "Deskripsi", "n_var", "aturan", "rerata aktif", "ns/inferensi")
	sample := [][]float64{}
	for i := 0; i < 4096; i++ {
		v := []float64{}
		x := i
		for j := 0; j < 5; j++ {
			v = append(v, float64((x%8)*100/7))
			x /= 8
		}
		sample = append(sample, v)
	}
	for _, c := range cfgs {
		nv := len(c.VarNames)
		totAct := 0
		for _, s := range sample {
			_, _, fired, _ := c.Infer(s[:nv])
			totAct += len(fired)
		}
		t0 := time.Now()
		iter := 20000
		for i := 0; i < iter; i++ {
			c.Score(sample[i%len(sample)][:nv])
		}
		el := time.Since(t0)
		p("%-4s %-48s %6d %10d %14.2f %14.0f\n", c.Name, trunc(c.Desc, 48), nv, len(c.Rules),
			float64(totAct)/float64(len(sample)), float64(el.Nanoseconds())/float64(iter))
	}

	p("\n6.2 Nilai Z data operasional pada tiap jumlah variabel (periode gabungan)\n")
	p("%-24s %8s %8s %8s %8s %8s", "Karyawan", "WLR", "N2", "N3", "N4", "N5")
	p("   | kategori N4 -> N5\n")
	series := map[string][]float64{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		full := []float64{r.TCR, r.OTR, r.TPS, r.WER, r.WLR}
		p("%-24s %8.2f", trunc(r.User, 24), r.WLR)
		var c4, c5 string
		for _, c := range cfgs {
			z, cat, _, _ := c.Infer(full[:len(c.VarNames)])
			p(" %8.2f", z)
			series[c.Name] = append(series[c.Name], z)
			if c.Name == "N4" {
				c4 = cat
			}
			if c.Name == "N5" {
				c5 = cat
			}
		}
		p("   | %s -> %s\n", c4, c5)
	}

	p("\n6.3 Daya pembeda dan kestabilan peringkat terhadap N4\n")
	p("%-4s %8s %8s %8s %8s %10s\n", "Cfg", "min", "maks", "rentang", "sd", "Spearman")
	for _, c := range cfgs {
		v := series[c.Name]
		mn, mx, sum := math.Inf(1), math.Inf(-1), 0.0
		for _, x := range v {
			mn = math.Min(mn, x)
			mx = math.Max(mx, x)
			sum += x
		}
		mean := sum / float64(len(v))
		vr := 0.0
		for _, x := range v {
			vr += (x - mean) * (x - mean)
		}
		p("%-4s %8.2f %8.2f %8.2f %8.2f %10.4f\n", c.Name, mn, mx, mx-mn,
			math.Sqrt(vr/float64(len(v))), spearman(v, series["N4"]))
	}

	p("\n6.4 Pengaruh satu variabel bermasalah terhadap nilai akhir (efek pengenceran)\n")
	p("Seluruh variabel lain bernilai 90; satu variabel diturunkan menjadi 20.\n")
	p("%-4s %10s %10s %10s\n", "Cfg", "Z semua 90", "Z satu 20", "penurunan")
	for _, c := range cfgs {
		nv := len(c.VarNames)
		hi := make([]float64, nv)
		for i := range hi {
			hi[i] = 90
		}
		lo := append([]float64{}, hi...)
		lo[nv-1] = 20
		zh, zl := c.Score(hi), c.Score(lo)
		p("%-4s %10.2f %10.2f %10.2f\n", c.Name, zh, zl, zh-zl)
	}
}

// ============================================================
// 7. Analisis sensitivitas
// ============================================================

func expSensitivity() {
	header("EKSPERIMEN 7 - ANALISIS SENSITIVITAS VARIABEL")
	a := baseline()

	p("\n7.1 Perubahan Z ketika satu variabel disapu 0-100, variabel lain tetap 70\n")
	p("%-8s", "x")
	for _, v := range a.VarNames {
		p(" %10s", v)
	}
	p("\n")
	for x := 0.0; x <= 100; x += 10 {
		p("%-8.0f", x)
		for i := range a.VarNames {
			b := []float64{70, 70, 70, 70}
			b[i] = x
			p(" %10.2f", a.Score(b))
		}
		p("\n")
	}

	p("\n7.2 Indeks sensitivitas rata-rata (rerata |dZ/dx| x 10 pada sapuan grid)\n")
	grid := [][]float64{}
	for a1 := 0.0; a1 <= 100; a1 += 20 {
		for a2 := 0.0; a2 <= 100; a2 += 20 {
			for a3 := 0.0; a3 <= 100; a3 += 20 {
				for a4 := 0.0; a4 <= 100; a4 += 20 {
					grid = append(grid, []float64{a1, a2, a3, a4})
				}
			}
		}
	}
	for i, v := range a.VarNames {
		sum, n := 0.0, 0
		for _, g := range grid {
			if g[i]+10 > 100 {
				continue
			}
			b := append([]float64{}, g...)
			z0 := a.Score(b)
			b[i] += 10
			sum += math.Abs(a.Score(b) - z0)
			n++
		}
		p("  %-5s indeks = %.4f poin per kenaikan 10 satuan\n", v, sum/float64(n))
	}

	p("\n7.3 Rentang efektif variabel OTR (nilai di bawah 40 dan di atas 90 tidak lagi mengubah derajat keanggotaan)\n")
	for _, v := range []float64{0, 20, 40, 45, 50, 55, 60, 70, 80, 90, 100} {
		b := []float64{100, v, 100, 100}
		z, c, _, _ := a.Infer(b)
		p("  OTR=%-6.0f -> Z=%7.2f (%s)\n", v, z, c)
	}

	p("\n7.4 Sebaran nilai Z pada seluruh semesta (sapuan langkah 10)\n")
	all := []float64{}
	for a1 := 0.0; a1 <= 100; a1 += 10 {
		for a2 := 0.0; a2 <= 100; a2 += 10 {
			for a3 := 0.0; a3 <= 100; a3 += 10 {
				for a4 := 0.0; a4 <= 100; a4 += 10 {
					all = append(all, a.Score([]float64{a1, a2, a3, a4}))
				}
			}
		}
	}
	sort.Float64s(all)
	q := func(f float64) float64 { return all[int(f*float64(len(all)-1))] }
	p("  n=%d  min=%.2f  Q1=%.2f  median=%.2f  Q3=%.2f  maks=%.2f\n",
		len(all), all[0], q(0.25), q(0.5), q(0.75), all[len(all)-1])
	cnt := map[string]int{}
	for _, z := range all {
		cnt[Category(z)]++
	}
	for _, k := range outLabels {
		p("  %-13s %6d (%.2f%%)\n", k, cnt[k], float64(cnt[k])/float64(len(all))*100)
	}
	sat := 0
	for _, z := range all {
		if z >= 99.999 {
			sat++
		}
	}
	p("  titik dengan Z = 100 (jenuh) : %d (%.2f%%)\n", sat, float64(sat)/float64(len(all))*100)
}

// dsnDariEnv membaca DSN basis data dari lingkungan. Kredensial sengaja tidak
// ditanam di dalam kode agar tidak ikut terpublikasi bersama repositori.
func dsnDariEnv() string {
	if v := os.Getenv("SEED_DSN"); v != "" {
		return v
	}
	return os.Getenv("POSTGRES_URL")
}
