package models

import (
	"net/url"
	"sort"
	"strings"
	"time"
)

type PageInfo struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int
	SortBy     string
	SortDir    string
	// Query adalah kata kunci pencarian yang sedang berlaku, dan Filters
	// adalah penyaring tambahan seperti status atau prioritas. Keduanya perlu
	// ikut terbawa pada tautan halaman berikutnya agar konteks tidak hilang.
	Query   string
	Filters map[string]string
}

// Params menyusun potongan query string untuk kata kunci dan penyaring yang
// sedang aktif. Kunci diurutkan agar tautan yang dihasilkan selalu sama untuk
// keadaan yang sama.
func (p PageInfo) Params() string {
	var b strings.Builder
	if p.Query != "" {
		b.WriteString("&q=" + url.QueryEscape(p.Query))
	}
	kunci := make([]string, 0, len(p.Filters))
	for k := range p.Filters {
		kunci = append(kunci, k)
	}
	sort.Strings(kunci)
	for _, k := range kunci {
		if v := p.Filters[k]; v != "" {
			b.WriteString("&" + k + "=" + url.QueryEscape(v))
		}
	}
	return b.String()
}

// Menyaring menyatakan apakah pencarian atau penyaring sedang aktif, dipakai
// untuk membedakan daftar yang memang kosong dari hasil pencarian yang nihil.
func (p PageInfo) Menyaring() bool {
	if p.Query != "" {
		return true
	}
	for _, v := range p.Filters {
		if v != "" {
			return true
		}
	}
	return false
}

type Notif struct {
	ID           int
	Action       string
	ResourceType string
	ResourceID   string
	Description  string
	Timestamp    time.Time
	ActorName    string
	ActorColor   string
}
