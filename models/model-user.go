package models

import "time"

type User struct {
	No            int        `gorm:"column:no;primaryKey" json:"-,omitempty" form:"-"`
	ID            string     `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Username      string     `gorm:"column:username;unique" json:"username,omitempty" form:"username"`
	FullName      string     `gorm:"column:full_name;" json:"full_name,omitempty" form:"full_name"`
	Email         string     `gorm:"column:email;unique" json:"email,omitempty" form:"email"`
	Password      string     `gorm:"column:password" json:"password,omitempty" form:"password"`
	Level         string     `gorm:"column:level" json:"level,omitempty" form:"level"`
	VerifiedEmail bool       `gorm:"column:verified_email" json:"-" form:"verified_email"`
	CreatedAt     *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy     string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy     string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

// CREATE TABLE users (
//     no SERIAL PRIMARY KEY,
//     id VARCHAR(128) NOT NULL DEFAULT '00000000000000',
//     username VARCHAR(50) UNIQUE NOT NULL DEFAULT ' ',
//     full_name VARCHAR(50) NOT NULL DEFAULT ' ',
//     email VARCHAR(50) NOT NULL DEFAULT ' ',
//     password VARCHAR(128) UNIQUE NOT NULL DEFAULT ' ',
//     level VARCHAR(10) NOT NULL DEFAULT 'USER',
//     verified_email BOOLEAN NOT NULL DEFAULT false,
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(128) DEFAULT '',
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(128) DEFAULT ''
// );
