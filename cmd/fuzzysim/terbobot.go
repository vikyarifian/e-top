package main

// Eksperimen 16: bagaimana bila keempat indikator dihitung terbobot?
//
// Saat ini hanya TVS yang memakai bobot. TCR, OTR, dan WER menghitung tugas
// secara setara. Eksperimen ini menyusun versi terbobot dari ketiganya:
//
//	TCR terbobot = SUM(w tugas selesai)      / SUM(w seluruh tugas)
//	OTR terbobot = SUM(w tugas tepat waktu)  / SUM(w tugas selesai)
//	WER terbobot = SUM(w x rasio)            / SUM(w)          , hanya tugas berjam
//
// dengan w = bobot prioritas x bobot dampak, sama seperti TVS.
//
// Pengukuran dilakukan pada data nyata dan pada data yang sebarannya dibuat
// merata, sebab keduanya menjawab pertanyaan yang berbeda.

import (
	"math"
	"sort"
)

// tugasJam adalah satu tugas selesai yang memiliki jam estimasi dan aktual.
type tugasJam struct {
	User     string
	Priority string
	Impact   string
	Rasio    float64
}

func muatRasio() map[string][]tugasJam {
	db := openDB()
	var baris []tugasJam
	db.Raw(`SELECT u.full_name AS user, tp.priority AS priority, ti.impact AS impact,
			(t.estimated_hours / NULLIF(t.actual_hours,0)) * 100 AS rasio
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		JOIN task_priorities tp ON tp.no = t.priority_id
		JOIN task_impacts ti ON ti.no = t.impact_id
		WHERE t.completed_at IS NOT NULL
		  AND t.estimated_hours > 0 AND t.actual_hours > 0`).Scan(&baris)
	out := map[string][]tugasJam{}
	for _, b := range baris {
		out[b.User] = append(out[b.User], b)
	}
	return out
}

func batas100(x float64) float64 {
	if x > 100 {
		return 100
	}
	return x
}

// indikator menampung keempat nilai dalam satu wadah.
type indikator struct{ TCR, OTR, TVS, WER float64 }

// hitungNyata menghitung indikator versi biasa dan versi terbobot dari data
// nyata satu karyawan.
func hitungNyata(ps []petak, jam []tugasJam, w skemaBobot) (biasa, bobot indikator) {
	var nAll, nDone, nOn float64
	var wAll, wDone, wOn float64
	for _, q := range ps {
		n := w.nilai(q.Priority, q.Impact)
		nAll += float64(q.NAll)
		nDone += float64(q.NDone)
		nOn += float64(q.NOnTime)
		wAll += n * float64(q.NAll)
		wDone += n * float64(q.NDone)
		wOn += n * float64(q.NOnTime)
	}
	if nAll > 0 {
		biasa.TCR = nDone / nAll * 100
	}
	if nDone > 0 {
		biasa.OTR = nOn / nDone * 100
	}
	if wAll > 0 {
		bobot.TCR = wDone / wAll * 100
		biasa.TVS = wDone / wAll * 100
		bobot.TVS = biasa.TVS
	}
	if wDone > 0 {
		bobot.OTR = wOn / wDone * 100
	}

	var sRasio, cRasio, sWRasio, sW float64
	for _, t := range jam {
		n := w.nilai(t.Priority, t.Impact)
		sRasio += t.Rasio
		cRasio++
		sWRasio += n * t.Rasio
		sW += n
	}
	if cRasio > 0 {
		biasa.WER = batas100(sRasio / cRasio)
	}
	if sW > 0 {
		bobot.WER = batas100(sWRasio / sW)
	}
	return
}

