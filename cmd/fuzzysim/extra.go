package main

// Eksperimen 8 dan 9: perbandingan terhadap agregasi linier tegas,
// serta kalibrasi fungsi keanggotaan terhadap sebaran data operasional.

import (
	"math"
	"sort"
)

// ============================================================
// 8. Perbandingan terhadap agregasi linier tegas
// ============================================================

// skor linier tegas dengan bobot yang dipakai aplikasi sebelum revisi
func linearScore(x []float64) float64 {
	return 0.3*x[0] + 0.3*x[1] + 0.2*x[2] + 0.2*x[3]
}

func expLinear() {
	header("EKSPERIMEN 8 - FUZZY TSUKAMOTO DIBANDING AGREGASI LINIER TEGAS")
	a := newConfig("A", "rancangan awal 2 himpunan", setsBahu(), defaultVars)
	b := baseline()

	p("\n8.1 Nilai dan kategori pada data operasional gabungan\n")
	p("%-24s %10s %-13s %10s %-13s %10s %-13s\n", "Karyawan", "Linier", "kategori", "Fuzzy A", "kategori", "Fuzzy B", "kategori")
	var vl, va, vb []float64
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		x := []float64{r.TCR, r.OTR, r.TPS, r.WER}
		l := linearScore(x)
		za, ca, _, _ := a.Infer(x)
		zb, cb, _, _ := b.Infer(x)
		p("%-24s %10.2f %-13s %10.2f %-13s %10.2f %-13s\n", trunc(r.User, 24), l, Category(l), za, ca, zb, cb)
		vl = append(vl, l)
		va = append(va, za)
		vb = append(vb, zb)
	}
	p("\nKorelasi peringkat Spearman: Linier vs A = %.4f ; Linier vs B = %.4f ; A vs B = %.4f\n",
		spearman(vl, va), spearman(vl, vb), spearman(va, vb))

	p("\n8.2 Perilaku di sekitar ambang kategori (TCR=TPS=100, WER=70, OTR disapu)\n")
	p("%-8s %10s %-13s %10s %-13s %10s %-13s\n", "OTR", "Linier", "kategori", "Fuzzy A", "kategori", "Fuzzy B", "kategori")
	prevL, prevA, prevB := "", "", ""
	jumpL, jumpA, jumpB := 0, 0, 0
	for otr := 30.0; otr <= 80; otr += 2 {
		x := []float64{100, otr, 100, 70}
		l := linearScore(x)
		za, ca, _, _ := a.Infer(x)
		zb, cb, _, _ := b.Infer(x)
		cl := Category(l)
		if prevL != "" {
			if cl != prevL {
				jumpL++
			}
			if ca != prevA {
				jumpA++
			}
			if cb != prevB {
				jumpB++
			}
		}
		prevL, prevA, prevB = cl, ca, cb
		p("%-8.0f %10.2f %-13s %10.2f %-13s %10.2f %-13s\n", otr, l, cl, za, ca, zb, cb)
	}
	p("Jumlah perpindahan kategori sepanjang sapuan: Linier=%d, Fuzzy A=%d, Fuzzy B=%d\n", jumpL, jumpA, jumpB)

	p("\n8.3 Ketahanan hasil terhadap gangguan kecil pada input\n")
	p("Satu variabel dinaikkan 1 poin; dihitung berapa persen titik uji berpindah kategori.\n")
	grid := [][]float64{}
	for a1 := 20.0; a1 <= 100; a1 += 4 {
		for a2 := 20.0; a2 <= 100; a2 += 4 {
			for a3 := 20.0; a3 <= 100; a3 += 4 {
				for a4 := 20.0; a4 <= 100; a4 += 4 {
					grid = append(grid, []float64{a1, a2, a3, a4})
				}
			}
		}
	}
	chL, chA, chB, n := 0, 0, 0, 0
	for _, g := range grid {
		l0 := Category(linearScore(g))
		_, a0, _, _ := a.Infer(g)
		_, b0, _, _ := b.Infer(g)
		for i := 0; i < 4; i++ {
			if g[i]+1 > 100 {
				continue
			}
			nx := append([]float64{}, g...)
			nx[i] += 1
			n++
			if Category(linearScore(nx)) != l0 {
				chL++
			}
			if _, c1, _, _ := a.Infer(nx); c1 != a0 {
				chA++
			}
			if _, c1, _, _ := b.Infer(nx); c1 != b0 {
				chB++
			}
		}
	}
	p("  titik uji = %d ; berpindah kategori: Linier %.3f%% , Fuzzy A %.3f%% , Fuzzy B %.3f%%\n",
		n, float64(chL)/float64(n)*100, float64(chA)/float64(n)*100, float64(chB)/float64(n)*100)
}

