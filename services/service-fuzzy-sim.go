package services

// Mesin inferensi Fuzzy Tsukamoto yang dapat dikonfigurasi, dipakai oleh
// halaman Simulasi. Berbeda dengan services/service-fuzzy.go yang memakai
// rancangan tetap (tiga himpunan, simpul 40-60-90, 81 aturan), mesin di sini
// menerima bentuk fungsi keanggotaan, parameter, jumlah himpunan, dan jumlah
// variabel input sebagai masukan sehingga perilakunya dapat diuji.
//
// Konsekuen tetap monoton pada seluruh konfigurasi karena hal itu merupakan
// syarat metode Tsukamoto: nilai crisp tiap aturan diperoleh dari invers
// fungsi keanggotaan konsekuen, dan invers hanya bernilai tunggal bila
// fungsinya monoton.

import (
	"fmt"
	"math"
	"time"
)

// Bentuk fungsi keanggotaan yang tersedia.
const (
	MFLinearTurun = "linear-turun"
	MFLinearNaik  = "linear-naik"
	MFSegitiga    = "segitiga"
	MFTrapesium   = "trapesium"
	MFGaussian    = "gaussian"
	MFGaussKiri   = "gauss-kiri"
	MFGaussKanan  = "gauss-kanan"
)

// ShapeOption menjelaskan satu bentuk beserta jumlah dan nama parameternya.
type ShapeOption struct {
	Value  string
	Label  string
	Params []string
}

// ShapeOptions adalah daftar bentuk yang dapat dipilih pengguna.
var ShapeOptions = []ShapeOption{
	{MFLinearTurun, "Descending linear (left shoulder)", []string{"a", "b"}},
	{MFLinearNaik, "Ascending linear (right shoulder)", []string{"a", "b"}},
	{MFSegitiga, "Triangular", []string{"a", "b", "c"}},
	{MFTrapesium, "Trapezoidal", []string{"a", "b", "c", "d"}},
	{MFGaussian, "Gaussian", []string{"center", "sigma"}},
	{MFGaussKiri, "Gaussian left shoulder", []string{"center", "sigma"}},
	{MFGaussKanan, "Gaussian right shoulder", []string{"center", "sigma"}},
}

// ShapeParams mengembalikan nama parameter untuk sebuah bentuk.
func ShapeParams(shape string) []string {
	for _, s := range ShapeOptions {
		if s.Value == shape {
			return s.Params
		}
	}
	return []string{"a", "b"}
}

// ShapeLabel mengembalikan label manusiawi untuk sebuah bentuk.
func ShapeLabel(shape string) string {
	for _, s := range ShapeOptions {
		if s.Value == shape {
			return s.Label
		}
	}
	return shape
}

// SimSet adalah satu himpunan linguistik beserta bentuk dan parameternya.
type SimSet struct {
	Label  string
	Shape  string
	Params []float64
}

// Degree menghitung derajat keanggotaan sebuah nilai crisp pada himpunan ini.
func (s SimSet) Degree(x float64) float64 {
	at := func(i int) float64 {
		if i < len(s.Params) {
			return s.Params[i]
		}
		return 0
	}
	switch s.Shape {
	case MFLinearTurun:
		a, b := at(0), at(1)
		switch {
		case b <= a:
			if x <= a {
				return 1
			}
			return 0
		case x <= a:
			return 1
		case x >= b:
			return 0
		default:
			return (b - x) / (b - a)
		}

	case MFLinearNaik:
		a, b := at(0), at(1)
		switch {
		case b <= a:
			if x >= b {
				return 1
			}
			return 0
		case x <= a:
			return 0
		case x >= b:
			return 1
		default:
			return (x - a) / (b - a)
		}

	case MFSegitiga:
		a, b, c := at(0), at(1), at(2)
		switch {
		case x <= a || x >= c:
			return 0
		case x == b:
			return 1
		case x < b:
			if b == a {
				return 1
			}
			return (x - a) / (b - a)
		default:
			if c == b {
				return 1
			}
			return (c - x) / (c - b)
		}

	case MFTrapesium:
		a, b, c, d := at(0), at(1), at(2), at(3)
		switch {
		case x < a || x > d:
			return 0
		case x >= b && x <= c:
			return 1
		case x < b:
			if b == a {
				return 1
			}
			return (x - a) / (b - a)
		default:
			if d == c {
				return 1
			}
			return (d - x) / (d - c)
		}

	case MFGaussian, MFGaussKiri, MFGaussKanan:
		c, sig := at(0), at(1)
		if sig <= 0 {
			if x == c {
				return 1
			}
			return 0
		}
		if s.Shape == MFGaussKiri && x <= c {
			return 1
		}
		if s.Shape == MFGaussKanan && x >= c {
			return 1
		}
		d := x - c
		return math.Exp(-(d * d) / (2 * sig * sig))
	}
	return 0
}

