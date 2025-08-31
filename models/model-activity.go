package models

import (
	"time"
)

type Log struct {
	ID     string `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID string `gorm:"column:user_id;not null;type:uuid" json:"user_id"`
	User   *User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`

	Action string `gorm:"column:action;type:varchar(50);not null;check:action IN ('created_task','updated_task','created_subtask','updated_subtask','completed_task','created_project','updated_project','completed_project','created_workspace','updated_workspace','added_comment','added_member','removed_member','joined_workspace','transferred_workspace_ownership','added_attachment')" json:"action"`

	ResourceType string `gorm:"column:resource_type;type:varchar(50);not null;check:resource_type IN ('Task','Project','Workspace','Comment','User')" json:"resource_type"`
	ResourceID   string `gorm:"column:resource_id;type:uuid;not null" json:"resource_id"`

	Details map[string]any `gorm:"column:details;type:jsonb" json:"details,omitempty"`

	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type Comment struct {
	ID string `gorm:"column:id;primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`

	Text string `gorm:"column:text;type:text;not null" json:"text"`

	TaskID string `gorm:"column:task_id;type:uuid;not null" json:"task_id"`
	Task   *Task  `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`

	AuthorID string `gorm:"column:author_id;type:uuid;not null" json:"author_id"`
	Author   *User  `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE" json:"author,omitempty"`

	// Mentions → simpan array objek (user_id, offset, length) ke JSONB
	Mentions    []Mention `gorm:"-" json:"mentions,omitempty"` // pakai struct slice, disimpan manual ke JSONB
	MentionsRaw []byte    `gorm:"column:mentions;type:jsonb" json:"-"`

	// Reactions → array JSONB juga
	Reactions    []Reaction `gorm:"-" json:"reactions,omitempty"`
	ReactionsRaw []byte     `gorm:"column:reactions;type:jsonb" json:"-"`

	// Attachments → array JSONB
	Attachments    []Attachment `gorm:"foreignKey:TaskID;references:ID" json:"attachments"`
	AttachmentsRaw []byte       `gorm:"column:attachments;type:jsonb" json:"-"`

	IsEdited bool `gorm:"column:is_edited;default:false" json:"is_edited"`

	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

// Mentions
type Mention struct {
	UserID string `json:"user_id"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// Reactions
type Reaction struct {
	Emoji  string `json:"emoji"`
	UserID string `json:"user_id"`
}
