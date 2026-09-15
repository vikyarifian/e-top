package features

import (
	"encoding/json"
	"fmt"

	"etop/models"
	"etop/services"
	"etop/templates/components/ui"
)

// myTasksScope menyusun cakupan Alpine untuk halaman My Tasks.
//
// Kata kunci dan ketiga penyaring disimpan dalam satu cakupan bersama, lalu
// fungsi url() menyusun alamat yang memuat semuanya sekaligus. Tanpa itu,
// mengubah satu penyaring akan menghapus penyaring lain yang sedang aktif,
// sebab tiap kendali hanya mengetahui nilainya sendiri.
//
// Nilai awalnya diambil dari keadaan yang sedang berlaku di server supaya
// tampilan kendali tetap sesuai setelah halaman dimuat ulang atau setelah
// pengguna menekan tombol kembali pada peramban.
func myTasksScope(page models.PageInfo) string {
	awal := map[string]string{
		"q":        page.Query,
		"status":   page.Filters["status"],
		"priority": page.Filters["priority"],
		"type":     page.Filters["type"],
	}
	b, err := json.Marshal(awal)
	if err != nil {
		b = []byte(`{"q":"","status":"","priority":"","type":""}`)
	}
	return fmt.Sprintf(`Object.assign(%s, {
		url() {
			const p = new URLSearchParams();
			p.set('page', '1');
			if (this.q && this.q.trim()) p.set('q', this.q.trim());
			if (this.status) p.set('status', this.status);
			if (this.priority) p.set('priority', this.priority);
			if (this.type) p.set('type', this.type);
			p.set('part', 'list');
			return '/my-tasks?' + p.toString();
		},
		menyaring() {
			return !!((this.q && this.q.trim()) || this.status || this.priority || this.type);
		}
	})`, string(b))
}

// statusOptions menyusun pilihan penyaring status dari tabel acuan, sehingga
// daftarnya ikut berubah bila acuannya disunting lewat Task Config.
func statusOptions() []ui.Opsi {
	opsi := []ui.Opsi{{Nilai: "", Label: "Semua"}}
	for _, s := range services.GetTaskStatuses() {
		opsi = append(opsi, ui.Opsi{Nilai: fmt.Sprintf("%d", s.No), Label: s.Label})
	}
	return opsi
}

// priorityOptions menyusun pilihan penyaring prioritas dari tabel acuan.
func priorityOptions() []ui.Opsi {
	opsi := []ui.Opsi{{Nilai: "", Label: "Semua"}}
	for _, p := range services.GetTaskPriorities() {
		opsi = append(opsi, ui.Opsi{Nilai: fmt.Sprintf("%d", p.No), Label: p.Label})
	}
	return opsi
}

// typeOptions menyusun pilihan penyaring jenis tugas. Daftarnya ditulis tetap
// karena kolom type pada tabel tasks tidak memiliki tabel acuan tersendiri.
func typeOptions() []ui.Opsi {
	return []ui.Opsi{
		{Nilai: "", Label: "Semua"},
		{Nilai: "DAILY", Label: "Daily"},
		{Nilai: "PROJECT", Label: "Project"},
		{Nilai: "TICKET", Label: "Ticket"},
	}
}
