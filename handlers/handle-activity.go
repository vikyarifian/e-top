package handlers

import (
	"etop/db"
	"etop/models"
	"etop/templates/features"
	"net/http"
)

func HandleTaskActivities(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		taskActivities := []models.Log{}
		taskID := r.URL.Query().Get("task_id")
		// Baris "selesai" dan "ditutup" berbagi cap waktu yang sama, sebab iTop
		// hanya mencatat satu resolution_date. Tanpa pemecah seri, keduanya
		// tampil dengan urutan sembarang dan riwayatnya terbaca terbalik.
		// Nomor baris ditulis menurut urutan kejadian, jadi dipakai sebagai
		// pemecah seri. Baris yang dibuat aplikasi belum bernomor dan
		// ditempatkan paling belakang di dalam cap waktu yang sama.
		if err := db.PgSql.Where("resource_type='Task' AND resource_id=?", taskID).
			Preload("User").
			Order("created_at DESC, id DESC NULLS LAST").
			Find(&taskActivities).Error; err != nil {
			return features.TaskActivities([]models.Log{}).Render(r.Context(), w)
		}
		return features.TaskActivities(taskActivities).Render(r.Context(), w)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
}
