package services

// Inferensi Fuzzy Tsukamoto untuk penilaian kinerja.
//
// Variabel input TCR, OTR, TPS, dan WER bersemesta 0-100 dan masing-masing
// memiliki tiga himpunan linguistik: Rendah, Sedang, dan Tinggi. Simpul
// fungsi keanggotaan ditetapkan pada P = (40, 60, 90) sehingga ketiga
// himpunan membentuk partisi yang derajat keanggotaannya selalu berjumlah 1.
//
// Konsekuen tiap aturan monoton, sehingga nilai crisp tiap aturan diperoleh
// dari invers fungsi keanggotaan output, lalu digabung dengan rata-rata
// terbobot: Z = sum(alpha_i * z_i) / sum(alpha_i).

import "fmt"

// Indeks himpunan linguistik pada tiap variabel input.
const (
	SetRendah = 0
	SetSedang = 1
	SetTinggi = 2
)

// Simpul fungsi keanggotaan. P1 batas atas Rendah penuh, P2 puncak Sedang,
// P3 batas bawah Tinggi penuh.
const (
	P1 = 40.0
	P2 = 60.0
	P3 = 90.0
)

// FuzzyVarNames adalah urutan variabel input yang dipakai pada premis aturan.
var FuzzyVarNames = [4]string{"TCR", "OTR", "TPS", "WER"}

// FuzzySetNames adalah label himpunan linguistik sesuai indeksnya.
var FuzzySetNames = [3]string{"Rendah", "Sedang", "Tinggi"}

// FuzzyOutputNames adalah label konsekuen dari terendah ke tertinggi.
var FuzzyOutputNames = [5]string{"Sangat Buruk", "Buruk", "Cukup", "Baik", "Sangat Baik"}

type FuzzyRule struct {
	Code    string
	Premise [4]int // urutan TCR, OTR, TPS, WER; nilai SetRendah/SetSedang/SetTinggi
	Output  string // Sangat Baik / Baik / Cukup / Buruk / Sangat Buruk
	Alpha   float64
	Z       float64
}

// PremiseLabels mengembalikan label himpunan tiap variabel pada premis aturan.
func (r FuzzyRule) PremiseLabels() [4]string {
	var out [4]string
	for i, s := range r.Premise {
		out[i] = FuzzySetNames[s]
	}
	return out
}

// Description menyusun bentuk IF-THEN aturan untuk ditampilkan pada antarmuka.
func (r FuzzyRule) Description() string {
	s := "IF"
	for i, v := range FuzzyVarNames {
		if i > 0 {
			s += " AND"
		}
		s += fmt.Sprintf(" %s %s", v, FuzzySetNames[r.Premise[i]])
	}
	return s + " THEN Kinerja " + r.Output
}

// fuzzyRules dibangkitkan sekali saat inisialisasi paket: 3^4 = 81 aturan.
var fuzzyRules = buildFuzzyRules()

// buildFuzzyRules membangkitkan basis aturan dari kaidah berikut.
//
// Derajat kebaikan tiap kombinasi diukur dari rasio skor himpunan
//
//	r = jumlah skor himpunan seluruh variabel / (jumlah variabel * 2)
//
// dengan skor Rendah = 0, Sedang = 1, dan Tinggi = 2. Rasio tersebut
// dipetakan ke lima konsekuen dengan lebar pita yang sama. Selain itu
// berlaku kaidah pembatas: bila OTR berada pada himpunan Rendah, konsekuen
// dibatasi paling tinggi "Cukup", karena disiplin terhadap tenggat merupakan
// syarat perlu bagi penilaian di atas cukup.
func buildFuzzyRules() []FuzzyRule {
	const nVar, nSet = 4, 3
	total := nSet * nSet * nSet * nSet
	rules := make([]FuzzyRule, 0, total)
	for k := 0; k < total; k++ {
		var premise [4]int
		rem := k
		for i := nVar - 1; i >= 0; i-- {
			// indeks dibalik agar aturan pertama adalah kombinasi seluruh Tinggi
			premise[i] = nSet - 1 - rem%nSet
			rem /= nSet
		}
		sum := 0
		for _, v := range premise {
			sum += v
		}
		level := int(float64(sum) / float64(nVar*(nSet-1)) * 5)
		if level > 4 {
			level = 4
		}
		// OTR adalah variabel kedua
		if premise[1] == SetRendah && level > 2 {
			level = 2
		}
		rules = append(rules, FuzzyRule{
			Code:    fmt.Sprintf("R%d", k+1),
			Premise: premise,
			Output:  FuzzyOutputNames[level],
		})
	}
	return rules
}

// FuzzyRules mengembalikan salinan basis aturan untuk keperluan dokumentasi.
func FuzzyRules() []FuzzyRule {
	out := make([]FuzzyRule, len(fuzzyRules))
	copy(out, fuzzyRules)
	return out
}

// FuzzyMembership menghitung derajat keanggotaan sebuah nilai crisp
// terhadap himpunan Rendah, Sedang, dan Tinggi.
func FuzzyMembership(x float64) [3]float64 {
	var mu [3]float64
	switch {
	case x <= P1:
		mu[SetRendah] = 1
	case x < P2:
		mu[SetRendah] = (P2 - x) / (P2 - P1)
	}
	switch {
	case x <= P1 || x >= P3:
		mu[SetSedang] = 0
	case x < P2:
		mu[SetSedang] = (x - P1) / (P2 - P1)
	default:
		mu[SetSedang] = (P3 - x) / (P3 - P2)
	}
	switch {
	case x >= P3:
		mu[SetTinggi] = 1
	case x > P2:
		mu[SetTinggi] = (x - P2) / (P3 - P2)
	}
	return mu
}

// fuzzyZ adalah invers fungsi keanggotaan konsekuen yang monoton, dihitung
// pada nilai alpha aturan.
func fuzzyZ(output string, alpha float64) float64 {
	switch output {
	case "Sangat Baik": // monoton naik pada 80-100
		return 80 + 20*alpha
	case "Baik": // monoton naik pada 60-80
		return 60 + 20*alpha
	case "Cukup": // monoton naik pada 40-60
		return 40 + 20*alpha
	case "Buruk": // monoton turun pada 20-40
		return 40 - 20*alpha
	default: // Sangat Buruk, monoton turun pada 0-20
		return 20 * (1 - alpha)
	}
}

func fuzzyCategory(z float64) string {
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

// FuzzyTsukamoto menghitung nilai kinerja 0-100 beserta kategorinya
// dan mengembalikan daftar aturan aktif (alpha > 0) untuk kebutuhan audit.
func FuzzyTsukamoto(tcr, otr, tps, wer float64) (float64, string, []FuzzyRule) {
	var mu [4][3]float64
	for i, v := range []float64{tcr, otr, tps, wer} {
		mu[i] = FuzzyMembership(v)
	}

	active := []FuzzyRule{}
	sumAlpha, sumAlphaZ := 0.0, 0.0
	for _, r := range fuzzyRules {
		alpha := 1.0
		for i := 0; i < 4; i++ {
			if v := mu[i][r.Premise[i]]; v < alpha {
				alpha = v
			}
		}
		if alpha <= 0 {
			continue
		}
		r.Alpha = alpha
		r.Z = fuzzyZ(r.Output, alpha)
		active = append(active, r)
		sumAlpha += alpha
		sumAlphaZ += alpha * r.Z
	}

	if sumAlpha == 0 {
		return 0, fuzzyCategory(0), active
	}
	z := sumAlphaZ / sumAlpha
	return z, fuzzyCategory(z), active
}
