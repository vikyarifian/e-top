package features

import (
	"encoding/json"
	"fmt"
)

// settingsSearchScope menyusun cakupan Alpine untuk kotak pencarian pada
// halaman Settings.
//
// Komponen ui.SearchBox memakai q dan url() dari cakupan di sekitarnya, jadi
// cakupan itu harus disediakan pemanggilnya. Nilai awalnya diambil dari kata
// kunci yang sedang berlaku di server supaya isian kotak tetap sesuai setelah
// halaman dimuat ulang atau setelah tombol kembali peramban ditekan.
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
		}
	})`, string(b), tab)
}
