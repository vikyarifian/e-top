package main

import "math"

// expMono menguji monotonisitas seluruh konfigurasi:
// menaikkan satu variabel seharusnya tidak menurunkan nilai Z.
func expMono() {
	header("EKSPERIMEN 10 - UJI MONOTONISITAS SELURUH KONFIGURASI")
	e, f, _ := calibConfigs()
	cfgs := append(mfConfigs(), e, f)

	p("\n%-4s %-52s %9s %9s %10s %10s %9s\n", "Cfg", "Deskripsi", "diuji", "langgar", "persen", "maksTurun", "jenuh%")
	for _, c := range cfgs {
		viol, tot := 0, 0
		sat, npoint := 0, 0
		maxDrop := 0.0
		var at []float64
		for a1 := 0.0; a1 <= 100; a1 += 5 {
			for a2 := 0.0; a2 <= 100; a2 += 5 {
				for a3 := 0.0; a3 <= 100; a3 += 5 {
					for a4 := 0.0; a4 <= 100; a4 += 5 {
						b := []float64{a1, a2, a3, a4}
						z0 := c.Score(b)
						npoint++
						if z0 >= 99.999 {
							sat++
						}
						for i := 0; i < 4; i++ {
							if b[i]+5 > 100 {
								continue
							}
							nx := append([]float64{}, b...)
							nx[i] += 5
							z1 := c.Score(nx)
							tot++
							if d := z0 - z1; d > 1e-9 {
								viol++
								if d > maxDrop {
									maxDrop = d
									at = append([]float64{}, b...)
								}
							}
						}
					}
				}
			}
		}
		p("%-4s %-52s %9d %9d %9.3f%% %10.4f %8.3f%%  %v\n", c.Name, trunc(c.Desc, 52), tot, viol,
			float64(viol)/float64(tot)*100, maxDrop, float64(sat)/float64(npoint)*100, at)
	}

	p("\n10.2 Profil Z pada input seragam untuk konfigurasi terpilih\n")
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
	_ = math.Abs
}