// hitungMerata menghitung indikator pada data yang sebarannya dibuat merata.
// Rasio efisiensi diurutkan lalu dibagi menjadi sembilan potongan yang
// ditempatkan pada petak menurut pola yang dipilih.
func hitungMerata(nAll, nDone, nOn float64, rasio []float64, pola polaTelat, w skemaBobot) (biasa, bobot indikator) {
	sel := bangunMerata(nAll, nDone, nOn, pola, w)

	var sNAll, sNDone, sNOn, sWAll, sWDone, sWOn float64
	for _, q := range sel {
		n := w.nilai(q.Priority, q.Impact)
		sNAll += q.NAll
		sNDone += q.NDone
		sNOn += q.NOnTime
		sWAll += n * q.NAll
		sWDone += n * q.NDone
		sWOn += n * q.NOnTime
	}
	if sNAll > 0 {
		biasa.TCR = sNDone / sNAll * 100
	}
	if sNDone > 0 {
		biasa.OTR = sNOn / sNDone * 100
	}
	if sWAll > 0 {
		bobot.TCR = sWDone / sWAll * 100
		biasa.TVS = bobot.TCR
		bobot.TVS = bobot.TCR
	}
	if sWDone > 0 {
		bobot.OTR = sWOn / sWDone * 100
	}

	// efisiensi: urutkan rasio, bagi sembilan, tempatkan menurut pola
	if len(rasio) > 0 {
		r := append([]float64{}, rasio...)
		sort.Float64s(r)
		urutPetak := make([]int, len(sel))
		for i := range urutPetak {
			urutPetak[i] = i
		}
		switch pola {
		case polaCondongTinggi: // rasio terburuk pada petak bernilai tertinggi
			sort.SliceStable(urutPetak, func(a, b int) bool {
				return w.nilai(sel[urutPetak[a]].Priority, sel[urutPetak[a]].Impact) >
					w.nilai(sel[urutPetak[b]].Priority, sel[urutPetak[b]].Impact)
			})
		case polaCondongRendah:
			sort.SliceStable(urutPetak, func(a, b int) bool {
				return w.nilai(sel[urutPetak[a]].Priority, sel[urutPetak[a]].Impact) <
					w.nilai(sel[urutPetak[b]].Priority, sel[urutPetak[b]].Impact)
			})
		}
		var sR, cR, sWR, sW float64
		per := float64(len(r)) / 9
		for k, idx := range urutPetak {
			n := w.nilai(sel[idx].Priority, sel[idx].Impact)
			mulai := int(float64(k) * per)
			akhir := int(float64(k+1) * per)
			if k == 8 {
				akhir = len(r)
			}
			if pola == polaIndependen {
				// setiap petak mendapat campuran yang sama
				mulai, akhir = 0, len(r)
				for _, x := range r {
					sR += x / 9
					sWR += n * x / 9
				}
				cR += float64(len(r)) / 9
				sW += n * float64(len(r)) / 9
				continue
			}
			for _, x := range r[mulai:akhir] {
				sR += x
				sWR += n * x
			}
			cR += float64(akhir - mulai)
			sW += n * float64(akhir-mulai)
		}
		if cR > 0 {
			biasa.WER = batas100(sR / cR)
		}
		if sW > 0 {
			bobot.WER = batas100(sWR / sW)
		}
	}
	return
}

