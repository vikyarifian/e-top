package handlers

import (
	"net/http"
	"strings"

	"etop/auth"
	"etop/db"
	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
	"etop/templates/features"
	"etop/utils"
)

// Menyetel ulang sandi orang lain adalah wewenang yang berat, jadi tingkat
// pengguna diperiksa di peladen, bukan sekadar disembunyikan tombolnya.
func bolehSetelSandi(w http.ResponseWriter, r *http.Request) bool {
	pengguna, _ := auth.GetAuth(w, r)
	return pengguna.Level == "ADMIN"
}

// HandleResetPasswordForm menyajikan formulir setel ulang sandi milik seorang
// pengguna. Hanya administrator yang boleh membukanya.
func HandleResetPasswordForm(w http.ResponseWriter, r *http.Request) error {
	if !bolehSetelSandi(w, r) {
		w.WriteHeader(http.StatusUnauthorized)
		return ui.Toast("user-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
	}
	var target models.User
	if err := db.PgSql.Where("id=?", r.URL.Query().Get("user_id")).First(&target).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return ui.Toast("user-error", "warning", "", "User not found!", "", nil).Render(r.Context(), w)
	}
	return features.ResetPasswordForm(target).Render(r.Context(), w)
}

// HandleResetPassword menyetel ulang sandi seorang pengguna tanpa menanyakan
// sandi lamanya. Berbeda dari HandleChangePassword yang dipakai seseorang untuk
// sandinya sendiri, di sini sandi lama memang tidak diketahui administrator.
func HandleResetPassword(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}
	admin, _ := auth.GetAuth(w, r)
	if admin.Level != "ADMIN" {
		w.WriteHeader(http.StatusUnauthorized)
		return ui.Toast("user-error", "warning", "", "You're not authorized!", "", nil).Render(r.Context(), w)
	}

	var target models.User
	if err := db.PgSql.Where("id=?", r.FormValue("user_id")).First(&target).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("user-error", "warning", "", "User is not found!", "", nil).Render(r.Context(), w)
	}

	baru := r.FormValue("new_password")
	ulang := r.FormValue("confirm_password")
	if len(strings.TrimSpace(baru)) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("user-error", "warning", "", "Password must be at least 6 characters", "", nil).Render(r.Context(), w)
	}
	if baru != ulang {
		w.WriteHeader(http.StatusBadRequest)
		return ui.Toast("user-error", "warning", "", "New & Confirm password is not match!", "", nil).Render(r.Context(), w)
	}

	sandi, err := utils.HashPassword(baru)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ui.Toast("user-error", "danger", "", "Failed to create new password!", "", nil).Render(r.Context(), w)
	}
	if err := db.PgSql.Model(&models.User{}).Where("id=?", target.ID).
		Updates(map[string]any{"password": sandi, "updated_by": admin.ID}).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return ui.Toast("user-error", "danger", "", "Failed to update password!", "", nil).Render(r.Context(), w)
	}

	// Dicatat atas nama administratornya, bukan atas nama pemilik akun, supaya
	// jejaknya menunjukkan siapa yang benar-benar melakukan.
	services.AddLog(admin.ID, "updated_user", "User", target.ID, map[string]any{
		"description": strings.TrimSpace(admin.FullName) + " reset the password of " + strings.TrimSpace(target.FullName),
	})
	return ui.Toast("user-success", "success", "", "Password reset successfully!", "", nil).Render(r.Context(), w)
}
