package models

import (
	"time"
)

type Workspace struct {
	No          int               `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID          string            `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Name        string            `gorm:"column:name;" json:"name" form:"name"`
	Description string            `gorm:"column:description;" json:"description" form:"description"`
	Color       string            `gorm:"column:color;" json:"color" form:"color"`
	Members     []WorkspaceMember `gorm:"foreignKey:WorkspaceID;references:ID" json:"members"`
	Projects    []Project         `gorm:"foreignKey:WorkspaceID;references:ID" json:"projects"`
	CreatedAt   *time.Time        `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy   string            `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt   *time.Time        `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy   string            `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type WorkspaceMember struct {
	No          int        `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID          string     `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	WorkspaceID string     `gorm:"column:workspace_id;not null;index" json:"workspace_id" form:"workspace_id"`
	UserID      string     `gorm:"column:user_id;not null;index" json:"user_id" form:"user_id"`
	User        User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Role        string     `gorm:"column:role;default:'MEMBER'" json:"role" form:"role"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy   string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy   string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

// -- Tabel workspace
// CREATE TABLE workspaces (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     name VARCHAR(255) NOT NULL,
//     description TEXT,
//     color VARCHAR(50),
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255),
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255)
// );

// -- Tabel workspace_members
// CREATE TABLE workspace_members (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     workspace_id VARCHAR(255) NOT NULL,
//     user_id VARCHAR(255) NOT NULL,
//     role VARCHAR(25) DEFAULT 'MEMBER',
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255),
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255),

//     CONSTRAINT fk_workspace FOREIGN KEY(workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
//     CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
// );

// -- Index tambahan
// CREATE INDEX idx_workspace_members_workspace_id ON workspace_members(workspace_id);
// CREATE INDEX idx_workspace_members_user_id ON workspace_members(user_id);
