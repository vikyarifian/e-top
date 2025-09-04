package models

import (
	"time"
)

type Project struct {
	No          int             `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID          string          `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Title       string          `gorm:"column:title;not null" json:"title" form:"title"`
	Description string          `gorm:"column:description" json:"description,omitempty" form:"description"`
	WorkspaceID string          `gorm:"column:workspace_id;not null;index" json:"workspace_id" form:"workspace_id"`
	Workspace   Workspace       `gorm:"foreignKey:WorkspaceID;references:ID" json:"workspace,omitempty"`
	Status      string          `gorm:"column:status;default:'PLANNING'" json:"status" form:"status"`
	StartDate   *time.Time      `gorm:"column:start_date;type:DATE" json:"start_date,omitempty" form:"start_date"`
	DueDate     *time.Time      `gorm:"column:due_date;type:DATE" json:"due_date,omitempty" form:"due_date"`
	Progress    int             `gorm:"column:progress;default:0" json:"progress" form:"progress"`
	Tasks       []Task          `gorm:"foreignKey:ProjectID;references:ID" json:"tasks,omitempty"`
	Members     []ProjectMember `gorm:"foreignKey:ProjectID;references:ID" json:"members,omitempty"`
	Tags        []ProjectTag    `gorm:"foreignKey:ProjectID;references:ID" json:"tags,omitempty"`
	IsArchived  bool            `gorm:"column:is_archived;default:false" json:"is_archived" form:"is_archived"`
	CreatedAt   *time.Time      `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy   string          `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy   string          `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type ProjectMember struct {
	No        int        `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID        string     `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	ProjectID string     `gorm:"column:project_id;not null;index" json:"project_id" form:"project_id"`
	UserID    string     `gorm:"column:user_id;not null;index" json:"user_id" form:"user_id"`
	User      User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Role      string     `gorm:"column:role;default:'CONTRIBUTOR'" json:"role" form:"role"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type ProjectTag struct {
	No        int    `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID        string `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	ProjectID string `gorm:"column:project_id;not null;index" json:"project_id"`
	Tag       string `gorm:"column:tag;not null" json:"tag"`
}

// -- Tabel projects
// CREATE TABLE projects (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     title VARCHAR(255) NOT NULL,
//     description TEXT,
//     workspace_id VARCHAR(255) NOT NULL,
//     status VARCHAR(50) DEFAULT 'PLANNING',
//     start_date DATE,
//     due_date DATE,
//     progress INT DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
//     is_archived BOOLEAN DEFAULT FALSE,
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     CONSTRAINT fk_workspace FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
// );

// -- Tabel project_members
// CREATE TABLE project_members (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     project_id VARCHAR(255) NOT NULL,
//     user_id VARCHAR(255) NOT NULL,
//     role VARCHAR(50) DEFAULT 'CONTRIBUTOR',
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
//     CONSTRAINT fk_project FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
//     CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
// );

// -- Tabel project_tags
// CREATE TABLE project_tags (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     project_id VARCHAR(255) NOT NULL,
//     tag VARCHAR(100) NOT NULL,
//     CONSTRAINT fk_project_tag FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
// );

// -- Index tambahan
// CREATE INDEX idx_project_members_project_id ON project_members(project_id);
// CREATE INDEX idx_project_members_user_id ON project_members(user_id);
// CREATE INDEX idx_project_tags_project_id ON project_tags(project_id);
