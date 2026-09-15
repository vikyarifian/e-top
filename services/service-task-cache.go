package services

import (
	"sync"

	"etop/db"
	"etop/models"
)

// Tabel acuan tugas berisi paling banyak beberapa belas baris dan hanya berubah
// ketika pengelola menyimpan halaman Task Config. Sebelum ini setiap render
// halaman menembak ulang ketiganya, dan pada basis data jarak jauh satu kueri
// sekecil apa pun tetap memakan satu perjalanan bolak-balik sekitar 20 ms.
// Halaman My Tasks sendiri membacanya sampai enam kali, jadi singgahan ini
// memangkas sekitar seratus milidetik pada tiap permintaan.
//
// Singgahan dibatalkan oleh BersihkanSinggahanAcuan, yang dipanggil setelah
// penyimpanan Task Config berhasil. Tanpa itu, perubahan bobot tidak akan
// terlihat sampai aplikasi dijalankan ulang.
var (
	kunciAcuan      sync.RWMutex
	cacheStatuses   []models.TaskStatus
	cachePriorities []models.TaskPriority
	cacheImpacts    []models.TaskImpact
)

// BersihkanSinggahanAcuan mengosongkan singgahan tabel acuan tugas.
func BersihkanSinggahanAcuan() {
	kunciAcuan.Lock()
	cacheStatuses, cachePriorities, cacheImpacts = nil, nil, nil
	kunciAcuan.Unlock()
}

func salinStatus(v []models.TaskStatus) []models.TaskStatus {
	out := make([]models.TaskStatus, len(v))
	copy(out, v)
	return out
}

func salinPrioritas(v []models.TaskPriority) []models.TaskPriority {
	out := make([]models.TaskPriority, len(v))
	copy(out, v)
	return out
}

func salinDampak(v []models.TaskImpact) []models.TaskImpact {
	out := make([]models.TaskImpact, len(v))
	copy(out, v)
	return out
}

// GetTaskStatuses mengembalikan acuan status tugas.
func GetTaskStatuses() []models.TaskStatus {
	kunciAcuan.RLock()
	if cacheStatuses != nil {
		defer kunciAcuan.RUnlock()
		return salinStatus(cacheStatuses)
	}
	kunciAcuan.RUnlock()

	hasil := []models.TaskStatus{}
	if err := db.PgSql.Find(&hasil).Error; err != nil {
		return []models.TaskStatus{}
	}
	kunciAcuan.Lock()
	cacheStatuses = hasil
	kunciAcuan.Unlock()
	return salinStatus(hasil)
}

// GetTaskPriorities mengembalikan acuan prioritas tugas, terurut menurut kolom
// no agar pilihan pada formulir selalu tampil dalam urutan yang sama.
func GetTaskPriorities() []models.TaskPriority {
	kunciAcuan.RLock()
	if cachePriorities != nil {
		defer kunciAcuan.RUnlock()
		return salinPrioritas(cachePriorities)
	}
	kunciAcuan.RUnlock()

	hasil := []models.TaskPriority{}
	if err := db.PgSql.Order("no").Find(&hasil).Error; err != nil {
		return []models.TaskPriority{}
	}
	kunciAcuan.Lock()
	cachePriorities = hasil
	kunciAcuan.Unlock()
	return salinPrioritas(hasil)
}

// GetTaskImpacts mengembalikan acuan luas dampak tugas.
func GetTaskImpacts() []models.TaskImpact {
	kunciAcuan.RLock()
	if cacheImpacts != nil {
		defer kunciAcuan.RUnlock()
		return salinDampak(cacheImpacts)
	}
	kunciAcuan.RUnlock()

	hasil := []models.TaskImpact{}
	if err := db.PgSql.Order("no").Find(&hasil).Error; err != nil {
		return []models.TaskImpact{}
	}
	kunciAcuan.Lock()
	cacheImpacts = hasil
	kunciAcuan.Unlock()
	return salinDampak(hasil)
}
