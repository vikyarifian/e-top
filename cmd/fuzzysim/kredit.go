package main

// Eksperimen 14: di tabel mana kredit keterlambatan sebaiknya dipasang?
//
// Rumus TVS usulan memberi tugas yang selesai terlambat sebagian nilai saja:
//
//	TVS = SUM( w_prioritas * w_dampak * k ) / SUM( w_prioritas * w_dampak )
//
//	k = 0            bila tugas belum selesai
//	k = 1            bila selesai tepat waktu
//	k = kredit telat bila selesai terlambat
//
// Pertanyaannya: kredit telat itu diambil dari tabel prioritas saja, dari tabel
// dampak saja, atau dari keduanya? Eksperimen ini membandingkan empat pilihan
// pada data yang sama.

import (
	"math"
	"sort"
)

// skemaKredit menentukan dari mana nilai kredit telat diambil.
type skemaKredit struct {
	nama  string
	ket   string
	prio  map[string]float64 // nil berarti dimensi ini tidak dipakai
	impak map[string]float64
	kali  bool // true: kedua kredit dikalikan; false: diambil yang terkecil
}

// kredit mengembalikan kredit telat untuk satu kombinasi prioritas dan dampak.
func (s skemaKredit) kredit(pr, im string) float64 {
	cp, adaP := s.prio[pr]
	ci, adaI := s.impak[im]
	switch {
	case adaP && adaI && s.kali:
		return cp * ci
	case adaP && adaI:
		return math.Min(cp, ci)
	case adaP:
		return cp
	case adaI:
		return ci
	default:
		return 0
	}
}

func skemaKreditUji() []skemaKredit {
	// Kredit menurun seiring naiknya tingkat: makin mendesak atau makin luas
	// dampaknya, makin sedikit nilai yang tersisa bila terlambat.
	tiga := func(a, b, c float64) map[string]float64 {
		return map[string]float64{"HIGH": a, "MEDIUM": b, "LOW": c}
	}
	rata := tiga(0.5, 0.5, 0.5)
	turun := tiga(0.25, 0.5, 0.75)
	return []skemaKredit{
		{"A-rata", "satu nilai untuk semua, 0,50", rata, nil, false},
		{"B-prio", "kredit dari tabel prioritas saja", turun, nil, false},
		{"C-dampak", "kredit dari tabel dampak saja", nil, turun, false},
		{"D-kali", "kedua tabel, kredit dikalikan", turun, turun, true},
		{"E-min", "kedua tabel, diambil yang terkecil", turun, turun, false},
	}
}

// tvsKredit menghitung TVS satu karyawan menurut skema bobot dan skema kredit.
func tvsKredit(ps []petak, w skemaBobot, k skemaKredit) float64 {
	var atas, bawah float64
	for _, q := range ps {
		nilai := w.nilai(q.Priority, q.Impact)
		telat := q.NDone - q.NOnTime
		bawah += nilai * float64(q.NAll)
		atas += nilai * (float64(q.NOnTime) + k.kredit(q.Priority, q.Impact)*float64(telat))
	}
	if bawah == 0 {
		return 0
	}
	return atas / bawah * 100
}

