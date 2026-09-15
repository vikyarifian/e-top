package main

// Eksperimen 15: bagaimana bila sebaran prioritas dan dampak merata?
//
// Kesimpulan eksperimen 14 sebagian bersandar pada kenyataan bahwa 95,78 persen
// tugas pada data penelitian berdampak HIGH. Eksperimen ini membangun data
// tandingan yang sebarannya merata pada sembilan kombinasi prioritas x dampak,
// lalu mengulang seluruh pengukuran.
//
// Jumlah tugas, jumlah tugas selesai, dan jumlah tugas tepat waktu tiap
// karyawan dipertahankan persis seperti aslinya, sehingga TCR dan OTR tidak
// berubah. Yang diubah hanya cara tugas itu tersebar pada kesembilan petak.
//
// Tiga pola disiapkan, sebab sebaran yang merata belum menentukan segalanya:
// yang menentukan adalah apakah keterlambatan menumpuk pada petak tertentu.
//
//	independen    : peluang telat sama pada seluruh petak
//	condong-tinggi: telat menumpuk pada tugas bernilai tinggi
//	condong-rendah: telat menumpuk pada tugas bernilai rendah

import (
	"math"
	"sort"
)

// petakF adalah petak dengan cacah pecahan, dipakai agar pembagian merata
// tidak terganggu galat pembulatan.
type petakF struct {
	Priority string
	Impact   string
	NAll     float64
	NDone    float64
	NOnTime  float64
}

type polaTelat int

const (
	polaIndependen polaTelat = iota
	polaCondongTinggi
	polaCondongRendah
)

func namaPola(p polaTelat) string {
	switch p {
	case polaCondongTinggi:
		return "condong-tinggi"
	case polaCondongRendah:
		return "condong-rendah"
	default:
		return "independen"
	}
}

// bangunMerata menyebar nAll tugas rata ke sembilan petak, lalu menempatkan
// tugas belum selesai secara merata dan tugas telat menurut pola yang dipilih.
func bangunMerata(nAll, nDone, nOnTime float64, pola polaTelat, w skemaBobot) []petakF {
	tingkat := []string{"HIGH", "MEDIUM", "LOW"}
	var sel []petakF
	for _, pr := range tingkat {
		for _, im := range tingkat {
			sel = append(sel, petakF{Priority: pr, Impact: im, NAll: nAll / 9})
		}
	}

	belum := nAll - nDone
	telat := nDone - nOnTime

	// tugas belum selesai selalu disebar merata, agar hanya pola keterlambatan
	// yang berbeda antar-skenario
	sisa := make([]float64, len(sel))
	for i := range sel {
		sel[i].NDone = sel[i].NAll - belum/9
		sisa[i] = sel[i].NDone
	}

	urut := make([]int, len(sel))
	for i := range urut {
		urut[i] = i
	}
	switch pola {
	case polaCondongTinggi:
		sort.SliceStable(urut, func(a, b int) bool {
			return w.nilai(sel[urut[a]].Priority, sel[urut[a]].Impact) >
				w.nilai(sel[urut[b]].Priority, sel[urut[b]].Impact)
		})
	case polaCondongRendah:
		sort.SliceStable(urut, func(a, b int) bool {
			return w.nilai(sel[urut[a]].Priority, sel[urut[a]].Impact) <
				w.nilai(sel[urut[b]].Priority, sel[urut[b]].Impact)
		})
	}

	if pola == polaIndependen {
		for i := range sel {
			t := telat / 9
			sel[i].NOnTime = sel[i].NDone - t
		}
	} else {
		lTersisa := telat
		telatPer := make([]float64, len(sel))
		for _, i := range urut {
			ambil := math.Min(lTersisa, sisa[i])
			telatPer[i] = ambil
			lTersisa -= ambil
			if lTersisa <= 1e-9 {
				break
			}
		}
		for i := range sel {
			sel[i].NOnTime = sel[i].NDone - telatPer[i]
		}
	}
	return sel
}

func jumlahBobotF(ps []petakF, w skemaBobot) (wAll, wDone, wOnTime float64) {
	for _, q := range ps {
		n := w.nilai(q.Priority, q.Impact)
		wAll += n * q.NAll
		wDone += n * q.NDone
		wOnTime += n * q.NOnTime
	}
	return
}

