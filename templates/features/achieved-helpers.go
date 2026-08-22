package features

import (
	"math"
	"time"

	"etop/services"
)

func formatDateOrDash(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("Mon, 02 Jan 2006")
}

type segmentData struct {
	Color        string
	Length       float64
	Offset       float64
	Circumference float64
}

func circumference(value float64) float64 {
	c := 2 * math.Pi * 15.5
	return c * clampEval(value) / 100
}

func maxDistCount(dist []services.StatusCount) int64 {
	var m int64
	for _, s := range dist {
		if s.Count > m {
			m = s.Count
		}
	}
	if m == 0 {
		m = 1
	}
	return m
}

func sumTypeCount(types []services.TypeCount) int64 {
	var total int64
	for _, t := range types {
		total += t.Count
	}
	if total == 0 {
		total = 1
	}
	return total
}

func computeSegments(types []services.TypeCount, total int64) []segmentData {
	colors := []string{"#3b82f6", "#22c55e", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#14b8a6"}
	c := 2 * math.Pi * 13
	var segments []segmentData
	offset := 0.0
	for i, t := range types {
		pct := float64(t.Count) / float64(total)
		length := c * pct
		segments = append(segments, segmentData{
			Color:         colors[i%len(colors)],
			Length:        length,
			Offset:        offset,
			Circumference: c,
		})
		offset += length
	}
	return segments
}

func maxMonthlyCount(ms []services.MonthlyCount) int64 {
	var m int64
	for _, v := range ms {
		if v.Count > m {
			m = v.Count
		}
	}
	if m == 0 {
		m = 1
	}
	return m
}

func statusColor(color string) string {
	if color == "" {
		return "bg-primary"
	}
	return color
}

func typeColor(t string) string {
	switch t {
	case "PROJECT":
		return "bg-blue-500"
	case "DAILY":
		return "bg-green-500"
	case "TICKET":
		return "bg-amber-500"
	default:
		return "bg-gray-500"
	}
}

func monthLabel(m int) string {
	names := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	if m >= 1 && m <= 12 {
		return names[m]
	}
	return ""
}

func progressBarColor(v float64) string {
	if v >= 80 {
		return "bg-green-500"
	}
	if v >= 50 {
		return "bg-amber-500"
	}
	return "bg-red-500"
}

// ---- jejak audit inferensi Fuzzy Tsukamoto ----

type fuzzyVarRow struct {
	Name   string
	Value  float64
	Rendah float64
	Sedang float64
	Tinggi float64
}

// fuzzyVarRows menyusun tabel fuzzifikasi keempat variabel input.
func fuzzyVarRows(e services.AchievedEvaluation) []fuzzyVarRow {
	vals := [4]float64{e.TCR, e.OTR, e.TPS, e.WER}
	rows := make([]fuzzyVarRow, 0, 4)
	for i, name := range services.FuzzyVarNames {
		mu := services.FuzzyMembership(vals[i])
		rows = append(rows, fuzzyVarRow{
			Name:   name,
			Value:  vals[i],
			Rendah: mu[services.SetRendah],
			Sedang: mu[services.SetSedang],
			Tinggi: mu[services.SetTinggi],
		})
	}
	return rows
}

func sumAlpha(rules []services.FuzzyRule) float64 {
	var s float64
	for _, r := range rules {
		s += r.Alpha
	}
	return s
}

func sumAlphaZ(rules []services.FuzzyRule) float64 {
	var s float64
	for _, r := range rules {
		s += r.Alpha * r.Z
	}
	return s
}

func rulePremise(r services.FuzzyRule) string {
	labels := r.PremiseLabels()
	s := ""
	for i, name := range services.FuzzyVarNames {
		if i > 0 {
			s += " · "
		}
		s += name + " " + labels[i]
	}
	return s
}

func ruleBadgeColor(output string) string {
	switch output {
	case "Sangat Baik":
		return "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
	case "Baik":
		return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
	case "Cukup":
		return "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400"
	case "Buruk":
		return "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400"
	default:
		return "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
	}
}

func clampPercentage(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