// ============================================================
// 9. Kalibrasi fungsi keanggotaan terhadap sebaran data operasional
// ============================================================

func quantile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	i := q * float64(len(sorted)-1)
	lo := int(math.Floor(i))
	hi := int(math.Ceil(i))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(i-float64(lo))
}

// nilai indikator yang teramati per variabel, dari seluruh baris periode tahunan
func observed() [4][]float64 {
	var v [4][]float64
	for _, r := range rows() {
		if r.Period == "GAB" {
			continue
		}
		for i, x := range []float64{r.TCR, r.OTR, r.TPS, r.WER} {
			if x > 100 {
				x = 100
			}
			v[i] = append(v[i], x)
		}
	}
	for i := range v {
		sort.Float64s(v[i])
	}
	return v
}

func calibConfigs() (*Config, *Config, [4][3]float64) {
	obs := observed()
	var pts [4][3]float64
	perVar2 := [][]FuzzySet{}
	perVar3 := [][]FuzzySet{}
	for i := 0; i < 4; i++ {
		q1 := quantile(obs[i], 0.25)
		med := quantile(obs[i], 0.50)
		q3 := quantile(obs[i], 0.75)
		// lebar transisi dijaga minimal 10 poin dan tetap berada di dalam semesta
		if q3-q1 < 10 {
			q1, q3 = med-5, med+5
		}
		if q1 < 0 {
			q3 -= q1
			q1 = 0
		}
		if q3 > 100 {
			q1 -= q3 - 100
			q3 = 100
		}
		if med <= q1 || med >= q3 {
			med = (q1 + q3) / 2
		}
		pts[i] = [3]float64{q1, med, q3}
		perVar2 = append(perVar2, []FuzzySet{
			{"Rendah", shoulderLeft(q1, q3)},
			{"Tinggi", shoulderRight(q1, q3)},
		})
		perVar3 = append(perVar3, []FuzzySet{
			{"Rendah", shoulderLeft(q1, med)},
			{"Sedang", triangle(q1, med, q3)},
			{"Tinggi", shoulderRight(med, q3)},
		})
	}
	e := &Config{Name: "G", Desc: "2 himpunan bahu, transisi [Q1,Q3] per variabel", PerVar: perVar2, VarNames: defaultVars}
	e.Rules = buildRules(defaultVars, 2)
	f := &Config{Name: "H", Desc: "3 himpunan terkalibrasi (Q1, median, Q3) per variabel", PerVar: perVar3, VarNames: defaultVars}
	f.Rules = buildRules(defaultVars, 3)
	return e, f, pts
}

