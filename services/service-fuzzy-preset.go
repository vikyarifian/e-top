package services

// Prasetel konfigurasi fungsi keanggotaan dan pemuatan nilai indikator dari
// basis data untuk halaman Simulasi.

import (
	"strconv"

	"etop/db"
	"etop/models"
)

// Katalog variabel input yang dapat dipakai simulasi, berurutan dari yang
// paling utama. Empat variabel pertama merupakan rancangan sistem.
var SimVarCatalog = []SimVarDef{
	{Name: "TCR", Label: "Task Completion Rate"},
	{Name: "OTR", Label: "On-Time Rate"},
	{Name: "TVS", Label: "Task Value Score"},
	{Name: "WER", Label: "Work Efficiency Rate"},
	{Name: "WLR", Label: "Workload Ratio"},
}

// SimSetLabels mengembalikan label linguistik untuk sejumlah himpunan.
func SimSetLabels(n int) []string {
	switch n {
	case 2:
		return []string{"Rendah", "Tinggi"}
	case 3:
		return []string{"Rendah", "Sedang", "Tinggi"}
	case 4:
		return []string{"Rendah", "Sedang", "Tinggi", "Sangat Tinggi"}
	case 5:
		return []string{"Sangat Rendah", "Rendah", "Sedang", "Tinggi", "Sangat Tinggi"}
	}
	out := make([]string, n)
	for i := range out {
		out[i] = "Set " + strconv.Itoa(i+1)
	}
	return out
}

// SimUniformSets membangun partisi segitiga yang tersebar merata pada semesta
// 0 sampai 100. Himpunan terendah berupa bahu kiri dan tertinggi bahu kanan,
// sehingga jumlah derajat keanggotaannya selalu satu.
func SimUniformSets(n int) []SimSet {
	if n < 2 {
		n = 2
	}
	labels := SimSetLabels(n)
	node := func(i int) float64 { return float64(i) * 100 / float64(n-1) }
	sets := make([]SimSet, n)
	for i := 0; i < n; i++ {
		switch {
		case i == 0:
			sets[i] = SimSet{labels[i], MFLinearTurun, []float64{node(0), node(1)}}
		case i == n-1:
			sets[i] = SimSet{labels[i], MFLinearNaik, []float64{node(n - 2), node(n - 1)}}
		default:
			sets[i] = SimSet{labels[i], MFSegitiga, []float64{node(i - 1), node(i), node(i + 1)}}
		}
	}
	return sets
}

// SimPreset menjelaskan satu prasetel konfigurasi.
type SimPreset struct {
	Value string
	Label string
	Desc  string
	Sets  []SimSet
}

// SimPresets adalah daftar prasetel yang tersedia pada antarmuka.
func SimPresets() []SimPreset {
	return []SimPreset{
		{
			Value: "sistem",
			Label: "System design (3 sets, knots 40-60-90)",
			Desc:  "The configuration the application actually uses on the Performance Evaluation page.",
			Sets: []SimSet{
				{"Rendah", MFLinearTurun, []float64{40, 60}},
				{"Sedang", MFSegitiga, []float64{40, 60, 90}},
				{"Tinggi", MFLinearNaik, []float64{60, 90}},
			},
		},
		{
			Value: "awal",
			Label: "Initial design (2 sets, transition 40-60)",
			Desc:  "The design before revision. It saturates: every value above 60 is treated alike.",
			Sets: []SimSet{
				{"Rendah", MFLinearTurun, []float64{40, 60}},
				{"Tinggi", MFLinearNaik, []float64{40, 60}},
			},
		},
		{
			Value: "segitiga",
			Label: "Uniform triangular partition (0-50-100)",
			Desc:  "Triangles spread evenly across the whole universe of discourse.",
			Sets:  SimUniformSets(3),
		},
		{
			Value: "trapesium",
			Label: "Trapezoidal (widened peak)",
			Desc:  "The peak is a flat plateau, so a whole range of values has full membership.",
			Sets: []SimSet{
				{"Rendah", MFTrapesium, []float64{0, 0, 40, 60}},
				{"Sedang", MFTrapesium, []float64{40, 55, 70, 90}},
				{"Tinggi", MFTrapesium, []float64{60, 80, 100, 100}},
			},
		},
		{
			Value: "gaussian",
			Label: "Gaussian (centres 40-60-90)",
			Desc:  "A bell curve. It never reaches zero, so every rule is always active.",
			Sets: []SimSet{
				{"Rendah", MFGaussKiri, []float64{40, 8.49}},
				{"Sedang", MFGaussian, []float64{60, 10.62}},
				{"Tinggi", MFGaussKanan, []float64{90, 12.74}},
			},
		},
	}
}

// SimPresetByValue mencari prasetel berdasarkan kodenya.
func SimPresetByValue(v string) (SimPreset, bool) {
	for _, p := range SimPresets() {
		if p.Value == v {
			return p, true
		}
	}
	return SimPreset{}, false
}

// CloneSets menyalin susunan himpunan agar perubahan pada salinan tidak
// memengaruhi prasetel aslinya.
func CloneSets(sets []SimSet) []SimSet {
	out := make([]SimSet, len(sets))
	for i, s := range sets {
		p := make([]float64, len(s.Params))
		copy(p, s.Params)
		out[i] = SimSet{Label: s.Label, Shape: s.Shape, Params: p}
	}
	return out
}

// SimData memuat nilai indikator seorang karyawan pada satu periode,
// dipakai sebagai nilai awal simulasi.
type SimData struct {
	UserID    string
	Year      string
	TaskCount int64
	DoneCount int64
	Values    map[string]float64
	Available bool
}

