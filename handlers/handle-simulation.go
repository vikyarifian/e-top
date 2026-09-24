package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"etop/auth"
	"etop/services"
	"etop/templates/features"
	"etop/templates/layouts"
)

const (
	simMinVar = 2
	simMinSet = 2
)

// simFloat membaca satu nilai desimal dari formulir, dengan nilai bawaan bila
// kosong atau tidak sah.
func simFloat(r *http.Request, key string, def float64) float64 {
	v := r.FormValue(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}

// simInt membaca satu bilangan bulat dari formulir dengan pembatasan rentang.
func simInt(r *http.Request, key string, def, min, max int) int {
	v := r.FormValue(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	if n < min {
		n = min
	}
	if n > max {
		n = max
	}
	return n
}

// simSetsFromForm menyusun kembali susunan himpunan dari isian formulir.
func simSetsFromForm(r *http.Request, nset int) []services.SimSet {
	labels := services.SimSetLabels(nset)
	sets := make([]services.SimSet, nset)
	for j := 0; j < nset; j++ {
		shape := r.FormValue(fmt.Sprintf("shape_%d", j))
		if services.ShapeLabel(shape) == shape {
			// bentuk tidak dikenal, pakai bawaan segitiga
			shape = services.MFSegitiga
		}
		names := services.ShapeParams(shape)
		params := make([]float64, len(names))
		for k := range names {
			params[k] = simFloat(r, fmt.Sprintf("p_%d_%d", j, k), 0)
		}
		sets[j] = services.SimSet{Label: labels[j], Shape: shape, Params: params}
	}
	return sets
}

// simSetsEqual membandingkan dua susunan himpunan untuk mengetahui apakah
// pengguna telah mengubah parameter sebuah prasetel.
func simSetsEqual(a, b []services.SimSet) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Shape != b[i].Shape || len(a[i].Params) != len(b[i].Params) {
			return false
		}
		for k := range a[i].Params {
			d := a[i].Params[k] - b[i].Params[k]
			if d > 1e-9 || d < -1e-9 {
				return false
			}
		}
	}
	return true
}

// HandleSimulation menyajikan halaman Simulasi Fuzzy Tsukamoto. Permintaan GET
// menghasilkan halaman penuh, sedangkan POST hanya mengembalikan bagian isi
// yang perlu diperbarui sehingga kendali formulir tidak kehilangan fokus.
func HandleSimulation(w http.ResponseWriter, r *http.Request) error {
	user, _ := auth.GetAuth(w, r)
	if err := r.ParseForm(); err != nil {
		return err
	}

	viewUsers := achievedViewUsers(user)

	// Pemilihan karyawan yang datanya dipakai sebagai nilai awal. Aturannya
	// sama dengan halaman penilaian: nama teratas menurut abjad, bukan pengguna
	// yang sedang masuk.
	targetID, selectedUser := pilihTargetPenilaian(user, viewUsers, r.FormValue("user_id"))
	years := services.SimYears(targetID)
	selectedYear := r.FormValue("year")
	if selectedYear != "" {
		if _, err := strconv.Atoi(selectedYear); err != nil {
			selectedYear = ""
		}
	}

	data := services.LoadSimData(targetID, selectedYear)

	nvar := simInt(r, "nvar", 4, simMinVar, len(services.SimVarCatalog))
	nset := simInt(r, "nset", 3, simMinSet, 5)
	preset := r.FormValue("preset")
	if preset == "" {
		preset = "sistem"
	}

	// Token penanda konfigurasi yang terakhir dikirim ke peramban. Bila token
	// berbeda dengan keadaan sekarang, berarti pengguna baru saja mengganti
	// prasetel atau jumlah himpunan, sehingga susunan himpunan perlu dibangun
	// ulang alih-alih dibaca dari isian formulir yang sudah usang.
	prevPreset, prevNset := "", 0
	if _, err := fmt.Sscanf(r.FormValue("applied"), "%s %d", &prevPreset, &prevNset); err != nil {
		prevPreset, prevNset = "", 0
	}

	var sets []services.SimSet
	switch {
	case r.Method != http.MethodPost || prevPreset == "":
		// muatan awal halaman
		if p, ok := services.SimPresetByValue(preset); ok {
			sets = services.CloneSets(p.Sets)
			nset = len(sets)
		} else {
			sets = services.SimUniformSets(nset)
		}
	case preset != prevPreset && preset != "custom":
		// pengguna memilih prasetel lain
		if p, ok := services.SimPresetByValue(preset); ok {
			sets = services.CloneSets(p.Sets)
			nset = len(sets)
		} else {
			sets = services.SimUniformSets(nset)
		}
	case nset != prevNset:
		// pengguna mengubah jumlah himpunan
		sets = services.SimUniformSets(nset)
		preset = "custom"
	default:
		// pengguna menyunting bentuk atau parameter
		sets = simSetsFromForm(r, nset)
		if p, ok := services.SimPresetByValue(preset); ok && !simSetsEqual(sets, p.Sets) {
			preset = "custom"
		}
	}
	nset = len(sets)

	// nilai input: diambil dari basis data pada muatan awal atau ketika
	// pengguna menekan tombol muat ulang, selain itu dari isian formulir
	useDB := r.Method != http.MethodPost || r.FormValue("action") == "reload"
	values := map[string]float64{}
	for _, v := range services.SimVarCatalog {
		if useDB {
			values[v.Name] = data.Values[v.Name]
		} else {
			values[v.Name] = simFloat(r, "x_"+v.Name, data.Values[v.Name])
		}
		if values[v.Name] < 0 {
			values[v.Name] = 0
		}
		if values[v.Name] > 100 {
			values[v.Name] = 100
		}
	}

	vars := make([]services.SimVarDef, 0, nvar)
	for i := 0; i < nvar; i++ {
		v := services.SimVarCatalog[i]
		v.Value = values[v.Name]
		vars = append(vars, v)
	}

	cfg := services.SimConfig{Vars: vars, Sets: sets}
	result := services.RunSimulation(cfg)

	rules := services.SimRuleBase(cfg)
	sample := rules
	if len(sample) > 12 {
		sample = sample[:12]
	}

	page := services.SimulationPage{
		ViewUsers:    viewUsers,
		SelectedUser: selectedUser,
		Years:        years,
		SelectedYear: selectedYear,
		Data:         data,
		NVar:         nvar,
		NSet:         nset,
		Preset:       preset,
		Applied:      fmt.Sprintf("%s %d", preset, nset),
		Cfg:          cfg,
		Result:       result,
		ShapeRows:    services.SimCompareShapes(vars, preset),
		VarRows:      services.SimCompareVarCount(sets, values, nvar),
		RuleSample:   sample,
	}

	if r.Method == http.MethodPost {
		return features.SimulationBody(page, user).Render(r.Context(), w)
	}
	return layouts.Layout("Fuzzy Simulation", user,
		features.Simulation(page, user)).Render(r.Context(), w)
}
