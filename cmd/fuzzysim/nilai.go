package main

// Eksperimen 18: dua pertanyaan yang saling berkaitan.
//
//	A. Bagaimana bila OTR sendiri dihitung berjenjang seperti DSS, yaitu
//	   menimbang seberapa dalam keterlambatannya, bukan sekadar telat atau tidak?
//	B. Bagaimana cara lain menghitung nilai sebuah tugas, agar indikator ketiga
//	   benar-benar bergerak?
//
// Keduanya berkaitan karena bila slot kedua sudah memuat ketepatan waktu yang
// berjenjang, slot ketiga harus benar-benar berbicara tentang nilai pekerjaan.
//
// Kolom yang tersedia untuk menyusun nilai tugas ternyata sangat terbatas:
// seluruh tugas bertipe TICKET, sedangkan subtask, lampiran, tag, dan watcher
// tidak ada satu pun. Yang benar-benar beragam hanya jam estimasi dan panjang
// deskripsi.

import (
	"math"
	"sort"
)

// tugasNilai adalah satu tugas beserta bahan yang dipakai menyusun nilainya.
type tugasNilai struct {
	User     string
	W        float64 // bobot prioritas x bobot dampak
	Est      float64 // jam estimasi
	DescLen  float64 // panjang deskripsi
	Done     bool
	OnTime   bool
	Severity float64 // 0 tepat waktu, 1 terlambat penuh atau telantar
	Judged   bool
}

func muatTugasNilai() []tugasNilai {
	db := openDB()
	var baris []struct {
		User     string
		W        float64
		Est      float64
		DescLen  float64
		Done     bool
		OnTime   bool
		Severity float64
		Judged   bool
	}
	db.Raw(`SELECT u.full_name AS user,
			tp.weight * ti.weight AS w,
			COALESCE(t.estimated_hours, 0) AS est,
			LENGTH(COALESCE(t.description, '')) AS desc_len,
			(t.completed_at IS NOT NULL) AS done,
			(t.completed_at IS NOT NULL AND t.completed_at <= t.due_date) AS on_time,
			CASE
				WHEN t.completed_at IS NOT NULL AND t.completed_at <= t.due_date THEN 0
				WHEN t.completed_at IS NOT NULL THEN LEAST(1.0,
					EXTRACT(EPOCH FROM (t.completed_at - t.due_date)) / 60.0
					/ NULLIF(GREATEST(tp.max_due_minutes,
						EXTRACT(EPOCH FROM (t.due_date - COALESCE(t.start_date, t.created_at)))/60.0), 0))
				ELSE 1
			END AS severity,
			(t.completed_at IS NOT NULL OR (t.due_date IS NOT NULL AND t.due_date < now())) AS judged
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		JOIN task_priorities tp ON tp.no = t.priority_id
		JOIN task_impacts ti ON ti.no = t.impact_id`).Scan(&baris)
	out := make([]tugasNilai, 0, len(baris))
	for _, b := range baris {
		out = append(out, tugasNilai(b))
	}
	return out
}

// fungsiNilai adalah satu cara menghitung nilai sebuah tugas.
type fungsiNilai struct {
	kode string
	ket  string
	f    func(t tugasNilai, medEst, medDesc float64) float64
}

func daftarFungsiNilai() []fungsiNilai {
	return []fungsiNilai{
		{"N1", "w prioritas x w dampak (berlaku sekarang)",
			func(t tugasNilai, _, _ float64) float64 { return t.W }},
		{"N2", "setiap tugas bernilai sama",
			func(t tugasNilai, _, _ float64) float64 { return 1 }},
		{"N3", "w x jam estimasi",
			func(t tugasNilai, me, _ float64) float64 { return t.W * bagi(t.Est, me) }},
		{"N4", "w x akar jam estimasi",
			func(t tugasNilai, me, _ float64) float64 { return t.W * math.Sqrt(bagi(t.Est, me)) }},
		{"N5", "w x logaritma jam estimasi",
			func(t tugasNilai, me, _ float64) float64 {
				return t.W * math.Log1p(bagi(t.Est, me)) / math.Log(2)
			}},
		{"N6", "jam estimasi saja",
			func(t tugasNilai, me, _ float64) float64 { return bagi(t.Est, me) }},
		{"N7", "w x panjang deskripsi",
			func(t tugasNilai, _, md float64) float64 { return t.W * bagi(t.DescLen, md) }},
		{"N8", "w x akar (jam estimasi x panjang deskripsi)",
			func(t tugasNilai, me, md float64) float64 {
				return t.W * math.Sqrt(bagi(t.Est, me)*bagi(t.DescLen, md))
			}},
	}
}

