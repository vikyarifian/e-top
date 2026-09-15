package main

// Eksperimen 17: mencari indikator pengganti untuk slot ketiga.
//
// TVS terbukti identik dengan TCR terbobot, sehingga slot ketiga sebenarnya
// tidak menyumbang informasi baru. Eksperimen ini menghitung sekumpulan
// indikator kandidat dari kolom yang benar-benar ada di basis data, lalu
// mengukur dua hal untuk masing-masing:
//
//	daya pembeda : rentang, simpangan baku, dan cacah nilai berbeda
//	kebaruan     : korelasi peringkat terhadap TCR, OTR, dan WER
//
// Kandidat yang baik adalah yang daya pembedanya besar sekaligus korelasinya
// terhadap indikator yang sudah ada rendah.

import (
	"math"
	"sort"
)

// bahanKandidat adalah cacahan mentah satu karyawan yang dibutuhkan seluruh
// kandidat indikator.
type bahanKandidat struct {
	User string

	NAll, NDone, NOnTime, NJudged float64
	WAll, WDone, WOnTime          float64

	// keterlambatan terukur
	NTelat      float64
	SumRasioSLA float64 // jumlah (hari telat / hari jatah) untuk tugas telat
	WTelat      float64
	SumWRasio   float64

	// tugas terbengkalai melewati tenggat
	NLewat, WLewat float64

	// efisiensi
	NJam, SumRasioJam float64

	// keberagaman jenis tugas yang ditangani
	NHigh, NMed, NLow float64
}

func muatBahan() map[string]*bahanKandidat {
	db := openDB()
	var baris []struct {
		User        string
		Priority    string
		NAll        float64
		NDone       float64
		NOnTime     float64
		NJudged     float64
		NTelat      float64
		SumRasioSla float64
		NLewat      float64
		W           float64
		NJam        float64
		SumRasioJam float64
	}
	// Jatah waktu tiap tugas diambil dari max_due_minutes prioritasnya; bila
	// belum diisi, dipakai selisih tanggal mulai ke tenggat.
	db.Raw(`SELECT u.full_name AS user, tp.priority AS priority,
			tp.weight * ti.weight AS w,
			COUNT(*) AS n_all,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL) AS n_done,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL AND t.completed_at <= t.due_date) AS n_on_time,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL
				OR (t.due_date IS NOT NULL AND t.due_date < now())) AS n_judged,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL AND t.completed_at > t.due_date) AS n_telat,
			COALESCE(SUM(
				LEAST(1.0, EXTRACT(EPOCH FROM (t.completed_at - t.due_date)) / 60.0
					/ NULLIF(GREATEST(tp.max_due_minutes,
						EXTRACT(EPOCH FROM (t.due_date - COALESCE(t.start_date, t.created_at)))/60.0), 0))
			) FILTER (WHERE t.completed_at IS NOT NULL AND t.completed_at > t.due_date), 0) AS sum_rasio_sla,
			COUNT(*) FILTER (WHERE t.completed_at IS NULL
				AND t.due_date IS NOT NULL AND t.due_date < now()) AS n_lewat,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL
				AND t.estimated_hours > 0 AND t.actual_hours > 0) AS n_jam,
			COALESCE(SUM((t.estimated_hours / NULLIF(t.actual_hours,0)) * 100)
				FILTER (WHERE t.completed_at IS NOT NULL
					AND t.estimated_hours > 0 AND t.actual_hours > 0), 0) AS sum_rasio_jam
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		JOIN task_priorities tp ON tp.no = t.priority_id
		JOIN task_impacts ti ON ti.no = t.impact_id
		GROUP BY 1,2,3`).Scan(&baris)

	out := map[string]*bahanKandidat{}
	for _, b := range baris {
		k, ada := out[b.User]
		if !ada {
			k = &bahanKandidat{User: b.User}
			out[b.User] = k
		}
		k.NAll += b.NAll
		k.NDone += b.NDone
		k.NOnTime += b.NOnTime
		k.NJudged += b.NJudged
		k.NTelat += b.NTelat
		k.SumRasioSLA += b.SumRasioSla
		k.NLewat += b.NLewat
		k.WAll += b.W * b.NAll
		k.WDone += b.W * b.NDone
		k.WOnTime += b.W * b.NOnTime
		k.WTelat += b.W * b.NTelat
		k.SumWRasio += b.W * b.SumRasioSla
		k.WLewat += b.W * b.NLewat
		k.NJam += b.NJam
		k.SumRasioJam += b.SumRasioJam
		switch b.Priority {
		case "HIGH":
			k.NHigh += b.NAll
		case "MEDIUM":
			k.NMed += b.NAll
		default:
			k.NLow += b.NAll
		}
	}
	return out
}

