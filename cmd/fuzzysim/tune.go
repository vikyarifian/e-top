package main

// Pencarian parameter fungsi keanggotaan tiga himpunan.
// Partisi Ruspini: untuk simpul P1 < P2 < P3,
//   Rendah = bahu kiri (1 sampai P1, turun ke 0 di P2)
//   Sedang = segitiga (P1, P2, P3)
//   Tinggi = bahu kanan (0 di P2, naik ke 1 di P3)
// sehingga derajat keanggotaan selalu berjumlah 1 pada seluruh semesta.

import (
	"fmt"
	"math"
	"sort"
)

func setsTri3(p1, p2, p3 float64) []FuzzySet {
	return []FuzzySet{
		{"Rendah", shoulderLeft(p1, p2)},
		{"Sedang", triangle(p1, p2, p3)},
		{"Tinggi", shoulderRight(p2, p3)},
	}
}

func cfgTri3(name string, p1, p2, p3 float64) *Config {
	c := &Config{
		Name:     name,
		Desc:     fmt.Sprintf("segitiga 3 himpunan P=(%.0f, %.0f, %.0f)", p1, p2, p3),
		Sets:     setsTri3(p1, p2, p3),
		VarNames: defaultVars,
	}
	c.Rules = buildRules(defaultVars, 3)
	return c
}

type metrik struct {
	P1, P2, P3 float64
	ViolPct    float64 // persen pelanggaran monotonisitas
	MaxDrop    float64 // penurunan Z terbesar
	SatPct     float64 // persen titik grid dengan Z = 100
	Unik       int     // jumlah nilai Z berbeda pada 8 karyawan
	Rentang    float64 // rentang Z pada 8 karyawan
	SD         float64 // simpangan baku Z pada 8 karyawan
	SpearmanF  float64 // korelasi peringkat terhadap konfigurasi terkalibrasi
	Z0, Z50    float64 // Z untuk seluruh input 0 dan 50
	Z100       float64 // Z untuk seluruh input 100
}

func ukur(c *Config, ref []float64) metrik {
	var m metrik
	// monotonisitas dan saturasi pada sapuan grid langkah 5
	viol, tot, sat, n := 0, 0, 0, 0
	for a1 := 0.0; a1 <= 100; a1 += 5 {
		for a2 := 0.0; a2 <= 100; a2 += 5 {
			for a3 := 0.0; a3 <= 100; a3 += 5 {
				for a4 := 0.0; a4 <= 100; a4 += 5 {
					b := []float64{a1, a2, a3, a4}
					z0 := c.Score(b)
					n++
					if z0 >= 99.999 {
						sat++
					}
					for i := 0; i < 4; i++ {
						if b[i]+5 > 100 {
							continue
						}
						nx := append([]float64{}, b...)
						nx[i] += 5
						tot++
						if d := z0 - c.Score(nx); d > 1e-9 {
							viol++
							if d > m.MaxDrop {
								m.MaxDrop = d
							}
						}
					}
				}
			}
		}
	}
	m.ViolPct = float64(viol) / float64(tot) * 100
	m.SatPct = float64(sat) / float64(n) * 100

	// daya pembeda pada data operasional gabungan
	vals := []float64{}
	u := map[float64]bool{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		z := c.Score([]float64{r.TCR, r.OTR, r.TPS, r.WER})
		vals = append(vals, z)
		u[math.Round(z*100)/100] = true
	}
	m.Unik = len(u)
	mn, mx, sum := math.Inf(1), math.Inf(-1), 0.0
	for _, x := range vals {
		mn = math.Min(mn, x)
		mx = math.Max(mx, x)
		sum += x
	}
	m.Rentang = mx - mn
	mean := sum / float64(len(vals))
	vr := 0.0
	for _, x := range vals {
		vr += (x - mean) * (x - mean)
	}
	m.SD = math.Sqrt(vr / float64(len(vals)))
	if ref != nil {
		m.SpearmanF = spearman(vals, ref)
	}
	m.Z0 = c.Score([]float64{0, 0, 0, 0})
	m.Z50 = c.Score([]float64{50, 50, 50, 50})
	m.Z100 = c.Score([]float64{100, 100, 100, 100})
	return m
}

