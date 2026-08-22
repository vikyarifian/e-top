package main

// Mesin inferensi Fuzzy Tsukamoto yang dapat dikonfigurasi.
// Dipakai untuk eksperimen Bab 4 TA: variasi bentuk fungsi keanggotaan
// (bahu linear, segitiga, trapesium, gaussian) dan variasi jumlah variabel input.
// Konsekuen tetap monoton (syarat Tsukamoto) pada semua konfigurasi
// agar perbandingan hanya menyoroti sisi anteseden.

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ---------- fungsi keanggotaan dasar ----------

func triangle(a, b, c float64) func(float64) float64 {
	return func(x float64) float64 {
		switch {
		case x <= a || x >= c:
			return 0
		case x == b:
			return 1
		case x < b:
			return (x - a) / (b - a)
		default:
			return (c - x) / (c - b)
		}
	}
}

func trapezoid(a, b, c, d float64) func(float64) float64 {
	return func(x float64) float64 {
		switch {
		case x <= a || x >= d:
			return 0
		case x >= b && x <= c:
			return 1
		case x < b:
			return (x - a) / (b - a)
		default:
			return (d - x) / (d - c)
		}
	}
}

// bahu kiri: 1 sampai b lalu turun linear menuju 0 di c
func shoulderLeft(b, c float64) func(float64) float64 {
	return func(x float64) float64 {
		switch {
		case x <= b:
			return 1
		case x >= c:
			return 0
		default:
			return (c - x) / (c - b)
		}
	}
}

// bahu kanan: 0 sampai a lalu naik linear menuju 1 di b
func shoulderRight(a, b float64) func(float64) float64 {
	return func(x float64) float64 {
		switch {
		case x <= a:
			return 0
		case x >= b:
			return 1
		default:
			return (x - a) / (b - a)
		}
	}
}

func gaussian(c, sigma float64) func(float64) float64 {
	return func(x float64) float64 {
		d := x - c
		return math.Exp(-(d * d) / (2 * sigma * sigma))
	}
}

// gaussian bahu: rata (=1) di sisi luar, meluruh gaussian di sisi dalam
func gaussShoulderLeft(c, sigma float64) func(float64) float64 {
	g := gaussian(c, sigma)
	return func(x float64) float64 {
		if x <= c {
			return 1
		}
		return g(x)
	}
}

func gaussShoulderRight(c, sigma float64) func(float64) float64 {
	g := gaussian(c, sigma)
	return func(x float64) float64 {
		if x >= c {
			return 1
		}
		return g(x)
	}
}

// ---------- konfigurasi ----------

type FuzzySet struct {
	Label string
	Mu    func(float64) float64
}

type Config struct {
	Name     string
	Desc     string
	Sets     []FuzzySet   // urut dari tingkat terendah ke tertinggi
	PerVar   [][]FuzzySet // opsional: himpunan berbeda untuk tiap variabel (kalibrasi)
	VarNames []string     // variabel input yang dipakai
	Rules    []Rule
}

// setsOf mengembalikan himpunan yang berlaku untuk variabel ke-i
func (c *Config) setsOf(i int) []FuzzySet {
	if c.PerVar != nil {
		return c.PerVar[i]
	}
	return c.Sets
}

type Rule struct {
	Code    string
	SetIdx  []int // indeks himpunan per variabel
	Output  string
	OutIdx  int
}

type FiredRule struct {
	Code   string
	Output string
	Alpha  float64
	Z      float64
	SetIdx []int
}

var outLabels = []string{"Sangat Buruk", "Buruk", "Cukup", "Baik", "Sangat Baik"}

// invers fungsi keanggotaan konsekuen monoton (identik dengan services/service-fuzzy.go)
func consequentZ(outIdx int, alpha float64) float64 {
	switch outIdx {
	case 4: // Sangat Baik, monoton naik pada 80-100
		return 80 + 20*alpha
	case 3: // Baik, monoton naik pada 60-80
		return 60 + 20*alpha
	case 2: // Cukup, monoton naik pada 40-60
		return 40 + 20*alpha
	case 1: // Buruk, monoton turun pada 20-40
		return 40 - 20*alpha
	default: // Sangat Buruk, monoton turun pada 0-20
		return 20 * (1 - alpha)
	}
}

func Category(z float64) string {
	switch {
	case z > 80:
		return "Sangat Baik"
	case z > 60:
		return "Baik"
	case z > 40:
		return "Cukup"
	case z > 20:
		return "Buruk"
	default:
		return "Sangat Buruk"
	}
}