func bagi(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// kandidat adalah satu indikator yang diuji.
type kandidat struct {
	kode string
	ket  string
	nyal func(k *bahanKandidat, rerataTugas float64) float64
}

func daftarKandidat() []kandidat {
	return []kandidat{
		{"TCR", "penuntasan, cacah selesai / seluruh tugas", func(k *bahanKandidat, _ float64) float64 {
			return bagi(k.NDone, k.NAll) * 100
		}},
		{"OTRlama", "ketepatan waktu, penyebut hanya tugas selesai", func(k *bahanKandidat, _ float64) float64 {
			return bagi(k.NOnTime, k.NDone) * 100
		}},
		{"OTRbaru", "ketepatan waktu, penyebut seluruh tugas yang sudah pasti", func(k *bahanKandidat, _ float64) float64 {
			return bagi(k.NOnTime, k.NJudged) * 100
		}},
		{"TVS", "nilai tugas tuntas, sama dengan TCR terbobot", func(k *bahanKandidat, _ float64) float64 {
			return bagi(k.WDone, k.WAll) * 100
		}},
		{"WER", "efisiensi, rata-rata estimasi terhadap aktual", func(k *bahanKandidat, _ float64) float64 {
			return batas100(bagi(k.SumRasioJam, k.NJam))
		}},
		{"VDS", "Value Delivered on Schedule: nilai tugas tepat waktu / nilai seluruh tugas",
			func(k *bahanKandidat, _ float64) float64 {
				return bagi(k.WOnTime, k.WAll) * 100
			}},
		{"DSS", "Delay Severity Score: 100 dikurangi rata-rata kedalaman keterlambatan",
			func(k *bahanKandidat, _ float64) float64 {
				// tugas telat menyumbang rasio 0..1, tugas terbengkalai dianggap 1
				beban := k.SumRasioSLA + k.NLewat
				return (1 - bagi(beban, k.NJudged)) * 100
			}},
		{"DSSw", "Delay Severity terbobot nilai tugas", func(k *bahanKandidat, _ float64) float64 {
			beban := k.SumWRasio + k.WLewat
			penyebut := k.WDone + k.WLewat
			return (1 - bagi(beban, penyebut)) * 100
		}},
		{"OES", "Overdue Exposure: 100 dikurangi porsi nilai yang terbengkalai lewat tenggat",
			func(k *bahanKandidat, _ float64) float64 {
				return (1 - bagi(k.WLewat, k.WAll)) * 100
			}},
		{"WLR", "Workload Ratio: volume tugas dibanding rata-rata rekan",
			func(k *bahanKandidat, rerata float64) float64 {
				return batas100(bagi(k.NAll, rerata) * 100)
			}},
		{"ATV", "Average Task Value: rata-rata nilai tugas yang ditangani",
			func(k *bahanKandidat, _ float64) float64 {
				return bagi(k.WAll, k.NAll) * 100
			}},
		{"PMX", "Priority Mix: porsi tugas berprioritas HIGH",
			func(k *bahanKandidat, _ float64) float64 {
				return bagi(k.NHigh, k.NAll) * 100
			}},
	}
}

func expKandidat() {
	header("EKSPERIMEN 17 - MENCARI INDIKATOR YANG BENAR-BENAR BERGERAK")

	bahan := muatBahan()
	nama := []string{}
	for n := range bahan {
		nama = append(nama, n)
	}
	sort.Strings(nama)

	var totalTugas float64
	for _, k := range bahan {
		totalTugas += k.NAll
	}
	rerataTugas := bagi(totalTugas, float64(len(bahan)))

	kand := daftarKandidat()
	nilai := map[string][]float64{}

	// ---------------------------------------------------------- 17.1
	p("\n17.1 Nilai tiap kandidat untuk kedelapan karyawan\n")
	p("%-22s", "Karyawan")
	for _, c := range kand {
		p(" %8s", c.kode)
	}
	p("\n")
	for _, n := range nama {
		k := bahan[n]
		p("%-22s", trunc(n, 22))
		for _, c := range kand {
			v := c.nyal(k, rerataTugas)
			nilai[c.kode] = append(nilai[c.kode], v)
			p(" %8.2f", v)
		}
		p("\n")
	}

	// ---------------------------------------------------------- 17.2
	p("\n17.2 Daya pembeda tiap kandidat\n")
	p("%-8s %-56s %8s %8s %8s %6s\n", "Kode", "Keterangan", "min", "maks", "rentang", "sd")
	type skor struct {
		kode    string
		rentang float64
		sd      float64
		unik    int
	}
	var skorS []skor
	for _, c := range kand {
		v := nilai[c.kode]
		mn, mx, sum := math.Inf(1), math.Inf(-1), 0.0
		u := map[float64]bool{}
		for _, x := range v {
			mn, mx = math.Min(mn, x), math.Max(mx, x)
			sum += x
			u[math.Round(x*100)/100] = true
		}
		rata := sum / float64(len(v))
		vr := 0.0
		for _, x := range v {
			vr += (x - rata) * (x - rata)
		}
		sd := math.Sqrt(vr / float64(len(v)))
		p("%-8s %-56s %8.2f %8.2f %8.2f %6.2f\n", c.kode, trunc(c.ket, 56), mn, mx, mx-mn, sd)
		skorS = append(skorS, skor{c.kode, mx - mn, sd, len(u)})
	}

	// ---------------------------------------------------------- 17.3
	p("\n17.3 Kebaruan: korelasi peringkat terhadap indikator yang sudah ada\n")
	p("%-8s %10s %10s %10s %10s %8s\n", "Kode", "vs TCR", "vs OTRbaru", "vs WER", "maks|r|", "unik")
	type baris struct {
		kode    string
		maksR   float64
		rentang float64
		sd      float64
		unik    int
	}
	var ringkas []baris
	for i, c := range kand {
		v := nilai[c.kode]
		rT := spearman(v, nilai["TCR"])
		rO := spearman(v, nilai["OTRbaru"])
		rW := spearman(v, nilai["WER"])
		maks := math.Max(math.Abs(rT), math.Max(math.Abs(rO), math.Abs(rW)))
		if c.kode == "TCR" || c.kode == "OTRbaru" || c.kode == "WER" {
			p("%-8s %10.4f %10.4f %10.4f %10s %8d\n", c.kode, rT, rO, rW, "(acuan)", skorS[i].unik)
			continue
		}
		p("%-8s %10.4f %10.4f %10.4f %10.4f %8d\n", c.kode, rT, rO, rW, maks, skorS[i].unik)
		ringkas = append(ringkas, baris{c.kode, maks, skorS[i].rentang, skorS[i].sd, skorS[i].unik})
	}

	// ---------------------------------------------------------- 17.4
	p("\n17.4 Peringkat kandidat: bergerak banyak sekaligus tidak mengulang yang ada\n")
	p("Skor = rentang dibagi 100, dikali (1 dikurangi korelasi tertinggi).\n")
	p("Semakin besar semakin layak menempati slot ketiga.\n\n")
	sort.Slice(ringkas, func(i, j int) bool {
		si := ringkas[i].rentang / 100 * (1 - ringkas[i].maksR)
		sj := ringkas[j].rentang / 100 * (1 - ringkas[j].maksR)
		return si > sj
	})
	p("%-8s %10s %10s %8s %10s\n", "Kode", "rentang", "maks|r|", "unik", "skor")
	for _, b := range ringkas {
		p("%-8s %10.2f %10.4f %8d %10.4f\n",
			b.kode, b.rentang, b.maksR, b.unik, b.rentang/100*(1-b.maksR))
	}

	// ---------------------------------------------------------- 17.5
	p("\n17.5 Pengaruh perubahan OTR terhadap hasil\n")
	p("%-22s %10s %10s %10s\n", "Karyawan", "OTR lama", "OTR baru", "selisih")
	for i, n := range nama {
		p("%-22s %10.2f %10.2f %+10.2f\n", trunc(n, 22),
			nilai["OTRlama"][i], nilai["OTRbaru"][i], nilai["OTRbaru"][i]-nilai["OTRlama"][i])
	}
	p("\nkorelasi peringkat OTR lama terhadap OTR baru: %.4f ; rerata selisih %.2f poin\n",
		spearman(nilai["OTRlama"], nilai["OTRbaru"]), rerataBeda(nilai["OTRlama"], nilai["OTRbaru"]))

	// ---------------------------------------------------------- 17.6
	p("\n17.6 Nilai akhir bila slot ketiga diganti tiga kandidat teratas\n")
	cfg := cfgVarian(varianKaidah()[1]) // rancangan sistem: agregasi saja
	// perbandingan dipilih tangan: TVS yang berlaku sekarang, dua kandidat
	// berbasis kinerja, dan satu kandidat berbasis penugasan sebagai pembanding
	pilih := []string{"TVS", "VDS", "DSS", "DSSw", "WLR"}
	_ = ringkas
	p("%-22s", "Karyawan")
	for _, kd := range pilih {
		p(" %9s %-13s", kd, "kategori")
	}
	p("\n")
	seriZ := map[string][]float64{}
	for i, n := range nama {
		p("%-22s", trunc(n, 22))
		for _, kd := range pilih {
			z, c, _, _ := cfg.Infer([]float64{
				nilai["TCR"][i], nilai["OTRbaru"][i], nilai[kd][i], nilai["WER"][i]})
			seriZ[kd] = append(seriZ[kd], z)
			p(" %9.2f %-13s", z, trunc(c, 13))
		}
		p("\n")
	}
	p("\n%-10s %10s %10s %8s\n", "slot ke-3", "rentang Z", "sd Z", "unik")
	for _, kd := range pilih {
		v := seriZ[kd]
		mn, mx, sum := math.Inf(1), math.Inf(-1), 0.0
		u := map[float64]bool{}
		for _, x := range v {
			mn, mx = math.Min(mn, x), math.Max(mx, x)
			sum += x
			u[math.Round(x*100)/100] = true
		}
		rata := sum / float64(len(v))
		vr := 0.0
		for _, x := range v {
			vr += (x - rata) * (x - rata)
		}
		p("%-10s %10.2f %10.2f %8d\n", kd, mx-mn, math.Sqrt(vr/float64(len(v))), len(u))
	}
}
