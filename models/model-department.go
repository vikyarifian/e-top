package models

import "time"

type Department struct {
	No          int                `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID          string             `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Name        string             `gorm:"column:name;" json:"name" form:"name"`
	Description string             `gorm:"column:description;" json:"description" form:"description"`
	DeptHead    User               `gorm:"foreignKey:DeptHeadID;references:ID" json:"manager"`
	DeptHeadID  string             `gorm:"column:dept_head_id" json:"dept_head_id,omitempty" form:"dept_head_id"`
	Members     []DepartmentMember `gorm:"foreignKey:DepartmentID;references:ID" json:"members"`
	Color       string             `gorm:"column:color;" json:"color" form:"color"`
	CreatedAt   *time.Time         `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy   string             `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt   *time.Time         `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy   string             `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type DepartmentMember struct {
	No           int        `gorm:"column:no;primaryKey" json:"-" form:"-"`
	DepartmentID string     `gorm:"column:department_id;not null;index" json:"department_id" form:"department_id"`
	UserID       string     `gorm:"column:user_id;not null;index" json:"user_id" form:"user_id"`
	User         User       `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Role         string     `gorm:"column:role;default:'MEMBER'" json:"role" form:"role"`
	CreatedAt    *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy    string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy    string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

// -- Tabel departments
// CREATE TABLE departments (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     name VARCHAR(255) NOT NULL,
//     description TEXT,
//     color VARCHAR(50),
//     dept_head_id VARCHAR(255),
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255),
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255)
// );

// -- Tabel department_members
// CREATE TABLE department_members (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     department_id VARCHAR(255) NOT NULL,
//     user_id VARCHAR(255) NOT NULL,
//     role VARCHAR(25) DEFAULT 'MEMBER',
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255),
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255),
//     CONSTRAINT fk_department FOREIGN KEY(department_id) REFERENCES departments(id) ON DELETE CASCADE,
//     CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
// );
