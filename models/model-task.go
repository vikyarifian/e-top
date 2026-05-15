package models

import (
	"time"
)

type Task struct {
	No          int     `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID          string  `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Type        string  `gorm:"column:type;default:'DAILY';check:type IN ('PROJECT','DAILY','TICKET')" json:"type,omitempty" form:"type"`
	Title       string  `gorm:"column:title;not null" json:"title" form:"title"`
	Description string  `gorm:"column:description" json:"description,omitempty" form:"description"`
	ProjectID   string  `gorm:"column:project_id;not null;index" json:"project_id" form:"project_id"`
	Project     Project `gorm:"foreignKey:ProjectID;references:ID"`

	Status   string `gorm:"column:status;default:'TO_DO'" json:"status"`
	Priority string `gorm:"column:priority;default:'MEDIUM'" json:"priority"`

	// Many-to-many: assignees
	// Assignees []TaskAssignee `gorm:"foreignKey:TaskID;references:ID" json:"assignees,omitempty"`

	// Assignee User `gorm:"foreignKey:TaskID;references:ID" json:"assignees,omitempty"`
	UserID   string `gorm:"column:user_id;not null;type:uuid" json:"user_id"`
	Assignee *User  `gorm:"foreignKey:UserID;references:ID" json:"assignee,omitempty"`
	// Many-to-many: watchers
	Watchers []TaskWatchers `gorm:"foreignKey:TaskID;references:ID;" json:"watchers,omitempty"`

	StartDate      *time.Time `gorm:"column:start_date;type:DATE" json:"start_date,omitempty" form:"start_date"`
	DueDate        *time.Time `gorm:"column:due_date;type:TIMESTAMP" json:"due_date,omitempty"`
	CompletedAt    *time.Time `gorm:"column:completed_at;type:TIMESTAMP" json:"completed_at,omitempty"`
	EstimatedHours float32    `gorm:"column:estimated_hours;default:0" json:"estimated_hours"`
	ActualHours    float32    `gorm:"column:actual_hours;default:0" json:"actual_hours"`

	// Array of tags → simpan sebagai TEXT[] di Postgres
	// Tags []string `gorm:"type:text[]" json:"tags"`
	Tags []TaskTag `gorm:"foreignKey:TaskID;references:ID" json:"tags,omitempty"`
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

// type TaskAssignee struct {
// 	TaskID      string     `gorm:"column:task_id;not null;index" json:"task_id"`
// 	Task        Task       `gorm:"foreignKey:TaskID;references:ID"`
// 	UserID      string     `gorm:"column:user_id;not null;index" json:"user_id"`
// 	User        User       `gorm:"foreignKey:UserID;references:ID"`
// 	Conclusion  string     `gorm:"column:conclusion" json:"conclusion,omitempty" form:"conclusion"`
// 	Status      string     `gorm:"column:status;default:'TO_DO'" json:"status"`
// 	CompletedAt *time.Time `gorm:"column:completed_at;type:TIMESTAMP" json:"completed_at,omitempty"`
// 	ActualHours float32    `gorm:"column:actual_hours;default:0" json:"actual_hours"`
// 	CreatedAt   *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
// 	CreatedBy   string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
// 	UpdatedAt   *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
// 	UpdatedBy   string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
// }

type TaskWatchers struct {
	TaskID    string     `gorm:"column:task_id;not null;index" json:"task_id"`
	Task      Task       `gorm:"foreignKey:TaskID;references:ID"`
	UserID    string     `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;references:ID"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type TaskStatus struct {
	No     int    `gorm:"column:no;primaryKey" json:"-" form:"-"`
	Status string `gorm:"column:status;not null;" json:"status" form:"status"`
	Label  string `gorm:"column:label;not null;" json:"label" form:"label"`
	Color  string `gorm:"column:color;not null;" json:"color" form:"color"`
	Form   int    `gorm:"form" json:"form,omitempty" form:"form"`
	Value  int    `gorm:"value" json:"value,omitempty" form:"value"`
}

type TaskPriority struct {
	No       int    `gorm:"column:no;primaryKey" json:"-" form:"-"`
	Priority string `gorm:"column:priority;not null;" json:"priority" form:"priority"`
	Label    string `gorm:"column:label;not null;" json:"label" form:"label"`
	Color    string `gorm:"column:color;not null;" json:"color" form:"color"`
	Value    int    `gorm:"value" json:"value,omitempty" form:"value"`
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

type TaskTag struct {
	No     int    `gorm:"column:no;primaryKey" json:"-" form:"-"`
	TaskID string `gorm:"column:task_id;not null;index" json:"task_id"`
	Tag    string `gorm:"column:tag;not null" json:"tag"`
}

// -- TABLE: tasks
// CREATE TABLE tasks (
// 	   no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     type VARCHAR(50) UNIQUE NOT NULL DEFAULT 'DAILY' CHECK (type IN ('PROJECT','DAILY','TICKET')),
//     title TEXT NOT NULL,
//     description TEXT,
//     project_id VARCHAR(255),
//     status VARCHAR(50) DEFAULT 'TO_DO' CHECK (status IN ('TO_DO','IN_PROGRESS','IN_REVIEW','DONE','CANCELLED')),
//     priority VARCHAR(50) DEFAULT 'MEDIUM' CHECK (priority IN ('LOW','MEDIUM','HIGH')),
//     user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
// 	   start_date TIMESTAMP,
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

// -- Tabel task_statuses
// CREATE TABLE task_statuses (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     status VARCHAR(100) NOT NULL,
//     label VARCHAR(100) NOT NULL,
//     color VARCHAR(225) NOT NULL,
//     form INT NOT NULL DEFAULT 1,
//     value INT NOT NULL DEFAULT 0
// );

// -- Tabel task_priorities
// CREATE TABLE task_priorities (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     priority VARCHAR(100) NOT NULL,
//     label VARCHAR(100) NOT NULL,
//     color VARCHAR(225) NOT NULL,
//     value INT NOT NULL DEFAULT 0
// );

// -- Join table: task_assignees (many-to-many)
// CREATE TABLE task_assignees (
//     task_id VARCHAR(255) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
//     user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     conclusion TEXT,
//     status VARCHAR(50) DEFAULT 'TO_DO' CHECK (status IN ('TO_DO','IN_PROGRESS','IN_REVIEW','DONE','CANCELLED')),
//     completed_at TIMESTAMP,
//     actual_hours INT DEFAULT 0 CHECK (actual_hours >= 0),
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     PRIMARY KEY (task_id, user_id)
// );

// -- Join table: task_watchers (many-to-many)
// CREATE TABLE task_watchers (
//     task_id VARCHAR(255) NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
//     user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     PRIMARY KEY (task_id, user_id)
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
// -- Tabel task_tags
// CREATE TABLE task_tags (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     task_id VARCHAR(255) NOT NULL,
//     tag VARCHAR(100) NOT NULL,
//     CONSTRAINT fk_task_tag FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
// );
