package main

// Eksperimen 12: peran kaidah agregasi dan kaidah pembatas pada basis aturan.
//
// Basis aturan sistem tidak ditulis satu per satu, melainkan dibangkitkan oleh
// dua kaidah yang bekerja berurutan pada setiap kombinasi premis:
//
//	kaidah agregasi  : r = jumlah indeks himpunan / (n_variabel * (m_himpunan-1))
//	                   level konsekuen = int(r * 5), dipotong maksimum 4
//	kaidah pembatas  : bila OTR berada pada himpunan terendah, level dibatasi
//	                   maksimum 2 yaitu "Cukup"
//
// Eksperimen ini membandingkan empat varian basis aturan untuk menunjukkan apa
// yang hilang bila salah satu atau kedua kaidah itu ditiadakan.

import (
	"fmt"
	"math"
	"sort"
)

// varian menentukan bagaimana konsekuen tiap aturan ditetapkan.
type varian struct {
	nama    string
	ket     string
	agregat bool // kaidah agregasi dipakai
	batas   bool // kaidah pembatas dipakai
	tunggal int  // >= 0: konsekuen diambil dari satu variabel ini saja
	tetap   int  // dipakai bila agregat mati dan tunggal < 0: level tetap
}

// bangunAturan membangkitkan basis aturan menurut varian yang diberikan.
// Bentuknya sengaja dibuat sedekat mungkin dengan buildRules agar perbedaan
// hasil benar-benar berasal dari kaidahnya, bukan dari cara lain menghitung.
func bangunAturan(varNames []string, nSets int, v varian) []Rule {
	n := len(varNames)
	otrPos := -1
	for i, nm := range varNames {
		if nm == "OTR" {
			otrPos = i
		}
	}
	total := 1
	for i := 0; i < n; i++ {
		total *= nSets
	}
	rules := make([]Rule, 0, total)
	idx := make([]int, n)
	for k := 0; k < total; k++ {
		rem := k
		for i := n - 1; i >= 0; i-- {
			idx[i] = nSets - 1 - rem%nSets
			rem /= nSets
		}

		level := v.tetap
		switch {
		case v.agregat:
			sum := 0
			for _, x := range idx {
				sum += x
			}
			level = int(float64(sum) / float64(n*(nSets-1)) * 5)
			if level > 4 {
				level = 4
			}
		case v.tunggal >= 0 && v.tunggal < n:
			// konsekuen hanya mengikuti satu variabel; variabel lain tidak berperan
			level = int(float64(idx[v.tunggal]) / float64(nSets-1) * 5)
			if level > 4 {
				level = 4
			}
		}

		if v.batas && otrPos >= 0 && idx[otrPos] == 0 && level > 2 {
			level = 2
		}

		cp := make([]int, n)
		copy(cp, idx)
		rules = append(rules, Rule{
			Code:   fmt.Sprintf("R%d", k+1),
			SetIdx: cp,
			Output: outLabels[level],
			OutIdx: level,
		})
	}
	return rules
}

func cfgVarian(v varian) *Config {
	c := &Config{Name: v.nama, Desc: v.ket, Sets: setsSegitiga(), VarNames: defaultVars}
	c.Rules = bangunAturan(defaultVars, len(c.Sets), v)
	return c
}

func varianKaidah() []varian {
	return []varian{
		{nama: "V1", ket: "agregasi + pembatas (rancangan sistem)", agregat: true, batas: true, tunggal: -1},
		{nama: "V2", ket: "agregasi saja, pembatas dihapus", agregat: true, batas: false, tunggal: -1},
		{nama: "V3", ket: "agregasi dihapus, konsekuen ikut TCR saja", agregat: false, batas: true, tunggal: 0},
		{nama: "V4", ket: "agregasi dihapus, konsekuen seragam Cukup", agregat: false, batas: true, tunggal: -1, tetap: 2},
		{nama: "V5", ket: "kedua kaidah dihapus, konsekuen seragam Cukup", agregat: false, batas: false, tunggal: -1, tetap: 2},
	}
}