func expTerbobot() {
	header("EKSPERIMEN 16 - BILA SELURUH INDIKATOR DIHITUNG TERBOBOT")

	petakUser := muatPetak()
	jamUser := muatRasio()
	w := skemaUji()[1]
	cfg := cfgVarian(varianKaidah()[1]) // rancangan sistem: agregasi saja

	nama := []string{}
	for _, r := range rows() {
		if r.Period == "GAB" {
			nama = append(nama, r.User)
		}
	}
	sort.Strings(nama)

	asli := map[string][3]float64{}
	for n, ps := range petakUser {
		var a, d, o float64
		for _, q := range ps {
			a += float64(q.NAll)
			d += float64(q.NDone)
			o += float64(q.NOnTime)
		}
		asli[n] = [3]float64{a, d, o}
	}

	// ---------------------------------------------------------- 16.1
	p("\n16.1 Rumus yang dibandingkan\n")
	p("  TCR biasa    = cacah selesai / cacah seluruh tugas\n")
	p("  TCR terbobot = SUM(w selesai) / SUM(w seluruh tugas)\n")
	p("  OTR biasa    = cacah tepat waktu / cacah selesai\n")
	p("  OTR terbobot = SUM(w tepat waktu) / SUM(w selesai)\n")
	p("  WER biasa    = rata-rata rasio estimasi terhadap aktual\n")
	p("  WER terbobot = rata-rata rasio yang ditimbang nilai tugas\n")
	p("  dengan w = bobot prioritas x bobot dampak\n")

	// ---------------------------------------------------------- 16.2
	p("\n16.2 Data nyata: indikator biasa dan terbobot\n")
	p("%-22s %8s %8s %9s %8s %8s %9s %8s %8s %9s\n",
		"Karyawan", "TCR", "TCRw", "selisih", "OTR", "OTRw", "selisih", "WER", "WERw", "selisih")
	var vTCR, vTCRw, vOTR, vOTRw, vTVS, vWER, vWERw []float64
	nyata := map[string][2]indikator{}
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		b, bw := hitungNyata(ps, jamUser[n], w)
		nyata[n] = [2]indikator{b, bw}
		p("%-22s %8.2f %8.2f %+9.2f %8.2f %8.2f %+9.2f %8.2f %8.2f %+9.2f\n",
			trunc(n, 22), b.TCR, bw.TCR, bw.TCR-b.TCR,
			b.OTR, bw.OTR, bw.OTR-b.OTR, b.WER, bw.WER, bw.WER-b.WER)
		vTCR = append(vTCR, b.TCR)
		vTCRw = append(vTCRw, bw.TCR)
		vOTR = append(vOTR, b.OTR)
		vOTRw = append(vOTRw, bw.OTR)
		vTVS = append(vTVS, b.TVS)
		vWER = append(vWER, b.WER)
		vWERw = append(vWERw, bw.WER)
	}

	// ---------------------------------------------------------- 16.3
	p("\n16.3 Temuan pokok: TCR terbobot sama dengan TVS\n")
	p("%-22s %12s %12s %12s\n", "Karyawan", "TCR terbobot", "TVS", "selisih")
	maksBeda := 0.0
	for _, n := range nama {
		if len(petakUser[n]) == 0 {
			continue
		}
		bw := nyata[n][1]
		b := nyata[n][0]
		p("%-22s %12.6f %12.6f %12.6f\n", trunc(n, 22), bw.TCR, b.TVS, math.Abs(bw.TCR-b.TVS))
		maksBeda = math.Max(maksBeda, math.Abs(bw.TCR-b.TVS))
	}
	p("\nselisih terbesar: %.10f\n", maksBeda)
	p("Keduanya rumus yang sama persis. Membobot TCR berarti menghitung ulang TVS,\n")
	p("sehingga sistem akan memiliki dua indikator yang identik.\n")

	// ---------------------------------------------------------- 16.4
	p("\n16.4 Daya gerak bobot pada tiap indikator, data nyata\n")
	p("Rentang nilai ketika lima skema bobot yang berbeda dicobakan.\n\n")
	p("%-22s %12s %12s %12s\n", "Karyawan", "TCR terbobot", "OTR terbobot", "WER terbobot")
	for _, n := range nama {
		ps := petakUser[n]
		if len(ps) == 0 {
			continue
		}
		var mnT, mxT, mnO, mxO, mnW, mxW = math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
		for _, sk := range skemaUji() {
			_, bw := hitungNyata(ps, jamUser[n], sk)
			mnT, mxT = math.Min(mnT, bw.TCR), math.Max(mxT, bw.TCR)
			mnO, mxO = math.Min(mnO, bw.OTR), math.Max(mxO, bw.OTR)
			mnW, mxW = math.Min(mnW, bw.WER), math.Max(mxW, bw.WER)
		}
		p("%-22s %12.4f %12.4f %12.4f\n", trunc(n, 22), mxT-mnT, mxO-mnO, mxW-mnW)
	}

	// ---------------------------------------------------------- 16.5
	p("\n16.5 Kemiripan antar-indikator, sebelum dan sesudah dibobot\n")
	p("Korelasi peringkat Spearman; 1,0000 berarti urutan karyawannya sama persis.\n\n")
	pasangan := []struct {
		ket  string
		a, b []float64
	}{
		{"TCR biasa    vs TVS", vTCR, vTVS},
		{"TCR terbobot vs TVS", vTCRw, vTVS},
		{"OTR biasa    vs OTR terbobot", vOTR, vOTRw},
		{"WER biasa    vs WER terbobot", vWER, vWERw},
		{"OTR terbobot vs TVS", vOTRw, vTVS},
		{"OTR terbobot vs WER terbobot", vOTRw, vWERw},
	}
	p("%-34s %12s %16s\n", "Pasangan", "Spearman", "rerata |beda|")
	for _, ps := range pasangan {
		p("%-34s %12.4f %16.4f\n", ps.ket, spearman(ps.a, ps.b), rerataBeda(ps.a, ps.b))
	}

	// ---------------------------------------------------------- 16.6
	p("\n16.6 Nilai akhir dan kategori\n")
	p("%-22s %9s %-14s %9s %-14s %s\n",
		"Karyawan", "Z biasa", "kategori", "Z terbobot", "kategori", "sama?")
	samaKat, totKat := 0, 0
	var zA, zB []float64
	for _, n := range nama {
		if len(petakUser[n]) == 0 {
			continue
		}
		b, bw := nyata[n][0], nyata[n][1]
		z1, c1, _, _ := cfg.Infer([]float64{b.TCR, b.OTR, b.TVS, b.WER})
		z2, c2, _, _ := cfg.Infer([]float64{bw.TCR, bw.OTR, bw.TVS, bw.WER})
		tanda := "ya"
		if c1 != c2 {
			tanda = "TIDAK"
		} else {
			samaKat++
		}
		totKat++
		zA = append(zA, z1)
		zB = append(zB, z2)
		p("%-22s %9.2f %-14s %9.2f %-14s %s\n", trunc(n, 22), z1, c1, z2, c2, tanda)
	}
	p("\nkategori sepakat: %d dari %d ; Spearman %.4f ; rerata |beda Z| %.2f\n",
		samaKat, totKat, spearman(zA, zB), rerataBeda(zA, zB))

	// ---------------------------------------------------------- 16.7
	p("\n16.7 Data dengan sebaran merata\n")
	for _, pl := range []polaTelat{polaIndependen, polaCondongTinggi, polaCondongRendah} {
		p("\n  pola: %s\n", namaPola(pl))
		p("  %-22s %8s %8s %9s %8s %8s %9s %8s %8s %9s\n",
			"Karyawan", "TCR", "TCRw", "selisih", "OTR", "OTRw", "selisih", "WER", "WERw", "selisih")
		var mTCR, mTCRw, mOTR, mOTRw, mWER, mWERw []float64
		for _, n := range nama {
			if len(petakUser[n]) == 0 {
				continue
			}
			var r []float64
			for _, t := range jamUser[n] {
				r = append(r, t.Rasio)
			}
			b, bw := hitungMerata(asli[n][0], asli[n][1], asli[n][2], r, pl, w)
			p("  %-22s %8.2f %8.2f %+9.2f %8.2f %8.2f %+9.2f %8.2f %8.2f %+9.2f\n",
				trunc(n, 22), b.TCR, bw.TCR, bw.TCR-b.TCR,
				b.OTR, bw.OTR, bw.OTR-b.OTR, b.WER, bw.WER, bw.WER-b.WER)
			mTCR = append(mTCR, b.TCR)
			mTCRw = append(mTCRw, bw.TCR)
			mOTR = append(mOTR, b.OTR)
			mOTRw = append(mOTRw, bw.OTR)
			mWER = append(mWER, b.WER)
			mWERw = append(mWERw, bw.WER)
		}
		p("  rerata |beda| : TCR %.4f  OTR %.4f  WER %.4f\n",
			rerataBeda(mTCR, mTCRw), rerataBeda(mOTR, mOTRw), rerataBeda(mWER, mWERw))
	}
}