// sebaranNilai merangkum satu fungsi nilai pada seluruh tugas.
type sebaranNilai struct {
	min, maks, rata float64
	rasio           float64
}

func expNilai() {
	header("EKSPERIMEN 18 - OTR BERJENJANG DAN CARA LAIN MENILAI TUGAS")

	semua := muatTugasNilai()
	if len(semua) == 0 {
		p("\ntidak ada data\n")
		return
	}

	// nilai tengah sebagai penormal
	var est, desc []float64
	for _, t := range semua {
		if t.Est > 0 {
			est = append(est, t.Est)
		}
		if t.DescLen > 0 {
			desc = append(desc, t.DescLen)
		}
	}
	sort.Float64s(est)
	sort.Float64s(desc)
	medEst := est[len(est)/2]
	medDesc := desc[len(desc)/2]

	perUser := map[string][]tugasNilai{}
	for _, t := range semua {
		perUser[t.User] = append(perUser[t.User], t)
	}
	nama := []string{}
	for n := range perUser {
		nama = append(nama, n)
	}
	sort.Strings(nama)

	// ============================================================ Bagian A
	p("\n=== BAGIAN A: OTR BERJENJANG ===\n")
	p("\n18.1 OTR biner dibanding OTR berjenjang\n")
	p("OTR biner     = cacah tepat waktu / cacah tugas yang sudah pasti\n")
	p("OTR berjenjang = 1 dikurangi rata-rata kedalaman keterlambatan\n\n")
	p("%-22s %10s %12s %10s\n", "Karyawan", "OTR biner", "OTR jenjang", "selisih")
	var vBiner, vJenjang []float64
	for _, n := range nama {
		var nJud, nOn, sSev float64
		for _, t := range perUser[n] {
			if !t.Judged {
				continue
			}
			nJud++
			if t.OnTime {
				nOn++
			}
			sSev += t.Severity
		}
		b := bagi(nOn, nJud) * 100
		j := (1 - bagi(sSev, nJud)) * 100
		vBiner = append(vBiner, b)
		vJenjang = append(vJenjang, j)
		p("%-22s %10.2f %12.2f %+10.2f\n", trunc(n, 22), b, j, j-b)
	}
	p("\nkorelasi peringkat: %.4f ; rerata selisih %.2f poin\n",
		spearman(vBiner, vJenjang), rerataBeda(vBiner, vJenjang))
	p("Rentang: biner %.2f, berjenjang %.2f\n",
		rentang(vBiner), rentang(vJenjang))

	// ============================================================ Bagian B
	p("\n=== BAGIAN B: CARA LAIN MENILAI TUGAS ===\n")

	fn := daftarFungsiNilai()

	p("\n18.2 Sebaran nilai tugas menurut tiap fungsi, seluruh %d tugas\n", len(semua))
	p("%-5s %-46s %10s %10s %10s %12s\n",
		"Kode", "Rumus nilai tugas", "min", "maks", "rata", "rasio maks:min")
	for _, f := range fn {
		mn, mx, sum := math.Inf(1), math.Inf(-1), 0.0
		for _, t := range semua {
			v := f.f(t, medEst, medDesc)
			mn, mx = math.Min(mn, v), math.Max(mx, v)
			sum += v
		}
		rasio := 0.0
		if mn > 0 {
			rasio = mx / mn
		}
		p("%-5s %-46s %10.4f %10.4f %10.4f %12.1f\n",
			f.kode, trunc(f.ket, 46), mn, mx, sum/float64(len(semua)), rasio)
	}

	p("\n18.3 Indikator yang dihasilkan tiap fungsi nilai\n")
	p("TVS  = nilai tugas selesai / nilai seluruh tugas\n")
	p("VDS  = nilai tugas tepat waktu / nilai seluruh tugas\n")
	p("DSSv = 1 dikurangi (nilai x kedalaman telat) / nilai seluruh tugas\n")

	hasil := map[string]map[string][]float64{}
	for _, f := range fn {
		hasil[f.kode] = map[string][]float64{}
		for _, n := range nama {
			var tot, done, on, sev float64
			for _, t := range perUser[n] {
				v := f.f(t, medEst, medDesc)
				tot += v
				if t.Done {
					done += v
				}
				if t.OnTime {
					on += v
				}
				if t.Judged {
					sev += v * t.Severity
				}
			}
			hasil[f.kode]["TVS"] = append(hasil[f.kode]["TVS"], bagi(done, tot)*100)
			hasil[f.kode]["VDS"] = append(hasil[f.kode]["VDS"], bagi(on, tot)*100)
			hasil[f.kode]["DSSv"] = append(hasil[f.kode]["DSSv"], (1-bagi(sev, tot))*100)
		}
	}

	for _, ind := range []string{"TVS", "VDS", "DSSv"} {
		p("\n  %s menurut fungsi nilai:\n", ind)
		p("  %-22s", "Karyawan")
		for _, f := range fn {
			p(" %8s", f.kode)
		}
		p("\n")
		for i, n := range nama {
			p("  %-22s", trunc(n, 22))
			for _, f := range fn {
				p(" %8.2f", hasil[f.kode][ind][i])
			}
			p("\n")
		}
		p("  %-22s", "rentang")
		for _, f := range fn {
			p(" %8.2f", rentang(hasil[f.kode][ind]))
		}
		p("\n  %-22s", "nilai unik")
		for _, f := range fn {
			p(" %8d", unik(hasil[f.kode][ind]))
		}
		p("\n")
	}

	// ---------------------------------------------------------- 18.4
	p("\n18.4 Ringkasan: daya pembeda dan kebaruan tiap gabungan\n")
	p("%-5s %-5s %-42s %9s %7s %10s %10s\n",
		"Nilai", "Ind", "Rumus nilai tugas", "rentang", "unik", "r vs TCR", "r vs OTRj")
	type calon struct {
		kode, ind string
		rentang   float64
		unik      int
		maksR     float64
	}
	var calons []calon
	var vTCR []float64
	for _, n := range nama {
		var nAll, nDone float64
		for _, t := range perUser[n] {
			nAll++
			if t.Done {
				nDone++
			}
		}
		vTCR = append(vTCR, bagi(nDone, nAll)*100)
	}
	for _, f := range fn {
		for _, ind := range []string{"TVS", "VDS", "DSSv"} {
			v := hasil[f.kode][ind]
			rT := math.Abs(spearman(v, vTCR))
			rO := math.Abs(spearman(v, vJenjang))
			p("%-5s %-5s %-42s %9.2f %7d %10.4f %10.4f\n",
				f.kode, ind, trunc(f.ket, 42), rentang(v), unik(v), rT, rO)
			calons = append(calons, calon{f.kode, ind, rentang(v), unik(v), math.Max(rT, rO)})
		}
	}

	p("\n18.5 Peringkat gabungan yang paling layak menempati slot ketiga\n")
	p("Syarat: nilai unik minimal 8, lalu diurutkan menurut rentang x (1 - korelasi tertinggi).\n\n")
	sort.Slice(calons, func(i, j int) bool {
		si := calons[i].rentang * (1 - calons[i].maksR)
		sj := calons[j].rentang * (1 - calons[j].maksR)
		return si > sj
	})
	p("%-5s %-6s %9s %7s %10s %10s\n", "Nilai", "Ind", "rentang", "unik", "maks r", "skor")
	n := 0
	for _, c := range calons {
		if c.unik < 8 {
			continue
		}
		p("%-5s %-6s %9.2f %7d %10.4f %10.4f\n",
			c.kode, c.ind, c.rentang, c.unik, c.maksR, c.rentang*(1-c.maksR))
		n++
		if n >= 12 {
			break
		}
	}
}

func rentang(v []float64) float64 {
	mn, mx := math.Inf(1), math.Inf(-1)
	for _, x := range v {
		mn, mx = math.Min(mn, x), math.Max(mx, x)
	}
	return mx - mn
}

func unik(v []float64) int {
	u := map[float64]bool{}
	for _, x := range v {
		u[math.Round(x*100)/100] = true
	}
	return len(u)
}
