package main

// Eksperimen 13: dapatkah kaidah pembatas OTR digantikan oleh penyetelan bobot
// pada task_priorities dan task_impacts?
//
// Kaidah pembatas bekerja pada konsekuen aturan: bila OTR berada pada himpunan
// terendah, kategori dibatasi maksimum "Cukup". Bobot prioritas dan dampak
// bekerja di tempat yang sama sekali berbeda, yaitu pada masukan TVS:
//
//	TVS = jumlah nilai tugas yang selesai / jumlah nilai seluruh tugas
//
// Eksperimen ini mengukur seberapa jauh bobot dapat menggerakkan TVS, lalu
// menguji satu usulan jalan keluar: memasukkan ketepatan waktu ke dalam
// pembilang TVS sehingga sifat tidak-saling-menutup muncul dari data, bukan
// dari pengecualian di dalam kode.

import (
	"fmt"
	"math"
	"sort"
)

// petak adalah cacah tugas satu karyawan pada satu kombinasi prioritas dan dampak.
type petak struct {
	User     string
	Priority string
	Impact   string
	NAll     int
	NDone    int
	NOnTime  int
}

// skemaBobot adalah satu penyetelan bobot pada kedua tabel acuan.
type skemaBobot struct {
	nama  string
	prio  map[string]float64
	impak map[string]float64
}

func skemaUji() []skemaBobot {
	tiga := func(a, b, c float64) map[string]float64 {
		return map[string]float64{"HIGH": a, "MEDIUM": b, "LOW": c}
	}
	return []skemaBobot{
		{"datar", tiga(1, 1, 1), tiga(1, 1, 1)},
		{"sistem", tiga(1, 0.9, 0.8), tiga(1, 0.9, 0.8)},
		{"lebar", tiga(1, 0.5, 0.1), tiga(1, 0.5, 0.1)},
		{"ekstrem", tiga(1, 0.1, 0.01), tiga(1, 0.1, 0.01)},
		{"terbalik", tiga(0.01, 0.1, 1), tiga(0.01, 0.1, 1)},
	}
}

func (s skemaBobot) nilai(p, i string) float64 {
	wp, ada := s.prio[p]
	if !ada {
		wp = 1
	}
	wi, ada := s.impak[i]
	if !ada {
		wi = 1
	}
	return wp * wi
}

// jumlahBobot mengembalikan total nilai seluruh tugas, tugas selesai, dan
// tugas selesai tepat waktu bagi satu karyawan menurut satu skema bobot.
func jumlahBobot(ps []petak, s skemaBobot) (wAll, wDone, wOnTime float64) {
	for _, q := range ps {
		w := s.nilai(q.Priority, q.Impact)
		wAll += w * float64(q.NAll)
		wDone += w * float64(q.NDone)
		wOnTime += w * float64(q.NOnTime)
	}
	return
}

func muatPetak() map[string][]petak {
	db := openDB()
	var baris []petak
	db.Raw(`SELECT u.full_name AS user, tp.priority AS priority, ti.impact AS impact,
			COUNT(*) AS n_all,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL) AS n_done,
			COUNT(*) FILTER (WHERE t.completed_at IS NOT NULL AND t.completed_at <= t.due_date) AS n_on_time
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		JOIN task_priorities tp ON tp.no = t.priority_id
		JOIN task_impacts ti ON ti.no = t.impact_id
		GROUP BY 1,2,3`).Scan(&baris)
	out := map[string][]petak{}
	for _, b := range baris {
		out[b.User] = append(out[b.User], b)
	}
	return out
}