func expKredit() {
	header("EKSPERIMEN 14 - DI TABEL MANA KREDIT KETERLAMBATAN DIPASANG")

	petakUser := muatPetak()
	w := skemaUji()[1] // bobot yang dipakai sistem: 1 / 0,9 / 0,8
	cfg := cfgVarian(varianKaidah()[0])

	type kpi struct{ TCR, OTR, WER float64 }
	lain := map[string]kpi{}
	nama := []string{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		lain[r.User] = kpi{r.TCR, r.OTR, r.WER}
		nama = append(nama, r.User)
	}
	sort.Strings(nama)

	tingkat := []string{"HIGH", "MEDIUM", "LOW"}

	// ---------------------------------------------------------- 14.1
	p("\n14.1 Kredit telat yang dihasilkan tiap skema, per kombinasi\n")
	p("%-10s %-36s", "Skema", "Keterangan")
	for _, pr := range tingkat {
		for _, im := range tingkat {
			p(" %s/%s", trunc(pr, 4), trunc(im, 4))
		}
	}
	p("\n")
	for _, k := range skemaKreditUji() {
		p("%-10s %-36s", k.nama, trunc(k.ket, 36))
		for _, pr := range tingkat {
			for _, im := range tingkat {
				p(" %9.3f", k.kredit(pr, im))
			}
		}
		p("\n")
	}
	p("\nPerhatikan skema D: kedua kredit 0,50 yang dikalikan menghasilkan 0,25.\n")
	p("Nilai yang disetel pengelola bukan nilai yang berlaku.\n")

	// ---------------------------------------------------------- 14.2
	p("\n14.2 Dampak sudah bekerja lewat bobot, bahkan tanpa kredit sendiri\n")
	p("Nilai yang hilang bila satu tugas terlambat, memakai skema B (kredit prioritas saja).\n")
	p("hilang = w_prioritas x w_dampak x (1 - kredit)\n\n")
	p("%-10s %-10s %10s %10s %10s %12s\n",
		"Prioritas", "Dampak", "w_prio", "w_dampak", "kredit", "nilai hilang")
	kB := skemaKreditUji()[1]
	for _, pr := range tingkat {
		for _, im := range tingkat {
			wp := w.prio[pr]
			wi := w.impak[im]
			c := kB.kredit(pr, im)
			p("%-10s %-10s %10.3f %10.3f %10.3f %12.4f\n",
				pr, im, wp, wi, c, wp*wi*(1-c))
		}
	}
	p("\nPada prioritas yang sama, tugas berdampak HIGH kehilangan %.1f persen lebih\n",
		(w.impak["HIGH"]/w.impak["LOW"]-1)*100)
	p("banyak daripada yang berdampak LOW. Pembedaan itu sudah ada tanpa perlu\n")
	p("kolom kredit tersendiri di tabel dampak.\n")

	// ---------------------------------------------------------- 14.3
	p("\n14.3 Nilai TVS kedelapan karyawan menurut tiap skema\n")
	p("%-22s %8s", "Karyawan", "OTR")
	for _, k := range skemaKreditUji() {
		p(" %10s", k.nama)
	}
	p(" %10s\n", "rentang")
	rentangMaks := 0.0
	nilai := map[string]map[string]float64{}
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		nilai[n] = map[string]float64{}
		p("%-22s %8.2f", trunc(n, 22), lain[n].OTR)
		mn, mx := math.Inf(1), math.Inf(-1)
		for _, k := range skemaKreditUji() {
			v := tvsKredit(ps, w, k)
			nilai[n][k.nama] = v
			mn, mx = math.Min(mn, v), math.Max(mx, v)
			p(" %10.4f", v)
		}
		p(" %10.4f\n", mx-mn)
		rentangMaks = math.Max(rentangMaks, mx-mn)
	}
	p("\nRentang terbesar antar-skema: %.4f poin.\n", rentangMaks)

	// ---------------------------------------------------------- 14.4
	p("\n14.4 Selisih skema yang memakai dampak terhadap skema prioritas saja\n")
	p("%-22s %12s %12s %12s\n", "Karyawan", "C-dampak", "D-kali", "E-min")
	var sumC, sumD, sumE float64
	var cacah int
	for _, n := range nama {
		if len(petakUser[n]) == 0 {
			continue
		}
		b := nilai[n]["B-prio"]
		p("%-22s %+12.4f %+12.4f %+12.4f\n", trunc(n, 22),
			nilai[n]["C-dampak"]-b, nilai[n]["D-kali"]-b, nilai[n]["E-min"]-b)
		sumC += math.Abs(nilai[n]["C-dampak"] - b)
		sumD += math.Abs(nilai[n]["D-kali"] - b)
		sumE += math.Abs(nilai[n]["E-min"] - b)
		cacah++
	}
	if cacah > 0 {
		p("\nselisih mutlak rata-rata terhadap B: C=%.4f  D=%.4f  E=%.4f poin\n",
			sumC/float64(cacah), sumD/float64(cacah), sumE/float64(cacah))
	}

	// ---------------------------------------------------------- 14.5
	p("\n14.5 Kategori akhir menurut tiap skema\n")
	p("%-22s", "Karyawan")
	for _, k := range skemaKreditUji() {
		p(" %-14s", k.nama)
	}
	p("\n")
	bedaKat := map[string]int{}
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		k0 := lain[n]
		p("%-22s", trunc(n, 22))
		var acuan string
		for i, k := range skemaKreditUji() {
			_, c, _, _ := cfg.Infer([]float64{k0.TCR, k0.OTR, nilai[n][k.nama], k0.WER})
			if i == 1 {
				acuan = c
			}
			p(" %-14s", trunc(c, 14))
		}
		for _, k := range skemaKreditUji() {
			_, c, _, _ := cfg.Infer([]float64{k0.TCR, k0.OTR, nilai[n][k.nama], k0.WER})
			if c != acuan {
				bedaKat[k.nama]++
			}
		}
		p("\n")
	}
	p("\njumlah karyawan yang kategorinya berbeda dari skema B:\n")
	for _, k := range skemaKreditUji() {
		p("  %-10s %d dari %d\n", k.nama, bedaKat[k.nama], cacah)
	}

	// ---------------------------------------------------------- 14.6
	p("\n14.6 Seberapa banyak data yang benar-benar tersentuh dimensi dampak\n")
	sebaran := map[string]int{}
	telatPer := map[string]int{}
	total, totalTelat := 0, 0
	for _, ps := range petakUser {
		for _, q := range ps {
			sebaran[q.Impact] += q.NAll
			total += q.NAll
			t := q.NDone - q.NOnTime
			telatPer[q.Impact] += t
			totalTelat += t
		}
	}
	p("%-10s %10s %10s %12s %12s\n", "Dampak", "tugas", "persen", "telat", "persen telat")
	for _, im := range tingkat {
		pt, pl := 0.0, 0.0
		if total > 0 {
			pt = float64(sebaran[im]) / float64(total) * 100
		}
		if totalTelat > 0 {
			pl = float64(telatPer[im]) / float64(totalTelat) * 100
		}
		p("%-10s %10d %9.2f%% %12d %11.2f%%\n", im, sebaran[im], pt, telatPer[im], pl)
	}
	p("\nKolom terakhir menentukan segalanya: kredit pada tabel dampak hanya dapat\n")
	p("mengubah nasib tugas telat yang dampaknya bukan HIGH.\n")
}