// ---------- pembangkit basis aturan ----------
//
// Prinsip: derajat kebaikan setiap kombinasi diukur dari rasio
//   r = (jumlah skor himpunan seluruh variabel) / (n_variabel * (m_himpunan - 1))
// lalu dipetakan ke lima kategori konsekuen dengan lebar pita yang sama.
// Kaidah tambahan: bila OTR berada pada himpunan terendah, konsekuen
// dibatasi maksimum "Cukup" karena disiplin tenggat adalah syarat perlu.
// Untuk 4 variabel dan 2 himpunan, kaidah ini menghasilkan tepat 16 aturan
// yang identik dengan tabel aturan pada services/service-fuzzy.go.

func buildRules(varNames []string, nSets int) []Rule {
	n := len(varNames)
	otrPos := -1
	for i, v := range varNames {
		if v == "OTR" {
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
		// dekode k menjadi kombinasi indeks himpunan (variabel pertama paling signifikan,
		// indeks dibalik agar himpunan tertinggi muncul lebih dulu seperti pada tabel TA)
		rem := k
		for i := n - 1; i >= 0; i-- {
			idx[i] = nSets - 1 - rem%nSets
			rem /= nSets
		}
		sum := 0
		for _, v := range idx {
			sum += v
		}
		r := float64(sum) / float64(n*(nSets-1))
		level := int(r * 5)
		if level > 4 {
			level = 4
		}
		if otrPos >= 0 && idx[otrPos] == 0 && level > 2 {
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

// ---------- inferensi ----------

// Infer mengembalikan nilai crisp, kategori, aturan yang menyala,
// dan jumlah aturan dengan alpha efektif (>= 0.001).
func (c *Config) Infer(x []float64) (float64, string, []FiredRule, int) {
	n := len(c.VarNames)
	mu := make([][]float64, n)
	for i := 0; i < n; i++ {
		sets := c.setsOf(i)
		mu[i] = make([]float64, len(sets))
		for j, s := range sets {
			mu[i][j] = s.Mu(x[i])
		}
	}
	fired := []FiredRule{}
	sumA, sumAZ := 0.0, 0.0
	eff := 0
	for _, r := range c.Rules {
		alpha := 1.0
		for i := 0; i < n; i++ {
			if v := mu[i][r.SetIdx[i]]; v < alpha {
				alpha = v
			}
		}
		if alpha <= 0 {
			continue
		}
		z := consequentZ(r.OutIdx, alpha)
		fired = append(fired, FiredRule{Code: r.Code, Output: r.Output, Alpha: alpha, Z: z, SetIdx: append([]int{}, r.SetIdx...)})
		sumA += alpha
		sumAZ += alpha * z
		if alpha >= 0.001 {
			eff++
		}
	}
	if sumA == 0 {
		return 0, Category(0), fired, eff
	}
	z := sumAZ / sumA
	return z, Category(z), fired, eff
}

func (c *Config) Score(x []float64) float64 {
	z, _, _, _ := c.Infer(x)
	return z
}

// ---------- definisi konfigurasi yang diuji ----------

var defaultVars = []string{"TCR", "OTR", "TPS", "WER"}

func setsBahu() []FuzzySet {
	return []FuzzySet{
		{"Rendah", shoulderLeft(40, 60)},
		{"Tinggi", shoulderRight(40, 60)},
	}
}

func setsBahuLebar() []FuzzySet {
	return []FuzzySet{
		{"Rendah", shoulderLeft(20, 80)},
		{"Tinggi", shoulderRight(20, 80)},
	}
}

// Simpul rancangan sistem: P1 batas atas Rendah penuh, P2 puncak Sedang,
// P3 batas bawah Tinggi penuh. Ketiga bentuk pembanding memakai simpul
// yang sama agar perbandingan hanya menyoroti bentuk kurva.
const (
	sP1 = 40.0
	sP2 = 60.0
	sP3 = 90.0
)

// Segitiga: rancangan sistem (partisi Ruspini, jumlah derajat keanggotaan = 1)
func setsSegitiga() []FuzzySet {
	return []FuzzySet{
		{"Rendah", shoulderLeft(sP1, sP2)},
		{"Sedang", triangle(sP1, sP2, sP3)},
		{"Tinggi", shoulderRight(sP2, sP3)},
	}
}

// Segitiga dengan partisi seragam pada seluruh semesta, sebagai pembanding
func setsSegitigaSeragam() []FuzzySet {
	return []FuzzySet{
		{"Rendah", shoulderLeft(0, 50)},
		{"Sedang", triangle(0, 50, 100)},
		{"Tinggi", shoulderRight(50, 100)},
	}
}

// Trapesium: puncak dilebarkan menjadi bidang datar di sekitar simpul yang sama
func setsTrapesium() []FuzzySet {
	return []FuzzySet{
		{"Rendah", trapezoid(-1e-9, 30, sP1, sP2)},
		{"Sedang", trapezoid(sP1, 55, 70, sP3)},
		{"Tinggi", trapezoid(sP2, 80, 100, 100 + 1e-9)},
	}
}

// Lebar sigma ditetapkan agar perpotongan antar himpunan bertetangga
// berada di sekitar derajat keanggotaan 0,5 (setengah jarak dibagi 1,1774).
const (
	gSigmaRendah = 8.49
	gSigmaSedang = 10.62
	gSigmaTinggi = 12.74
)

func setsGaussian() []FuzzySet {
	return []FuzzySet{
		{"Rendah", gaussShoulderLeft(sP1, gSigmaRendah)},
		{"Sedang", gaussian(sP2, gSigmaSedang)},
		{"Tinggi", gaussShoulderRight(sP3, gSigmaTinggi)},
	}
}

// baseline mengembalikan konfigurasi rancangan sistem yang berlaku pada aplikasi
func baseline() *Config {
	return newConfig("B", "Segitiga 3 himpunan P=(40, 60, 90) - rancangan sistem", setsSegitiga(), defaultVars)
}

func newConfig(name, desc string, sets []FuzzySet, vars []string) *Config {
	c := &Config{Name: name, Desc: desc, Sets: sets, VarNames: vars}
	c.Rules = buildRules(vars, len(sets))
	return c
}

func mfConfigs() []*Config {
	return []*Config{
		newConfig("A", "Bahu linear 2 himpunan (rancangan awal), transisi 40-60", setsBahu(), defaultVars),
		baseline(),
		newConfig("C", "Trapesium 3 himpunan, simpul sama dengan B", setsTrapesium(), defaultVars),
		newConfig("D", "Gaussian 3 himpunan, pusat sama dengan B", setsGaussian(), defaultVars),
		newConfig("E", "Segitiga 3 himpunan partisi seragam P=(0, 50, 100)", setsSegitigaSeragam(), defaultVars),
		newConfig("A2", "Bahu linear 2 himpunan, transisi 20-80", setsBahuLebar(), defaultVars),
	}
}

func inputConfigs() []*Config {
	return []*Config{
		newConfig("N2", "2 variabel: TCR, OTR", setsSegitiga(), []string{"TCR", "OTR"}),
		newConfig("N3", "3 variabel: TCR, OTR, TPS", setsSegitiga(), []string{"TCR", "OTR", "TPS"}),
		newConfig("N4", "4 variabel: TCR, OTR, TPS, WER (rancangan sistem)", setsSegitiga(), []string{"TCR", "OTR", "TPS", "WER"}),
		newConfig("N5", "5 variabel: TCR, OTR, TPS, WER, WLR", setsSegitiga(), []string{"TCR", "OTR", "TPS", "WER", "WLR"}),
	}
}

// ---------- utilitas ----------

func (c *Config) RuleTable() string {
	var b strings.Builder
	for _, r := range c.Rules {
		parts := make([]string, len(r.SetIdx))
		for i, s := range r.SetIdx {
			parts[i] = fmt.Sprintf("%s %s", c.VarNames[i], c.setsOf(i)[s].Label)
		}
		fmt.Fprintf(&b, "%-5s IF %s THEN Kinerja %s\n", r.Code, strings.Join(parts, " AND "), r.Output)
	}
	return b.String()
}

func spearman(a, b []float64) float64 {
	rank := func(v []float64) []float64 {
		type p struct {
			val float64
			idx int
		}
		ps := make([]p, len(v))
		for i, x := range v {
			ps[i] = p{x, i}
		}
		sort.Slice(ps, func(i, j int) bool { return ps[i].val < ps[j].val })
		r := make([]float64, len(v))
		i := 0
		for i < len(ps) {
			j := i
			for j+1 < len(ps) && ps[j+1].val == ps[i].val {
				j++
			}
			avg := float64(i+j)/2 + 1
			for k := i; k <= j; k++ {
				r[ps[k].idx] = avg
			}
			i = j + 1
		}
		return r
	}
	ra, rb := rank(a), rank(b)
	n := float64(len(a))
	var ma, mb float64
	for i := range ra {
		ma += ra[i]
		mb += rb[i]
	}
	ma /= n
	mb /= n
	var num, da, db float64
	for i := range ra {
		x, y := ra[i]-ma, rb[i]-mb
		num += x * y
		da += x * x
		db += y * y
	}
	if da == 0 || db == 0 {
		return math.NaN()
	}
	return num / math.Sqrt(da*db)
}
