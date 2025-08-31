package models

import (
	"time"
)

type Task struct {
	No          int     `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID          string  `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Title       string  `gorm:"column:title;not null" json:"title" form:"title"`
	Description string  `gorm:"column:description" json:"description,omitempty" form:"description"`
	ProjectID   string  `gorm:"column:project_id;not null;index" json:"project_id" form:"project_id"`
	Project     Project `gorm:"foreignKey:ProjectID;references:ID"`

	Status   string `gorm:"column:status;default:'TO_DO'" json:"status"`
	Priority string `gorm:"column:priority;default:'MEDIUM'" json:"priority"`

	// Many-to-many: assignees
	Assignees []User `gorm:"many2many:task_assignees;" json:"assignees,omitempty"`
	// Many-to-many: watchers
	Watchers []User `gorm:"many2many:task_watchers;" json:"watchers,omitempty"`

	StartDate      *time.Time `gorm:"column:start_date;type:DATE" json:"start_date,omitempty" form:"start_date"`
	DueDate        *time.Time `gorm:"column:due_date;type:TIMESTAMP" json:"due_date,omitempty"`
	CompletedAt    *time.Time `gorm:"column:completed_at;type:TIMESTAMP" json:"completed_at,omitempty"`
	EstimatedHours int        `gorm:"column:estimated_hours;default:0" json:"estimated_hours"`
	ActualHours    int        `gorm:"column:actual_hours;default:0" json:"actual_hours"`

	// Array of tags → simpan sebagai TEXT[] di Postgres
	Tags []string `gorm:"type:text[]" json:"tags"`

	// One-to-many subtasks
	Subtasks []Subtask `gorm:"foreignKey:TaskID;references:ID" json:"subtasks"`

	// One-to-many comments
	// Comments      []Comment    `gorm:"foreignKey:TaskID;references:ID" json:"comments"`

	// One-to-many attachments
	Attachments []Attachment `gorm:"foreignKey:TaskID;references:ID" json:"attachments"`

	IsArchived bool `gorm:"column:is_archived;default:false" json:"is_archived"`

	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

// Subtask model
type Subtask struct {
	No        int    `gorm:"column:no;primaryKey" json:"-"`
	ID        string `gorm:"column:id;unique" json:"id"`
	TaskID    string `gorm:"column:task_id;not null;index" json:"task_id"`
	Title     string `gorm:"column:title;not null" json:"title"`
	Completed bool   `gorm:"column:completed;default:false" json:"completed"`

	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

// -- TABLE: tasks
// CREATE TABLE tasks (
// 	   no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     title TEXT NOT NULL,
//     description TEXT,
//     project_id VARCHAR(255) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
//     status VARCHAR(50) DEFAULT 'TO_DO' CHECK (status IN ('TO_DO','IN_PROGRESS','IN_REVIEW','DONE')),
//     priority VARCHAR(50) DEFAULT 'MEDIUM' CHECK (priority IN ('LOW','MEDIUM','HIGH')),
// 	   start_date DATE,
//     due_date TIMESTAMP,
//     completed_at TIMESTAMP,
//     estimated_hours INT DEFAULT 0 CHECK (estimated_hours >= 0),
//     actual_hours INT DEFAULT 0 CHECK (actual_hours >= 0),
//     tags TEXT[],
//     is_archived BOOLEAN DEFAULT FALSE,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT
// );

// -- TABLE: subtasks
// CREATE TABLE subtasks (
// 	   no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     task_id VARCHAR(255) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
//     title TEXT NOT NULL,
//     completed BOOLEAN DEFAULT FALSE,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT
// );

// -- TABLE: attachments
// CREATE TABLE attachments (
// 	   no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     task_id VARCHAR(255) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
//     file_name TEXT NOT NULL,
//     file_url TEXT NOT NULL,
//     file_type TEXT,
//     file_size BIGINT,
//     uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     uploaded_by VARCHAR(100) REFERENCES users(id) ON DELETE SET NULL,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT
// );
