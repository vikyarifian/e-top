package models

import "time"

type PageInfo struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int
	SortBy     string
	SortDir    string
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
