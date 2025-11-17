package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONB map[string]interface{}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &j)
}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Log struct {
	ID     int    `gorm:"column:id;primaryKey;type:int;" json:"id"`
	UserID string `gorm:"column:user_id;not null;type:uuid" json:"user_id"`
	User   User   `gorm:"foreignKey:UserID;references:ID"`

	Action string `gorm:"column:action;type:varchar(50);not null;check:action IN ('created_task','updated_task','created_subtask','updated_subtask','completed_task','created_project','updated_project','completed_project','created_workspace','updated_workspace','added_comment','added_member','removed_member','joined_workspace','transferred_workspace_ownership','added_attachment')" json:"action"`

	ResourceType string `gorm:"column:resource_type;type:varchar(50);not null;check:resource_type IN ('Task','Project','Workspace','Comment','User')" json:"resource_type"`
	ResourceID   string `gorm:"column:resource_id;type:uuid;not null" json:"resource_id"`

	Details JSONB `gorm:"column:details;type:jsonb" json:"details,omitempty"`

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

// -- TABLE: logs
// CREATE TABLE logs (
//     id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

//     action VARCHAR(128) NOT NULL CHECK (action IN (
//         'created_dept',
//         'updated_dept',
//         'created_task',
//         'updated_task',
//         'created_subtask',
//         'updated_subtask',
//         'completed_task',
//         'created_project',
//         'updated_project',
//         'completed_project',
//         'created_workspace',
//         'updated_workspace',
//			'watched_task',
//			'unwatched_task',
//         'added_comment',
//         'added_member',
//         'removed_member',
//		   'invited_workspace',
//         'joined_workspace',
//         'transferred_workspace_ownership',
//         'added_attachment',
//		   'registered_user',
//			'login_user',
//			'forgot_password_user',
//			'resend_email_user',
//			'verified_user'
//     )),

//     resource_type VARCHAR(128) NOT NULL CHECK (resource_type IN ('User','Dept','Workspace','Project','Task','Comment')),
//     resource_id VARCHAR(255) NOT NULL,

//     details JSONB, -- simpan info tambahan fleksibel

//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT
// );

// -- index untuk pencarian cepat by resource
// CREATE INDEX idx_logs_resource ON logs(resource_type, resource_id);

// -- index untuk pencarian by user
// CREATE INDEX idx_logs_user ON logs(user_id);

// -- index GIN untuk details JSONB biar bisa query dalam JSON
// CREATE INDEX idx_logs_details_gin ON logs USING GIN (details);

// ALTER CHECK
// ALTER TABLE logs
// DROP CONSTRAINT logs_action_check;
// ALTER TABLE logs
// ADD CONSTRAINT action CHECK (action IN (
//         'created_dept',
//         'updated_dept',
//         'created_task',
//         'updated_task',
//         'created_subtask',
//         'updated_subtask',
//         'completed_task',
//         'created_project',
//         'updated_project',
//         'completed_project',
//         'created_workspace',
//         'updated_workspace',
// 			'watched_task',
// 			'unwatched_task',
//         'added_comment',
//         'added_member',
//         'removed_member',
// 		   'invited_workspace',
//         'joined_workspace',
// 		'declined_workspace',
//         'transferred_workspace_ownership',
//         'added_attachment',
// 		   'registered_user',
// 			'login_user',
// 			'forgot_password_user',
// 			'resend_email_user',
// 			'verified_user'
//     ));