// SimVarDef mendefinisikan satu variabel input beserta nilai crisp-nya.
type SimVarDef struct {
	Name  string
	Label string
	Value float64
}

// SimConfig adalah konfigurasi lengkap sebuah simulasi. Susunan himpunan
// berlaku sama untuk seluruh variabel, sebagaimana rancangan pada aplikasi.
type SimConfig struct {
	Vars []SimVarDef
	Sets []SimSet
}

// SimSetTrace memuat derajat keanggotaan satu himpunan pada satu variabel.
type SimSetTrace struct {
	Label  string
	Shape  string
	Degree float64
}

// SimVarTrace memuat hasil fuzzifikasi satu variabel.
type SimVarTrace struct {
	Name  string
	Label string
	Value float64
	Sets  []SimSetTrace
	SumMu float64
}

// SimRule adalah satu aturan pada basis aturan.
type SimRule struct {
	Code    string
	SetIdx  []int
	Output  string
	OutIdx  int
	Alpha   float64
	Z       float64
	Premise string
}

// SimResult memuat seluruh jejak perhitungan satu simulasi.
type SimResult struct {
	Fuzzification []SimVarTrace
	Active        []SimRule
	RuleTotal     int
	SumAlpha      float64
	SumAlphaZ     float64
	Score         float64
	Category      string
	NanosPerInfer float64
}

// buildSimRules membangkitkan basis aturan memakai kaidah yang sama dengan
// services/service-fuzzy.go: skor tiap himpunan dijumlah, lalu rasionya
// terhadap skor maksimum dipetakan ke lima konsekuen dengan lebar pita sama.
func buildSimRules(cfg SimConfig) []SimRule {
	nVar, nSet := len(cfg.Vars), len(cfg.Sets)
	if nVar == 0 || nSet == 0 {
		return nil
	}
	total := 1
	for i := 0; i < nVar; i++ {
		total *= nSet
	}
	rules := make([]SimRule, 0, total)
	for k := 0; k < total; k++ {
		idx := make([]int, nVar)
		rem := k
		for i := nVar - 1; i >= 0; i-- {
			// dibalik agar aturan pertama adalah kombinasi seluruh himpunan tertinggi
			idx[i] = nSet - 1 - rem%nSet
			rem /= nSet
		}
		sum := 0
		for _, v := range idx {
			sum += v
		}
		level := 0
		if nSet > 1 {
			level = int(float64(sum) / float64(nVar*(nSet-1)) * 5)
		}
		if level > 4 {
			level = 4
		}

		premise := ""
		for i := range cfg.Vars {
			if i > 0 {
				premise += " dan "
			}
			premise += cfg.Vars[i].Name + " " + cfg.Sets[idx[i]].Label
		}
		rules = append(rules, SimRule{
			Code:    fmt.Sprintf("R%d", k+1),
			SetIdx:  idx,
			Output:  FuzzyOutputNames[level],
			OutIdx:  level,
			Premise: premise,
		})
	}
	return rules
}

