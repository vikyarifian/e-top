package models

import (
	"time"
)

// Attachment model
type Attachment struct {
	No         int        `gorm:"column:no;primaryKey" json:"-"`
	ID         string     `gorm:"column:id;unique" json:"id"`
	TaskID     string     `gorm:"column:task_id;not null;index" json:"task_id"`
	FileName   string     `gorm:"column:file_name;not null" json:"file_name"`
	FileUrl    string     `gorm:"column:file_url;not null" json:"file_url"`
	FileType   string     `gorm:"column:file_type" json:"file_type"`
	FileSize   int64      `gorm:"column:file_size" json:"file_size"`
	UploadedAt *time.Time `gorm:"column:uploaded_at;type:TIMESTAMP;default:CURRENT_TIMESTAMP" json:"uploaded_at"`
	UploadedBy string     `gorm:"column:uploaded_by;not null" json:"uploaded_by"`
	UpdatedAt  *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy  string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}