func expCalibrated() {
	header("EKSPERIMEN 9 - KALIBRASI FUNGSI KEANGGOTAAN TERHADAP SEBARAN DATA OPERASIONAL")
	obs := observed()
	p("\n9.1 Sebaran nilai indikator pada data operasional (n=%d baris tahunan)\n", len(obs[0]))
	p("%-6s %8s %8s %8s %8s %8s %8s\n", "Var", "min", "Q1", "median", "Q3", "maks", "sd")
	for i, nm := range defaultVars {
		v := obs[i]
		mean, sd := 0.0, 0.0
		for _, x := range v {
			mean += x
		}
		mean /= float64(len(v))
		for _, x := range v {
			sd += (x - mean) * (x - mean)
		}
		sd = math.Sqrt(sd / float64(len(v)))
		p("%-6s %8.2f %8.2f %8.2f %8.2f %8.2f %8.2f\n", nm, v[0], quantile(v, 0.25), quantile(v, 0.5), quantile(v, 0.75), v[len(v)-1], sd)
	}

	e, f, pts := calibConfigs()
	p("\n9.2 Titik potong hasil kalibrasi\n")
	for i, nm := range defaultVars {
		p("  %-5s Q1=%.2f  median=%.2f  Q3=%.2f\n", nm, pts[i][0], pts[i][1], pts[i][2])
	}

	a := newConfig("A", "rancangan awal 2 himpunan 40-60", setsBahu(), defaultVars)
	b := baseline()
	cfgs := []*Config{a, b, e, f}

	p("\n9.3 Nilai Z pada data operasional gabungan\n")
	p("%-24s", "Karyawan")
	for _, c := range cfgs {
		p(" %10s", c.Name)
	}
	p("   | kategori E, F\n")
	series := map[string][]float64{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		x := []float64{r.TCR, r.OTR, r.TPS, r.WER}
		p("%-24s", trunc(r.User, 24))
		ce, cf := "", ""
		for _, c := range cfgs {
			z, cat, _, _ := c.Infer(x)
			p(" %10.2f", z)
			series[c.Name] = append(series[c.Name], z)
			if c.Name == "G" {
				ce = cat
			}
			if c.Name == "H" {
				cf = cat
			}
		}
		p("   | %s, %s\n", ce, cf)
	}

	p("\n9.4 Daya pembeda\n")
	p("%-4s %-52s %8s %8s %8s %8s %6s\n", "Cfg", "Deskripsi", "min", "maks", "rentang", "sd", "unik")
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
		p("%-4s %-52s %8.2f %8.2f %8.2f %8.2f %6d\n", c.Name, trunc(c.Desc, 52), mn, mx, mx-mn, math.Sqrt(vr/float64(len(v))), len(u))
	}

	p("\n9.5 Peringkat karyawan menurut konfigurasi F (usulan)\n")
	type ur struct {
		n string
		z float64
		c string
	}
	list := []ur{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		z, cat, _, _ := f.Infer([]float64{r.TCR, r.OTR, r.TPS, r.WER})
		list = append(list, ur{r.User, z, cat})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].z > list[j].z })
	for i, u := range list {
		p("  %d. %-24s Z=%6.2f  %s\n", i+1, trunc(u.n, 24), u.z, u.c)
	}

	p("\n9.6 Sebaran kategori pada sapuan grid semesta (langkah 10) per konfigurasi\n")
	p("%-4s", "Cfg")
	for _, k := range outLabels {
		p(" %13s", k)
	}
	p(" %10s\n", "Z=100")
	for _, c := range cfgs {
		cnt := map[string]int{}
		sat, n := 0, 0
		for a1 := 0.0; a1 <= 100; a1 += 10 {
			for a2 := 0.0; a2 <= 100; a2 += 10 {
				for a3 := 0.0; a3 <= 100; a3 += 10 {
					for a4 := 0.0; a4 <= 100; a4 += 10 {
						z, cat, _, _ := c.Infer([]float64{a1, a2, a3, a4})
						cnt[cat]++
						n++
						if z >= 99.999 {
							sat++
						}
					}
				}
			}
		}
		p("%-4s", c.Name)
		for _, k := range outLabels {
			p(" %12.2f%%", float64(cnt[k])/float64(n)*100)
		}
		p(" %9.2f%%\n", float64(sat)/float64(n)*100)
	}
}
