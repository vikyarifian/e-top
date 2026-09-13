package features

import (
	"fmt"
	"sort"
	"strings"

	"etop/services"
)

// Ukuran bidang gambar grafik fungsi keanggotaan.
const (
	simChartW    = 640.0
	simChartH    = 200.0
	simPadLeft   = 34.0
	simPadRight  = 12.0
	simPadTop    = 12.0
	simPadBottom = 26.0
)

func simPlotW() float64 { return simChartW - simPadLeft - simPadRight }
func simPlotH() float64 { return simChartH - simPadTop - simPadBottom }

// simX memetakan nilai semesta 0-100 ke koordinat mendatar.
func simX(v float64) float64 { return simPadLeft + v/100*simPlotW() }

// simY memetakan derajat keanggotaan 0-1 ke koordinat tegak.
func simY(d float64) float64 { return simPadTop + (1-d)*simPlotH() }

// simCurvePoints menghasilkan daftar titik polyline untuk satu himpunan.
func simCurvePoints(s services.SimSet) string {
	const steps = 200
	pts := services.SimCurve(s, steps)
	var b strings.Builder
	for i, d := range pts {
		fmt.Fprintf(&b, "%.2f,%.2f ", simX(float64(i)/steps*100), simY(d))
	}
	return strings.TrimSpace(b.String())
}

// simSetPalette adalah warna garis untuk tiap himpunan, berurutan.
var simSetPalette = []string{"#ef4444", "#f59e0b", "#22c55e", "#3b82f6", "#8b5cf6"}

func simSetColor(i int) string { return simSetPalette[i%len(simSetPalette)] }

// simVarPalette adalah warna penanda nilai tiap variabel input.
var simVarPalette = []string{"#0ea5e9", "#e11d48", "#7c3aed", "#059669", "#d97706"}

func simVarColor(i int) string { return simVarPalette[i%len(simVarPalette)] }

// simNum memformat bilangan desimal dengan dua angka di belakang koma.
func simNum(v float64) string { return fmt.Sprintf("%.2f", v) }

// simMu memformat derajat keanggotaan dengan empat angka di belakang koma.
func simMu(v float64) string { return fmt.Sprintf("%.4f", v) }

// simIsZero menandai derajat keanggotaan yang tidak berpengaruh.
func simIsZero(v float64) bool { return v <= 0 }

// simRuleShare menghitung sumbangan sebuah aturan terhadap nilai akhir,
// dinyatakan dalam persen dari jumlah alpha.
func simRuleShare(alpha, sumAlpha float64) float64 {
	if sumAlpha <= 0 {
		return 0
	}
	return alpha / sumAlpha * 100
}

// simPresetLabel mengembalikan label prasetel, termasuk keadaan tersuai.
func simPresetLabel(v string) string {
	if v == "custom" {
		return "Konfigurasi tersuai"
	}
	if p, ok := services.SimPresetByValue(v); ok {
		return p.Label
	}
	return v
}

// simParamValue mengambil parameter ke-k sebuah himpunan dengan aman.
func simParamValue(s services.SimSet, k int) float64 {
	if k < len(s.Params) {
		return s.Params[k]
	}
	return 0
}

// simSelected membantu menandai pilihan terpilih pada elemen select.
func simSelected(a, b string) bool { return a == b }

// simDelta menghitung selisih sebuah baris perbandingan terhadap nilai acuan.
func simDelta(v, ref float64) string {
	d := v - ref
	switch {
	case d > 0.005:
		return fmt.Sprintf("+%.2f", d)
	case d < -0.005:
		return fmt.Sprintf("%.2f", d)
	default:
		return "0,00"
	}
}

// simScoreRef mencari nilai acuan pada daftar perbandingan, yaitu baris yang
// sedang dipakai sekarang.
func simScoreRef(rows []services.SimComparisonRow) float64 {
	for _, r := range rows {
		if r.Current {
			return r.Score
		}
	}
	if len(rows) > 0 {
		return rows[0].Score
	}
	return 0
}

// simVarNames merangkai nama variabel yang sedang dipakai.
func simVarNames(vars []services.SimVarDef) string {
	names := make([]string, len(vars))
	for i, v := range vars {
		names[i] = v.Name
	}
	return strings.Join(names, ", ")
}

// simSumLabel menandai apakah jumlah derajat keanggotaan membentuk partisi
// yang utuh, yaitu selalu bernilai satu.
func simSumLabel(sum float64) string {
	switch {
	case sum > 1.0001:
		return "tumpang tindih"
	case sum < 0.9999:
		return "ada celah"
	default:
		return "partisi utuh"
	}
}

func simSumClass(sum float64) string {
	if sum > 1.0001 || sum < 0.9999 {
		return "text-amber-600 dark:text-amber-400"
	}
	return "text-muted-foreground"
}

// simSortedRules mengurutkan aturan aktif dari kekuatan terbesar agar aturan
// yang paling menentukan hasil tampil lebih dulu.
func simSortedRules(rules []services.SimRule) []services.SimRule {
	out := make([]services.SimRule, len(rules))
	copy(out, rules)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Alpha > out[j].Alpha })
	return out
}

// simEffectiveCount menghitung aturan yang kekuatannya cukup berarti.
func simEffectiveCount(rules []services.SimRule) int {
	n := 0
	for _, r := range rules {
		if r.Alpha >= 0.001 {
			n++
		}
	}
	return n
}

// simAxisTicks adalah nilai sumbu mendatar pada grafik fungsi keanggotaan.
func simAxisTicks() []float64 { return []float64{0, 20, 40, 60, 80, 100} }
