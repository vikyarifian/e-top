package models

import (
	"time"
)

type User struct {
	No            int        `gorm:"column:no;primaryKey" json:"-" form:"-"`
	ID            string     `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Username      string     `gorm:"column:username;unique" json:"username,omitempty" form:"username"`
	FullName      string     `gorm:"column:full_name;" json:"full_name,omitempty" form:"full_name"`
	Email         string     `gorm:"column:email;unique" json:"email,omitempty" form:"email"`
	Password      string     `gorm:"column:password" json:"password,omitempty" form:"password"`
	Level         string     `gorm:"column:level" json:"level,omitempty" form:"level"`
	VerifiedEmail bool       `gorm:"column:verified_email" json:"-" form:"verified_email"`
	Color         string     `gorm:"column:color" json:"color" form:"color"`
	CreatedAt     *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy     string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy     string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type ResetPassword struct {
	ID        int        `gorm:"column:id;primaryKey" json:"-,omitempty" form:"-"`
	Email     string     `gorm:"column:email;unique" json:"email,omitempty" form:"email"`
	TokenHash string     `gorm:"column:token_hash;size:64;not null;index:idx_tokenhash"`
	Used      int        `gorm:"column:used;default:0;not null"`
	ExpiresAt time.Time  `gorm:"column:expires_at;type:TIMESTAMP" json:"expires_at" form:"expires_at"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
}

type UserRole struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Color    string `json:"color"`
	Role     string `json:"role"`
}

// CREATE TABLE users (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     username VARCHAR(255) UNIQUE NOT NULL DEFAULT ' ',
//     full_name VARCHAR(255) NOT NULL DEFAULT ' ',
//     email VARCHAR(255) NOT NULL DEFAULT ' ',
//     password TEXT NOT NULL,
//     level VARCHAR(25) NOT NULL DEFAULT 'USER',
//     verified_email BOOLEAN NOT NULL DEFAULT false,
//     color VARCHAR(50) NOT NULL DEFAULT 'bg-gray-500 text-white',
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(128) DEFAULT '',
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(128) DEFAULT ''
// );

// CREATE TABLE reset_passwords (
//     id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     email VARCHAR(255) NOT NULL DEFAULT ' ',
//     token_hash VARCHAR(255) NOT NULL DEFAULT ' ',
//     used INTEGER NOT NULL DEFAULT 0,
//     expires_at TIMESTAMP DEFAULT NOW(),
//     created_at TIMESTAMP DEFAULT NOW(),
//     updated_at TIMESTAMP DEFAULT NOW()
// );