func expKaidah() {
	header("EKSPERIMEN 12 - PERAN KAIDAH AGREGASI DAN KAIDAH PEMBATAS")

	vs := varianKaidah()
	cfgs := make([]*Config, len(vs))
	for i, v := range vs {
		cfgs[i] = cfgVarian(v)
	}
	dasar := cfgs[0]

	// ---------------------------------------------------------- 12.1
	p("\n12.1 Susunan konsekuen pada 81 aturan\n")
	p("%-4s %-46s %8s %s\n", "Var", "Keterangan", "beda", "sebaran konsekuen")
	for i, c := range cfgs {
		beda := 0
		for j := range c.Rules {
			if c.Rules[j].OutIdx != dasar.Rules[j].OutIdx {
				beda++
			}
		}
		cnt := map[int]int{}
		for _, r := range c.Rules {
			cnt[r.OutIdx]++
		}
		s := ""
		for lv := 4; lv >= 0; lv-- {
			if cnt[lv] > 0 {
				s += fmt.Sprintf("%s=%d ", trunc(outLabels[lv], 12), cnt[lv])
			}
		}
		p("%-4s %-46s %8d %s\n", c.Name, trunc(vs[i].ket, 46), beda, s)
	}
	p("\nKolom beda adalah jumlah aturan yang konsekuennya berbeda dari V1, dari 81 aturan.\n")

	// ---------------------------------------------------------- 12.2
	p("\n12.2 Aturan yang diubah oleh kaidah pembatas\n")
	p("Aturan berikut adalah aturan yang konsekuennya diturunkan karena OTR Rendah.\n\n")
	p("%-6s %-42s %-14s %-14s\n", "Kode", "Premis", "tanpa pembatas", "dengan pembatas")
	n := 0
	for j := range dasar.Rules {
		if dasar.Rules[j].OutIdx != cfgs[1].Rules[j].OutIdx {
			r := dasar.Rules[j]
			prem := ""
			for i, nm := range defaultVars {
				if i > 0 {
					prem += ", "
				}
				prem += fmt.Sprintf("%s %s", nm, setLabel(r.SetIdx[i]))
			}
			p("%-6s %-42s %-14s %-14s\n", r.Code, trunc(prem, 42),
				cfgs[1].Rules[j].Output, r.Output)
			n++
		}
	}
	p("\njumlah aturan yang diturunkan: %d dari 81\n", n)

	// ---------------------------------------------------------- 12.3
	p("\n12.3 Batas nilai tertinggi ketika OTR rendah\n")
	p("Tiga indikator lain ditahan di 100, OTR disapu dari 0 sampai 100.\n\n")
	p("%-6s", "OTR")
	for _, c := range cfgs {
		p(" %10s %-13s", c.Name, "kategori")
	}
	p("\n")
	for otr := 0.0; otr <= 100; otr += 10 {
		p("%-6.0f", otr)
		for _, c := range cfgs {
			z, cat, _, _ := c.Infer([]float64{100, otr, 100, 100})
			p(" %10.2f %-13s", z, cat)
		}
		p("\n")
	}

	// ---------------------------------------------------------- 12.4
	p("\n12.4 Nilai kedelapan karyawan pada data operasional\n")
	p("%-22s %6s %6s %6s %6s", "Karyawan", "TCR", "OTR", "TVS", "WER")
	for _, c := range cfgs {
		p(" %8s %-13s", c.Name, "kategori")
	}
	p("\n")
	series := map[string][]float64{}
	for _, r := range rows() {
		if r.Period != "GAB" {
			continue
		}
		x := []float64{r.TCR, r.OTR, r.TVS, r.WER}
		p("%-22s %6.2f %6.2f %6.2f %6.2f", trunc(r.User, 22), r.TCR, r.OTR, r.TVS, r.WER)
		for _, c := range cfgs {
			z, cat, _, _ := c.Infer(x)
			p(" %8.2f %-13s", z, cat)
			series[c.Name] = append(series[c.Name], z)
		}
		p("\n")
	}

	// ---------------------------------------------------------- 12.5
	p("\n12.5 Daya pembeda dan kesesuaian peringkat\n")
	p("%-4s %-46s %8s %8s %8s %6s %10s\n",
		"Var", "Keterangan", "min", "maks", "rentang", "unik", "SpearmanV1")
	acuan := series[dasar.Name]
	for i, c := range cfgs {
		v := series[c.Name]
		mn, mx := math.Inf(1), math.Inf(-1)
		u := map[float64]bool{}
		for _, x := range v {
			mn = math.Min(mn, x)
			mx = math.Max(mx, x)
			u[math.Round(x*100)/100] = true
		}
		p("%-4s %-46s %8.2f %8.2f %8.2f %6d %10.4f\n",
			c.Name, trunc(vs[i].ket, 46), mn, mx, mx-mn, len(u), spearman(v, acuan))
	}
	p("\nNilai unik adalah banyaknya nilai Z yang berbeda di antara delapan karyawan.\n")
	p("Semakin kecil, semakin banyak karyawan yang dinilai sama persis.\n")

	// ---------------------------------------------------------- 12.6
	p("\n12.6 Sapuan seluruh semesta, langkah 10 poin (14.641 titik)\n")
	p("%-4s %9s %9s %9s %9s %9s %9s\n",
		"Var", "bedaKat%", "rerataZ", "sdZ", "minZ", "maksZ", "jenuh%")
	for _, c := range cfgs {
		var vals []float64
		beda, tot, jenuh := 0, 0, 0
		for a1 := 0.0; a1 <= 100; a1 += 10 {
			for a2 := 0.0; a2 <= 100; a2 += 10 {
				for a3 := 0.0; a3 <= 100; a3 += 10 {
					for a4 := 0.0; a4 <= 100; a4 += 10 {
						x := []float64{a1, a2, a3, a4}
						z, cat, _, _ := c.Infer(x)
						_, cat0, _, _ := dasar.Infer(x)
						tot++
						if cat != cat0 {
							beda++
						}
						if z >= 99.999 {
							jenuh++
						}
						vals = append(vals, z)
					}
				}
			}
		}
		sum, mn, mx := 0.0, math.Inf(1), math.Inf(-1)
		for _, x := range vals {
			sum += x
			mn = math.Min(mn, x)
			mx = math.Max(mx, x)
		}
		rerata := sum / float64(len(vals))
		vr := 0.0
		for _, x := range vals {
			vr += (x - rerata) * (x - rerata)
		}
		p("%-4s %9.2f %9.2f %9.2f %9.2f %9.2f %9.2f\n",
			c.Name, float64(beda)/float64(tot)*100, rerata,
			math.Sqrt(vr/float64(len(vals))), mn, mx,
			float64(jenuh)/float64(tot)*100)
	}

	// ---------------------------------------------------------- 12.7
	p("\n12.7 Berapa banyak titik yang naik kategori bila pembatas dihapus\n")
	naik, turun, tetap2 := 0, 0, 0
	contoh := []string{}
	for a1 := 0.0; a1 <= 100; a1 += 10 {
		for a2 := 0.0; a2 <= 100; a2 += 10 {
			for a3 := 0.0; a3 <= 100; a3 += 10 {
				for a4 := 0.0; a4 <= 100; a4 += 10 {
					x := []float64{a1, a2, a3, a4}
					z1, c1, _, _ := dasar.Infer(x)
					z2, c2, _, _ := cfgs[1].Infer(x)
					switch {
					case z2 > z1+1e-9:
						naik++
						if a2 <= 30 && len(contoh) < 8 {
							contoh = append(contoh, fmt.Sprintf(
								"  TCR=%3.0f OTR=%3.0f TVS=%3.0f WER=%3.0f : %6.2f %-13s -> %6.2f %-13s",
								a1, a2, a3, a4, z1, c1, z2, c2))
						}
					case z2 < z1-1e-9:
						turun++
					default:
						tetap2++
					}
				}
			}
		}
	}
	p("  naik: %d titik, turun: %d titik, tetap: %d titik\n", naik, turun, tetap2)
	p("\n  contoh titik dengan OTR rendah yang terangkat bila pembatas dihapus:\n")
	for _, s := range contoh {
		p("%s\n", s)
	}

	// ---------------------------------------------------------- 12.8
	p("\n12.8 Peringkat karyawan menurut tiap varian\n")
	type ur struct {
		n string
		z float64
	}
	for _, c := range cfgs {
		list := []ur{}
		for _, r := range rows() {
			if r.Period != "GAB" {
				continue
			}
			z, _, _, _ := c.Infer([]float64{r.TCR, r.OTR, r.TVS, r.WER})
			list = append(list, ur{r.User, z})
		}
		sort.Slice(list, func(i, j int) bool { return list[i].z > list[j].z })
		p("  %-4s ", c.Name)
		for i, u := range list {
			if i > 0 {
				p(" > ")
			}
			p("%s", trunc(u.n, 12))
		}
		p("\n")
	}
}

func setLabel(i int) string {
	switch i {
	case 0:
		return "Rendah"
	case 1:
		return "Sedang"
	default:
		return "Tinggi"
	}
}
