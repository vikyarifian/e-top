package dto

import "etop/models"

// DeptTasks memuat daftar tugas seluruh anggota sebuah departemen beserta
// keadaan pencarian dan penyaring yang sedang berlaku.
//
// Halaman departemen sebelumnya hanya menampilkan daftar anggota, sehingga
// kepala departemen dapat melihat nilai kinerja bawahannya lewat halaman
// penilaian tetapi tidak dapat melihat pekerjaan apa yang sedang mereka
// kerjakan. Daftar ini menutup jarak itu.
type DeptTasks struct {
	// DeptID dipakai menyusun tautan penomoran halaman dari dalam pecahan
	// daftar, yang tidak menerima objek departemennya.
	DeptID  string
	Tasks   []models.Task
	Page    models.PageInfo
	Members []models.User
}