// SimRuleBase mengembalikan basis aturan untuk sebuah konfigurasi.
func SimRuleBase(cfg SimConfig) []SimRule { return buildSimRules(cfg) }

// RunSimulation menjalankan seluruh tahapan Fuzzy Tsukamoto pada konfigurasi
// yang diberikan dan mengembalikan jejaknya secara lengkap.
func RunSimulation(cfg SimConfig) SimResult {
	res := SimResult{}
	nVar, nSet := len(cfg.Vars), len(cfg.Sets)
	if nVar == 0 || nSet == 0 {
		return res
	}

	// tahap 1: fuzzifikasi
	mu := make([][]float64, nVar)
	for i, v := range cfg.Vars {
		mu[i] = make([]float64, nSet)
		tr := SimVarTrace{Name: v.Name, Label: v.Label, Value: v.Value}
		for j, s := range cfg.Sets {
			d := s.Degree(v.Value)
			mu[i][j] = d
			tr.SumMu += d
			tr.Sets = append(tr.Sets, SimSetTrace{Label: s.Label, Shape: s.Shape, Degree: d})
		}
		res.Fuzzification = append(res.Fuzzification, tr)
	}

	// tahap 2: inferensi dengan fungsi implikasi MIN
	rules := buildSimRules(cfg)
	res.RuleTotal = len(rules)
	for _, r := range rules {
		alpha := 1.0
		for i := 0; i < nVar; i++ {
			if v := mu[i][r.SetIdx[i]]; v < alpha {
				alpha = v
			}
		}
		if alpha <= 0 {
			continue
		}
		r.Alpha = alpha
		r.Z = fuzzyZ(r.Output, alpha)
		res.Active = append(res.Active, r)
		res.SumAlpha += alpha
		res.SumAlphaZ += alpha * r.Z
	}

	// tahap 3: defuzzifikasi rata-rata terbobot
	if res.SumAlpha > 0 {
		res.Score = res.SumAlphaZ / res.SumAlpha
	}
	res.Category = fuzzyCategory(res.Score)

	// Pengukuran beban komputasi. Jumlah pengulangan ditambah sampai waktu yang
	// terukur melampaui ambang, karena basis aturan yang kecil selesai lebih
	// cepat daripada resolusi pewaktu sistem sehingga akan terbaca nol.
	const minElapsed = 2 * time.Millisecond
	const maxIter = 200000
	iter := 0
	t0 := time.Now()
	acc := 0.0
	for {
		sa, saz := 0.0, 0.0
		for _, r := range rules {
			alpha := 1.0
			for i := 0; i < nVar; i++ {
				if v := mu[i][r.SetIdx[i]]; v < alpha {
					alpha = v
				}
			}
			if alpha <= 0 {
				continue
			}
			z := fuzzyZ(r.Output, alpha)
			sa += alpha
			saz += alpha * z
		}
		if sa > 0 {
			acc += saz / sa
		}
		iter++
		if iter >= maxIter || (iter%64 == 0 && time.Since(t0) >= minElapsed) {
			break
		}
	}
	_ = acc
	if iter > 0 {
		res.NanosPerInfer = float64(time.Since(t0).Nanoseconds()) / float64(iter)
	}
	return res
}

// SimScore mengembalikan nilai akhir dan kategorinya saja, dipakai untuk
// tabel perbandingan.
func SimScore(cfg SimConfig) (float64, string) {
	r := RunSimulation(cfg)
	return r.Score, r.Category
}

// SimCurve menghasilkan derajat keanggotaan sebuah himpunan pada semesta
// 0 sampai 100 untuk digambar sebagai grafik.
func SimCurve(s SimSet, steps int) []float64 {
	if steps < 2 {
		steps = 2
	}
	pts := make([]float64, steps+1)
	for i := 0; i <= steps; i++ {
		pts[i] = s.Degree(float64(i) / float64(steps) * 100)
	}
	return pts
}
