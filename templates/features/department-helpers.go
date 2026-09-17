package features

import (
	"encoding/json"
	"fmt"

	"etop/dto"
	"etop/templates/components/ui"
)

// deptTasksScope menyusun cakupan Alpine untuk pencarian dan penyaring pada
// daftar tugas anggota departemen.
//
// Pola yang dipakai sama dengan halaman My Tasks: seluruh kendali berbagi satu
// cakupan sehingga mengubah satu penyaring tidak menghapus penyaring lain, dan
// pengiriman memakai htmx.ajax dari penangan peristiwa. htmx memeriksa elemen
// sekali saat halaman dimuat, sehingga atribut hasil pengikatan Alpine yang
// muncul sesudahnya tidak akan pernah terdaftar.
func deptTasksScope(deptID string, t dto.DeptTasks) string {
	awal := map[string]string{
		"q":        t.Page.Query,
		"status":   t.Page.Filters["status"],
		"priority": t.Page.Filters["priority"],
		"member":   t.Page.Filters["member"],
	}
	b, err := json.Marshal(awal)
	if err != nil {
		b = []byte(`{"q":"","status":"","priority":"","member":""}`)
	}
	return fmt.Sprintf(`Object.assign(%s, {
		url() {
			const p = new URLSearchParams();
			p.set('id', %q);
			p.set('page', '1');
			p.set('part', 'tasks');
			if (this.q && this.q.trim()) p.set('q', this.q.trim());
			if (this.status) p.set('status', this.status);
			if (this.priority) p.set('priority', this.priority);
			if (this.member) p.set('member', this.member);
			return '/department?' + p.toString();
		},
		kirim() {
			htmx.ajax('POST', this.url(), { target: '#dept-task-list', swap: 'innerHTML' });
		},
		bersihkan() {
			this.q = ''; this.status = ''; this.priority = ''; this.member = '';
			this.kirim();
		},
		menyaring() {
			return !!((this.q && this.q.trim()) || this.status || this.priority || this.member);
		}
	})`, string(b), deptID)
}

// memberOptions menyusun pilihan penyaring anggota departemen.
func memberOptions(t dto.DeptTasks) []ui.Opsi {
	opsi := []ui.Opsi{{Nilai: "", Label: "Semua anggota"}}
	for _, m := range t.Members {
		nama := m.FullName
		if nama == "" {
			nama = m.Username
		}
		opsi = append(opsi, ui.Opsi{Nilai: m.ID, Label: nama})
	}
	return opsi
}

// deptTaskParams menyusun potongan query untuk tautan penomoran halaman pada
// daftar tugas departemen, lengkap dengan pengenal departemen dan penanda
// pecahan agar jawabannya tetap berupa daftar saja.
func deptTaskParams(deptID string, t dto.DeptTasks) string {
	return "&id=" + deptID + "&part=tasks" + t.Page.Params()
}