// LoadSimData menghitung indikator KPI dari basis data untuk dipakai sebagai
// nilai awal simulasi, termasuk Workload Ratio sebagai variabel kelima.
func LoadSimData(userID, year string) SimData {
	d := SimData{UserID: userID, Year: year, Values: map[string]float64{}}
	e := GetAchievedEvaluation(userID, year)
	d.TaskCount = e.TaskCount
	d.DoneCount = e.DoneCount
	d.Available = e.TaskCount > 0
	d.Values["TCR"] = e.TCR
	d.Values["OTR"] = e.OTR
	d.Values["TVS"] = e.TVS
	d.Values["WER"] = e.WER

	// Workload Ratio: jumlah tugas karyawan dibanding rata-rata tugas seluruh
	// karyawan yang aktif pada periode yang sama, dibatasi maksimum 100.
	var avg float64
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			db.PgSql.Raw(`SELECT COALESCE(AVG(c), 0) FROM (
				SELECT user_id, COUNT(*) AS c FROM tasks
				WHERE EXTRACT(YEAR FROM created_at) = ? GROUP BY user_id) s`, y).Scan(&avg)
		}
	} else {
		db.PgSql.Raw(`SELECT COALESCE(AVG(c), 0) FROM (
			SELECT user_id, COUNT(*) AS c FROM tasks GROUP BY user_id) s`).Scan(&avg)
	}
	wlr := 0.0
	if avg > 0 {
		wlr = float64(e.TaskCount) / avg * 100
		if wlr > 100 {
			wlr = 100
		}
	}
	d.Values["WLR"] = wlr
	return d
}

// SimYears mengembalikan daftar tahun yang memiliki tugas bagi seorang pengguna.
func SimYears(userID string) []int {
	var years []int
	db.PgSql.Model(&models.Task{}).
		Where("user_id = ?", userID).
		Select("DISTINCT EXTRACT(YEAR FROM created_at)::int as year").
		Order("year DESC").
		Pluck("year", &years)
	return years
}

// SimUsers mengembalikan pengguna yang memiliki tugas, untuk pemilih karyawan.
func SimUsers() []models.User {
	var users []models.User
	db.PgSql.
		Where("id IN (SELECT DISTINCT user_id FROM tasks)").
		Order("full_name").
		Find(&users)
	return users
}

// SimComparisonRow adalah satu baris tabel perbandingan pada halaman Simulasi.
type SimComparisonRow struct {
	Key         string
	Label       string
	Desc        string
	SetCount    int
	VarCount    int
	RuleTotal   int
	ActiveCount int
	Score       float64
	Category    string
	Nanos       float64
	Current     bool
}

// SimCompareShapes menghitung hasil akhir untuk nilai input yang sama pada
// seluruh prasetel bentuk fungsi keanggotaan. Ini menjawab pertanyaan
// mengenai dampak perubahan bentuk fungsi keanggotaan.
func SimCompareShapes(vars []SimVarDef, currentPreset string) []SimComparisonRow {
	rows := []SimComparisonRow{}
	for _, p := range SimPresets() {
		cfg := SimConfig{Vars: vars, Sets: CloneSets(p.Sets)}
		r := RunSimulation(cfg)
		rows = append(rows, SimComparisonRow{
			Key:         p.Value,
			Label:       p.Label,
			Desc:        p.Desc,
			SetCount:    len(p.Sets),
			VarCount:    len(vars),
			RuleTotal:   r.RuleTotal,
			ActiveCount: len(r.Active),
			Score:       r.Score,
			Category:    r.Category,
			Nanos:       r.NanosPerInfer,
			Current:     p.Value == currentPreset,
		})
	}
	return rows
}

// SimCompareVarCount menghitung hasil akhir memakai susunan himpunan yang
// sama tetapi dengan jumlah variabel input yang berbeda. Ini menjawab
// pertanyaan mengenai dampak penambahan jumlah input.
func SimCompareVarCount(sets []SimSet, values map[string]float64, current int) []SimComparisonRow {
	rows := []SimComparisonRow{}
	for n := 2; n <= len(SimVarCatalog); n++ {
		vars := make([]SimVarDef, 0, n)
		for i := 0; i < n; i++ {
			v := SimVarCatalog[i]
			v.Value = values[v.Name]
			vars = append(vars, v)
		}
		cfg := SimConfig{Vars: vars, Sets: sets}
		r := RunSimulation(cfg)
		label := ""
		for i, v := range vars {
			if i > 0 {
				label += ", "
			}
			label += v.Name
		}
		rows = append(rows, SimComparisonRow{
			Key:         strconv.Itoa(n),
			Label:       label,
			SetCount:    len(sets),
			VarCount:    n,
			RuleTotal:   r.RuleTotal,
			ActiveCount: len(r.Active),
			Score:       r.Score,
			Category:    r.Category,
			Nanos:       r.NanosPerInfer,
			Current:     n == current,
		})
	}
	return rows
}

// SimulationPage adalah model tampilan lengkap halaman Simulasi.
type SimulationPage struct {
	ViewUsers    []models.User
	SelectedUser string
	Years        []int
	SelectedYear string
	Data         SimData
	NVar         int
	NSet         int
	Preset       string
	Applied      string
	Cfg          SimConfig
	Result       SimResult
	ShapeRows    []SimComparisonRow
	VarRows      []SimComparisonRow
	RuleSample   []SimRule
}