func tvsKreditF(ps []petakF, w skemaBobot, k skemaKredit) float64 {
	var atas, bawah float64
	for _, q := range ps {
		n := w.nilai(q.Priority, q.Impact)
		bawah += n * q.NAll
		atas += n * (q.NOnTime + k.kredit(q.Priority, q.Impact)*(q.NDone-q.NOnTime))
	}
	if bawah == 0 {
		return 0
	}
	return atas / bawah * 100
}

func expMerata() {
	header("EKSPERIMEN 15 - BILA SEBARAN PRIORITAS DAN DAMPAK DIBUAT MERATA")

	petakUser := muatPetak()
	w := skemaUji()[1]                  // bobot sistem 1 / 0,9 / 0,8
	cfg := cfgVarian(varianKaidah()[1]) // rancangan sistem: agregasi saja

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

	// cacah asli tiap karyawan
	asli := map[string][3]float64{} // nAll, nDone, nOnTime
	for n, ps := range petakUser {
		var a, d, o float64
		for _, q := range ps {
			a += float64(q.NAll)
			d += float64(q.NDone)
			o += float64(q.NOnTime)
		}
		asli[n] = [3]float64{a, d, o}
	}

	polaSemua := []polaTelat{polaIndependen, polaCondongTinggi, polaCondongRendah}

	// ---------------------------------------------------------- 15.1
	p("\n15.1 Sebaran sebelum dan sesudah dibuat merata\n")
	var real9 [9]float64
	tingkat := []string{"HIGH", "MEDIUM", "LOW"}
	totReal := 0.0
	for _, ps := range petakUser {
		for _, q := range ps {
			for i, pr := range tingkat {
				for j, im := range tingkat {
					if q.Priority == pr && q.Impact == im {
						real9[i*3+j] += float64(q.NAll)
						totReal += float64(q.NAll)
					}
				}
			}
		}
	}
	p("%-16s %14s %14s\n", "Prioritas/Dampak", "data asli", "data merata")
	for i, pr := range tingkat {
		for j, im := range tingkat {
			pct := 0.0
			if totReal > 0 {
				pct = real9[i*3+j] / totReal * 100
			}
			p("%-16s %9.0f %4.2f%% %9.0f %5.2f%%\n", pr+"/"+im,
				real9[i*3+j], pct, totReal/9, 100.0/9)
		}
	}

	// ---------------------------------------------------------- 15.2
	p("\n15.2 Daya gerak bobot pada TVS rumus sekarang (tanpa kredit telat)\n")
	p("Rentang TVS ketika lima skema bobot yang sangat berbeda dicobakan.\n\n")
	p("%-22s %10s", "Karyawan", "asli")
	for _, pl := range polaSemua {
		p(" %16s", "merata "+namaPola(pl))
	}
	p("\n")
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		p("%-22s", trunc(n, 22))
		// data asli
		mn, mx := math.Inf(1), math.Inf(-1)
		for _, sk := range skemaUji() {
			wa, wd, _ := jumlahBobot(ps, sk)
			v := 0.0
			if wa > 0 {
				v = wd / wa * 100
			}
			mn, mx = math.Min(mn, v), math.Max(mx, v)
		}
		p(" %10.4f", mx-mn)
		// data merata
		for _, pl := range polaSemua {
			m := bangunMerata(asli[n][0], asli[n][1], asli[n][2], pl, w)
			mn, mx = math.Inf(1), math.Inf(-1)
			for _, sk := range skemaUji() {
				wa, wd, _ := jumlahBobotF(m, sk)
				v := 0.0
				if wa > 0 {
					v = wd / wa * 100
				}
				mn, mx = math.Min(mn, v), math.Max(mx, v)
			}
			p(" %16.4f", mx-mn)
		}
		p("\n")
	}
	p("\nAngka adalah rentang, yaitu selisih TVS tertinggi dan terendah antar-skema bobot.\n")

	// ---------------------------------------------------------- 15.3
	p("\n15.3 Pertanyaan pokok: apakah kredit di tabel dampak kini berbeda dari di prioritas\n")
	for _, pl := range polaSemua {
		p("\n  pola keterlambatan: %s\n", namaPola(pl))
		p("  %-22s", "Karyawan")
		for _, k := range skemaKreditUji() {
			p(" %10s", k.nama)
		}
		p(" %12s\n", "|C-B|")
		var jumBedaCB, jumBedaDB, jumBedaEB float64
		var cacah int
		for _, n := range nama {
			if len(petakUser[n]) == 0 {
				continue
			}
			m := bangunMerata(asli[n][0], asli[n][1], asli[n][2], pl, w)
			p("  %-22s", trunc(n, 22))
			nilai := map[string]float64{}
			for _, k := range skemaKreditUji() {
				v := tvsKreditF(m, w, k)
				nilai[k.nama] = v
				p(" %10.4f", v)
			}
			p(" %12.4f\n", math.Abs(nilai["C-dampak"]-nilai["B-prio"]))
			jumBedaCB += math.Abs(nilai["C-dampak"] - nilai["B-prio"])
			jumBedaDB += math.Abs(nilai["D-kali"] - nilai["B-prio"])
			jumBedaEB += math.Abs(nilai["E-min"] - nilai["B-prio"])
			cacah++
		}
		if cacah > 0 {
			p("  selisih mutlak rata-rata terhadap B: C=%.4f  D=%.4f  E=%.4f poin\n",
				jumBedaCB/float64(cacah), jumBedaDB/float64(cacah), jumBedaEB/float64(cacah))
		}
	}

	// ---------------------------------------------------------- 15.4
	p("\n15.4 Kategori akhir pada data merata\n")
	for _, pl := range polaSemua {
		p("\n  pola keterlambatan: %s\n", namaPola(pl))
		p("  %-22s", "Karyawan")
		for _, k := range skemaKreditUji() {
			p(" %-14s", k.nama)
		}
		p("\n")
		bedaB := map[string]int{}
		for _, n := range nama {
			if len(petakUser[n]) == 0 {
				continue
			}
			m := bangunMerata(asli[n][0], asli[n][1], asli[n][2], pl, w)
			k0 := lain[n]
			p("  %-22s", trunc(n, 22))
			kat := map[string]string{}
			for _, k := range skemaKreditUji() {
				_, c, _, _ := cfg.Infer([]float64{k0.TCR, k0.OTR, tvsKreditF(m, w, k), k0.WER})
				kat[k.nama] = c
				p(" %-14s", trunc(c, 14))
			}
			for _, k := range skemaKreditUji() {
				if kat[k.nama] != kat["B-prio"] {
					bedaB[k.nama]++
				}
			}
			p("\n")
		}
		p("  berbeda kategori dari B-prio:")
		for _, k := range skemaKreditUji() {
			p(" %s=%d", k.nama, bedaB[k.nama])
		}
		p("\n")
	}

	// ---------------------------------------------------------- 15.5
	p("\n15.5 Apakah sebaran merata menyembuhkan kemubaziran TVS terhadap TCR\n")
	p("TVS rumus sekarang, tanpa kredit telat.\n\n")
	p("%-22s %8s %10s", "Karyawan", "TCR", "asli")
	for _, pl := range polaSemua {
		p(" %16s", namaPola(pl))
	}
	p("\n")
	var vTCR []float64
	seri := map[string][]float64{}
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		vTCR = append(vTCR, lain[n].TCR)
		p("%-22s %8.2f", trunc(n, 22), lain[n].TCR)
		wa, wd, _ := jumlahBobot(ps, w)
		v := 0.0
		if wa > 0 {
			v = wd / wa * 100
		}
		seri["asli"] = append(seri["asli"], v)
		p(" %10.4f", v)
		for _, pl := range polaSemua {
			m := bangunMerata(asli[n][0], asli[n][1], asli[n][2], pl, w)
			wa, wd, _ := jumlahBobotF(m, w)
			v := 0.0
			if wa > 0 {
				v = wd / wa * 100
			}
			seri[namaPola(pl)] = append(seri[namaPola(pl)], v)
			p(" %16.4f", v)
		}
		p("\n")
	}
	p("\n%-20s %14s %18s\n", "Data", "Spearman TCR", "rerata |beda| TCR")
	for _, kunci := range []string{"asli", "independen", "condong-tinggi", "condong-rendah"} {
		if len(seri[kunci]) == 0 {
			continue
		}
		p("%-20s %14.4f %18.4f\n", kunci, spearman(seri[kunci], vTCR), rerataBeda(seri[kunci], vTCR))
	}

	// ---------------------------------------------------------- 15.6
	p("\n15.6 Berapa banyak tugas telat yang tersentuh dimensi dampak\n")
	p("%-20s %14s %14s\n", "Data", "telat impact HIGH", "telat bukan HIGH")
	var tH, tL float64
	for _, ps := range petakUser {
		for _, q := range ps {
			t := float64(q.NDone - q.NOnTime)
			if q.Impact == "HIGH" {
				tH += t
			} else {
				tL += t
			}
		}
	}
	p("%-20s %13.0f%% %13.0f%%\n", "asli", tH/(tH+tL)*100, tL/(tH+tL)*100)
	for _, pl := range polaSemua {
		var h, l float64
		for _, n := range nama {
			if len(petakUser[n]) == 0 {
				continue
			}
			m := bangunMerata(asli[n][0], asli[n][1], asli[n][2], pl, w)
			for _, q := range m {
				t := q.NDone - q.NOnTime
				if q.Impact == "HIGH" {
					h += t
				} else {
					l += t
				}
			}
		}
		if h+l > 0 {
			p("%-20s %13.0f%% %13.0f%%\n", "merata "+namaPola(pl), h/(h+l)*100, l/(h+l)*100)
		}
	}
	// ---------------------------------------------------------- 15.7
	p("\n15.7 Daya gerak bobot setelah kredit telat dipakai, pada data merata\n")
	p("Rentang TVS ketika lima skema bobot dicobakan, memakai kredit skema B.\n\n")
	kB := skemaKreditUji()[1]
	p("%-22s %10s", "Karyawan", "asli")
	for _, pl := range polaSemua {
		p(" %16s", namaPola(pl))
	}
	p("\n")
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		p("%-22s", trunc(n, 22))
		mn, mx := math.Inf(1), math.Inf(-1)
		for _, sk := range skemaUji() {
			v := tvsKredit(ps, sk, kB)
			mn, mx = math.Min(mn, v), math.Max(mx, v)
		}
		p(" %10.4f", mx-mn)
		for _, pl := range polaSemua {
			m := bangunMerata(asli[n][0], asli[n][1], asli[n][2], pl, w)
			mn, mx = math.Inf(1), math.Inf(-1)
			for _, sk := range skemaUji() {
				v := tvsKreditF(m, sk, kB)
				mn, mx = math.Min(mn, v), math.Max(mx, v)
			}
			p(" %16.4f", mx-mn)
		}
		p("\n")
	}

	// ---------------------------------------------------------- 15.8
	p("\n15.8 Kesetaraan aljabar pada data merata dengan keterlambatan independen\n")
	p("Karena tiap petak berisi cacah yang sama, jumlah pada pembilang dapat\n")
	p("dipisah menjadi hasil kali dua jumlah yang berdiri sendiri.\n\n")
	kC := skemaKreditUji()[2]
	sw, scwP, swI, scwI := 0.0, 0.0, 0.0, 0.0
	for _, t := range tingkat {
		sw += w.prio[t]
		scwP += kB.kredit(t, "HIGH") * w.prio[t]
		swI += w.impak[t]
		scwI += kC.kredit("HIGH", t) * w.impak[t]
	}
	p("  jumlah bobot prioritas            = %.4f\n", sw)
	p("  jumlah bobot dampak               = %.4f\n", swI)
	p("  jumlah (kredit x bobot) prioritas = %.4f\n", scwP)
	p("  jumlah (kredit x bobot) dampak    = %.4f\n", scwI)
	p("\n  kredit dipasang di prioritas : %.4f x %.4f = %.4f\n", scwP, swI, scwP*swI)
	p("  kredit dipasang di dampak    : %.4f x %.4f = %.4f\n", sw, scwI, sw*scwI)
	if math.Abs(scwP*swI-sw*scwI) < 1e-9 {
		p("\n  Keduanya sama persis, sehingga TVS yang dihasilkan identik.\n")
		p("  Kesamaan ini muncul karena kedua tabel memakai deret bobot dan deret\n")
		p("  kredit yang sama. Bila salah satu tabel memakai deret yang berbeda,\n")
		p("  kesamaan ini tidak lagi berlaku.\n")
	} else {
		p("\n  Keduanya berbeda, sehingga penempatan kredit berpengaruh.\n")
	}
}
