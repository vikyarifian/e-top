package services

import (
	"etop/db"
	"etop/models"
)

// IsiJumlahTugas mengisi kolom semu TaskCount pada sekumpulan project.
//
// Halaman rincian workspace hanya perlu tahu berapa tugas yang dimiliki tiap
// project, bukan isinya. Sebelum ini jumlah itu diperoleh dengan memuat seluruh
// baris tugas lewat Preload, dan pada basis data berisi puluhan ribu tugas
// biayanya lebih dari empat detik untuk satu halaman.
//
// Satu kueri agregat dipakai untuk seluruh project sekaligus, bukan satu kueri
// per project, agar jumlah perjalanan ke basis data tetap satu berapa pun
// banyaknya project.
func IsiJumlahTugas(projects []models.Project) {
	if len(projects) == 0 {
		return
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		if p.ID != "" {
			ids = append(ids, p.ID)
		}
	}
	if len(ids) == 0 {
		return
	}

	var baris []struct {
		ProjectID string
		Jumlah    int
	}
	if err := db.PgSql.Raw(`SELECT project_id, COUNT(*) AS jumlah
		FROM tasks WHERE project_id IN (?) GROUP BY project_id`, ids).Scan(&baris).Error; err != nil {
		return
	}

	jumlah := make(map[string]int, len(baris))
	for _, b := range baris {
		jumlah[b.ProjectID] = b.Jumlah
	}
	for i := range projects {
		projects[i].TaskCount = jumlah[projects[i].ID]
	}
}