func expTune() {
	header("EKSPERIMEN 11 - PENENTUAN PARAMETER FUNGSI KEANGGOTAAN TIGA HIMPUNAN")

	// peringkat acuan: konfigurasi terkalibrasi kuantil, dipakai sebagai pembanding peringkat
	_, f, _ := calibConfigs()
	ref := []float64{}
	for _, r := range rows() {
		if r.Period == "GAB" {
			ref = append(ref, f.Score([]float64{r.TCR, r.OTR, r.TPS, r.WER}))
		}
	}

	p("\n11.1 Sapuan kombinasi simpul P1 < P2 < P3\n")
	p("%-20s %9s %9s %8s %6s %8s %8s %9s %7s %7s %7s\n",
		"P=(P1,P2,P3)", "langgar%", "maksTurun", "jenuh%", "unik", "rentang", "sd", "Spearman", "Z(0)", "Z(50)", "Z(100)")
	type baris struct {
		nama string
		m    metrik
	}
	hasil := []baris{}
	for _, p1 := range []float64{0, 20, 30, 40, 50} {
		for _, p2 := range []float64{40, 50, 60, 70, 80} {
			for _, p3 := range []float64{70, 80, 90, 100} {
				if !(p1 < p2 && p2 < p3) {
					continue
				}
				c := cfgTri3("T", p1, p2, p3)
				m := ukur(c, ref)
				m.P1, m.P2, m.P3 = p1, p2, p3
				nm := fmt.Sprintf("(%.0f, %.0f, %.0f)", p1, p2, p3)
				hasil = append(hasil, baris{nm, m})
			}
		}
	}
	// urutkan dari pelanggaran monotonisitas terkecil
	sort.Slice(hasil, func(i, j int) bool { return hasil[i].m.ViolPct < hasil[j].m.ViolPct })
	for _, h := range hasil {
		p("%-20s %9.3f %9.3f %8.3f %6d %8.2f %8.2f %9.4f %7.2f %7.2f %7.2f\n",
			h.nama, h.m.ViolPct, h.m.MaxDrop, h.m.SatPct, h.m.Unik, h.m.Rentang, h.m.SD,
			h.m.SpearmanF, h.m.Z0, h.m.Z50, h.m.Z100)
	}

	p("\n11.2 Penyaringan menurut kriteria kelayakan\n")
	p("Kriteria: pelanggaran monotonisitas < 5 persen, saturasi < 1 persen,\n")
	p("nilai Z berbeda untuk kedelapan karyawan minimal 7, dan Z(0)=0 serta Z(100)=100.\n\n")
	p("%-20s %9s %9s %8s %6s %8s %8s\n", "P=(P1,P2,P3)", "langgar%", "maksTurun", "jenuh%", "unik", "rentang", "sd")
	lolos := []baris{}
	for _, h := range hasil {
		if h.m.ViolPct < 5 && h.m.SatPct < 1 && h.m.Unik >= 7 && h.m.Z0 <= 0.01 && h.m.Z100 >= 99.99 {
			lolos = append(lolos, h)
			p("%-20s %9.3f %9.3f %8.3f %6d %8.2f %8.2f\n",
				h.nama, h.m.ViolPct, h.m.MaxDrop, h.m.SatPct, h.m.Unik, h.m.Rentang, h.m.SD)
		}
	}
	if len(lolos) == 0 {
		p("  tidak ada kombinasi yang memenuhi seluruh kriteria\n")
	}

	_ = lolos

	p("\n11.3 Rincian kandidat unggulan pada data operasional\n")
	kandidat := []*Config{
		newConfig("A", "baseline 2 himpunan, transisi 40-60", setsBahu(), defaultVars),
		cfgTri3("K1", 40, 60, 90),
		cfgTri3("K2", 50, 60, 90),
		cfgTri3("K3", 50, 70, 90),
		cfgTri3("K4", 30, 60, 90),
		cfgTri3("K5", 0, 50, 100),
	}
	p("\n%-4s %-38s %9s %9s %8s %6s %8s %8s %7s %7s\n",
		"Cfg", "Deskripsi", "langgar%", "maksTurun", "jenuh%", "unik", "rentang", "sd", "Z(50)", "Z(100)")
	for _, c := range kandidat {
		m := ukur(c, ref)
		p("%-4s %-38s %9.3f %9.3f %8.3f %6d %8.2f %8.2f %7.2f %7.2f\n",
			c.Name, trunc(c.Desc, 38), m.ViolPct, m.MaxDrop, m.SatPct, m.Unik, m.Rentang, m.SD, m.Z50, m.Z100)
	}

	p("\n%-24s", "Karyawan")
	for _, c := range kandidat {
		p(" %10s", c.Name)
	}
	p("   | kategori K1\n")
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		p("%-24s", trunc(r.User, 24))
		for _, c := range kandidat {
			p(" %10.2f", c.Score([]float64{r.TCR, r.OTR, r.TPS, r.WER}))
		}
		_, cat, _, _ := kandidat[1].Infer([]float64{r.TCR, r.OTR, r.TPS, r.WER})
		p("   | %s\n", cat)
	}

	p("\n11.4 Profil Z pada input seragam untuk kandidat\n")
	p("%-8s", "x")
	for _, c := range kandidat {
		p(" %10s", c.Name)
	}
	p("\n")
	for x := 0.0; x <= 100; x += 5 {
		p("%-8.0f", x)
		for _, c := range kandidat {
			p(" %10.2f", c.Score([]float64{x, x, x, x}))
		}
		p("\n")
	}
}
