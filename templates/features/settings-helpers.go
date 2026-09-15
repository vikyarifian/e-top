package features

import (
	"encoding/json"
	"fmt"
)

// settingsSearchScope menyusun cakupan Alpine untuk kotak pencarian pada
// halaman Settings.
//
// Sama seperti My Tasks, pengiriman memakai htmx.ajax dari penangan peristiwa
// Alpine agar tidak bergantung pada atribut hx-post yang baru muncul setelah
// htmx selesai memeriksa halaman.
func settingsSearchScope(tab string, query string) string {
	b, err := json.Marshal(map[string]string{"q": query})
	if err != nil {
		b = []byte(`{"q":""}`)
	}
	return fmt.Sprintf(`Object.assign(%s, {
		url() {
			const p = new URLSearchParams();
			p.set('tab', %q);
			p.set('page', '1');
			if (this.q && this.q.trim()) p.set('q', this.q.trim());
			return '/settings?' + p.toString();
		},
		kirim() {
			const u = this.url();
			htmx.ajax('POST', u, { target: '#content', swap: 'innerHTML' });
			try { history.replaceState({}, '', u); } catch (e) {}
		}
	})`, string(b), tab)
}