func expBobot() {
	header("EKSPERIMEN 13 - DAPATKAH BOBOT MENGGANTIKAN KAIDAH PEMBATAS OTR")

	petakUser := muatPetak()
	dasar := cfgVarian(varianKaidah()[0])      // agregasi + pembatas (rancangan lama)
	tanpaBatas := cfgVarian(varianKaidah()[1]) // agregasi saja (rancangan sekarang)

	// indikator lain diambil dari data operasional gabungan
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

	// ---------------------------------------------------------- 13.1
	p("\n13.1 Seberapa jauh bobot dapat menggerakkan TVS\n")
	p("Lima skema bobot yang sangat berbeda diterapkan pada data yang sama.\n\n")
	p("%-22s %8s", "Karyawan", "TCR")
	for _, s := range skemaUji() {
		p(" %10s", s.nama)
	}
	p(" %10s\n", "rentang")
	maksRentangLama := 0.0
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		p("%-22s %8.2f", trunc(n, 22), lain[n].TCR)
		mn, mx := math.Inf(1), math.Inf(-1)
		for _, s := range skemaUji() {
			wa, wd, _ := jumlahBobot(ps, s)
			v := 0.0
			if wa > 0 {
				v = wd / wa * 100
			}
			mn, mx = math.Min(mn, v), math.Max(mx, v)
			p(" %10.4f", v)
		}
		p(" %10.4f\n", mx-mn)
		maksRentangLama = math.Max(maksRentangLama, mx-mn)
	}
	p("\nRentang terbesar pada seluruh karyawan: %.4f poin.\n", maksRentangLama)
	p("Sebabnya aljabar, bukan pilihan angka: bila seluruh tugas seseorang selesai,\n")
	p("pembilang dan penyebut TVS memuat himpunan yang sama, sehingga rasionya\n")
	p("tepat 1 untuk bobot positif apa pun.\n")

	// ---------------------------------------------------------- 13.2
	p("\n13.2 Berapa nilai TVS yang dibutuhkan agar pembatas tidak diperlukan\n")
	p("Titik uji: TCR=100, OTR=0, WER=100, TVS disapu. Tanpa kaidah pembatas.\n\n")
	p("%-8s %10s %-14s %10s %-14s\n", "TVS", "Z tanpa", "kategori", "Z dengan", "kategori")
	butuh := -1.0
	for tvs := 100.0; tvs >= 0; tvs -= 10 {
		z2, c2, _, _ := tanpaBatas.Infer([]float64{100, 0, tvs, 100})
		z1, c1, _, _ := dasar.Infer([]float64{100, 0, tvs, 100})
		p("%-8.0f %10.2f %-14s %10.2f %-14s\n", tvs, z2, c2, z1, c1)
		if butuh < 0 && z2 <= 60.0001 {
			butuh = tvs
		}
	}
	if butuh >= 0 {
		p("\nTVS harus turun sampai sekitar %.0f agar hasilnya setara dengan pembatas.\n", butuh)
	} else {
		p("\nTidak ada nilai TVS yang menyamai hasil pembatas pada titik ini.\n")
	}

	// ---------------------------------------------------------- 13.3
	p("\n13.3 Usulan jalan keluar: ketepatan waktu masuk ke pembilang TVS\n")
	p("TVS(lambda) = [nilai tugas tepat waktu + lambda x nilai tugas telat] / nilai seluruh tugas\n")
	p("lambda=1 sama dengan TVS sekarang; lambda=0 berarti tugas telat tidak dihitung.\n\n")
	lambdas := []float64{1, 0.75, 0.5, 0.25, 0}
	s := skemaUji()[1] // skema bobot yang dipakai sistem
	p("%-22s %8s %8s", "Karyawan", "OTR", "TCR")
	for _, l := range lambdas {
		p("   L=%.2f", l)
	}
	p("\n")
	tvsL := map[string]map[float64]float64{}
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		wa, wd, wo := jumlahBobot(ps, s)
		tvsL[n] = map[float64]float64{}
		p("%-22s %8.2f %8.2f", trunc(n, 22), lain[n].OTR, lain[n].TCR)
		for _, l := range lambdas {
			v := 0.0
			if wa > 0 {
				v = (wo + l*(wd-wo)) / wa * 100
			}
			tvsL[n][l] = v
			p(" %7.2f", v)
		}
		p("\n")
	}

	// ---------------------------------------------------------- 13.4
	p("\n13.4 Hasil akhir bila pembatas dihapus dan TVS memakai lambda=0\n")
	p("%-22s %-28s %-28s %s\n", "Karyawan",
		"sistem sekarang (ada pembatas)", "usulan (tanpa pembatas, L=0)", "sama?")
	sama, total := 0, 0
	var zLama, zBaru []float64
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		k := lain[n]
		z1, c1, _, _ := dasar.Infer([]float64{k.TCR, k.OTR, tvsL[n][1], k.WER})
		z2, c2, _, _ := tanpaBatas.Infer([]float64{k.TCR, k.OTR, tvsL[n][0], k.WER})
		tanda := "ya"
		if c1 != c2 {
			tanda = "TIDAK"
		} else {
			sama++
		}
		total++
		zLama = append(zLama, z1)
		zBaru = append(zBaru, z2)
		p("%-22s %8.2f %-19s %8.2f %-19s %s\n", trunc(n, 22), z1, c1, z2, c2, tanda)
	}
	p("\nkategori yang sepakat: %d dari %d ; korelasi peringkat Spearman: %.4f\n",
		sama, total, spearman(zLama, zBaru))

	// ---------------------------------------------------------- 13.5
	p("\n13.5 Titik kritis: karyawan sempurna yang tidak pernah tepat waktu\n")
	p("Seluruh tugas selesai, tidak satu pun tepat waktu.\n\n")
	p("%-46s %8s %-14s\n", "Rancangan", "Z", "kategori")
	uji := []struct {
		ket string
		cfg *Config
		x   []float64
	}{
		{"sistem sekarang: TVS=100, ada pembatas", dasar, []float64{100, 0, 100, 100}},
		{"pembatas dihapus, TVS tetap 100", tanpaBatas, []float64{100, 0, 100, 100}},
		{"pembatas dihapus, TVS lambda=0 sehingga 0", tanpaBatas, []float64{100, 0, 0, 100}},
		{"pembatas dihapus, TVS lambda=0,5 sehingga 50", tanpaBatas, []float64{100, 0, 50, 100}},
	}
	for _, u := range uji {
		z, c, _, _ := u.cfg.Infer(u.x)
		p("%-46s %8.2f %-14s\n", u.ket, z, c)
	}

	// ---------------------------------------------------------- 13.6
	p("\n13.6 Apakah TVS usulan menjadi kembar dengan indikator lain\n")
	var vOTR, vTCR, vT0, vT1 []float64
	for _, n := range nama {
		if len(petakUser[n]) == 0 {
			continue
		}
		vOTR = append(vOTR, lain[n].OTR)
		vTCR = append(vTCR, lain[n].TCR)
		vT0 = append(vT0, tvsL[n][0])
		vT1 = append(vT1, tvsL[n][1])
	}
	p("  TVS sekarang (L=1) terhadap TCR : Spearman %.4f\n", spearman(vT1, vTCR))
	p("  TVS usulan   (L=0) terhadap TCR : Spearman %.4f\n", spearman(vT0, vTCR))
	p("  TVS usulan   (L=0) terhadap OTR : Spearman %.4f\n", spearman(vT0, vOTR))
	p("\n  selisih mutlak rata-rata terhadap indikator lain:\n")
	p("    TVS sekarang vs TCR : %.2f poin\n", rerataBeda(vT1, vTCR))
	p("    TVS usulan   vs TCR : %.2f poin\n", rerataBeda(vT0, vTCR))
	p("    TVS usulan   vs OTR : %.2f poin\n", rerataBeda(vT0, vOTR))
	p("\n  selisih TVS usulan terhadap OTR per karyawan; positif berarti keterlambatan\n")
	p("  lebih banyak terjadi pada tugas bernilai rendah, negatif sebaliknya:\n")
	for _, n := range nama {
		if len(petakUser[n]) == 0 {
			continue
		}
		p("    %-22s OTR=%6.2f  TVS(L=0)=%6.2f  selisih=%+7.2f\n",
			trunc(n, 22), lain[n].OTR, tvsL[n][0], tvsL[n][0]-lain[n].OTR)
	}

	// ---------------------------------------------------------- 13.7
	p("\n13.7 Pengaruh bobot setelah TVS memakai lambda=0\n")
	p("Kini bobot punya pijakan, sebab tugas telat keluar dari pembilang.\n\n")
	p("%-22s", "Karyawan")
	for _, sk := range skemaUji() {
		p(" %10s", sk.nama)
	}
	p(" %10s\n", "rentang")
	maksRentangBaru := 0.0
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		p("%-22s", trunc(n, 22))
		mn, mx := math.Inf(1), math.Inf(-1)
		for _, sk := range skemaUji() {
			wa, _, wo := jumlahBobot(ps, sk)
			v := 0.0
			if wa > 0 {
				v = wo / wa * 100
			}
			mn, mx = math.Min(mn, v), math.Max(mx, v)
			p(" %10.4f", v)
		}
		p(" %10.4f\n", mx-mn)
		maksRentangBaru = math.Max(maksRentangBaru, mx-mn)
	}
	p("\nRentang terbesar: %.4f poin pada rumus usulan, %.4f poin pada rumus sekarang.\n",
		maksRentangBaru, maksRentangLama)

	// ---------------------------------------------------------- 13.8
	p("\n13.8 Menentukan lambda: kategori kedelapan karyawan pada tiap nilai\n")
	uji8 := []float64{1, 0.75, 0.5, 0.25, 0}
	for _, pakaiBatas := range []bool{false, true} {
		cfg := tanpaBatas
		ket := "tanpa kaidah pembatas"
		if pakaiBatas {
			cfg = dasar
			ket = "kaidah pembatas tetap dipasang"
		}
		p("\n  %s\n", ket)
		p("  %-24s", "Karyawan")
		for _, l := range uji8 {
			p(" %-14s", fmt.Sprintf("L=%.2f", l))
		}
		p("\n")
		for _, n := range nama {
			ps := petakUser[n]
			if len(ps) == 0 {
				continue
			}
			k := lain[n]
			p("  %-24s", trunc(n, 24))
			for _, l := range uji8 {
				_, c, _, _ := cfg.Infer([]float64{k.TCR, k.OTR, tvsL[n][l], k.WER})
				p(" %-14s", trunc(c, 14))
			}
			p("\n")
		}
		p("  %-24s", "(sempurna, selalu telat)")
		for _, l := range uji8 {
			_, c, _, _ := cfg.Infer([]float64{100, 0, l * 100, 100})
			p(" %-14s", trunc(c, 14))
		}
		p("\n")
	}

	// ---------------------------------------------------------- 13.9
	p("\n13.9 Apakah kaidah pembatas masih berguna setelah lambda diturunkan\n")
	p("%-10s %14s %14s\n", "lambda", "bedaKategori", "bedaZ maks")
	for _, l := range uji8 {
		beda := 0
		maksBeda := 0.0
		for _, n := range nama {
			ps := petakUser[n]
			if len(ps) == 0 {
				continue
			}
			k := lain[n]
			x := []float64{k.TCR, k.OTR, tvsL[n][l], k.WER}
			z1, c1, _, _ := dasar.Infer(x)
			z2, c2, _, _ := tanpaBatas.Infer(x)
			if c1 != c2 {
				beda++
			}
			maksBeda = math.Max(maksBeda, math.Abs(z1-z2))
		}
		p("%-10.2f %14d %14.2f\n", l, beda, maksBeda)
	}
	p("\nAngka 0 pada kolom bedaKategori berarti kaidah pembatas tidak lagi mengubah\n")
	p("keputusan apa pun: sifat tidak-saling-menutup sudah dibawa oleh data.\n")
}

func rerataBeda(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	s := 0.0
	for i := range a {
		s += math.Abs(a[i] - b[i])
	}
	return s / float64(len(a))
}
