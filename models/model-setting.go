package models

import "time"

type Setting struct {
	No        int        `gorm:"column:no;primaryKey;autoIncrement" json:"-" form:"-"`
	ID        string     `gorm:"column:id;unique" json:"id,omitempty" form:"id"`
	Name      string     `gorm:"column:name;" json:"name" form:"name"`
	Label     string     `gorm:"column:label;" json:"label" form:"label"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

type UserSetting struct {
	No        int        `gorm:"column:no;unique;primaryKey;autoIncrement" json:"no" form:"no"`
	SettingID string     `gorm:"column:setting_id;" json:"setting_id" form:"setting_id"`
	Setting   Setting    `gorm:"foreignKey:SettingID;references:ID" json:"setting"`
	UserID    string     `gorm:"column:user_id;" json:"user_id" form:"user_id"`
	User      User       `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Value     bool       `gorm:"column:value;" json:"value" form:"value"`
	CreatedAt *time.Time `gorm:"column:created_at;type:TIMESTAMP" json:"created_at,omitempty" form:"created_at"`
	CreatedBy string     `gorm:"column:created_by" json:"created_by,omitempty" form:"created_by"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at,omitempty" form:"updated_at"`
	UpdatedBy string     `gorm:"column:updated_by" json:"updated_by,omitempty" form:"updated_by"`
}

func (UserSetting) TableName() string {
	return "users_settings"
}

// -- Tabel settings
// CREATE TABLE settings (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     id VARCHAR(255) UNIQUE NOT NULL DEFAULT '00000000000000',
//     name VARCHAR(255) NOT NULL,
//     label VARCHAR(255) NOT NULL,
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255),
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255)
// );

// CREATE TABLE users_settings (
//     no INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
//     setting_id VARCHAR(255) NOT NULL,
//     user_id VARCHAR(255) NOT NULL,
//     value boolean NOT NULL,
//     created_at TIMESTAMP DEFAULT NOW(),
//     created_by VARCHAR(255),
//     updated_at TIMESTAMP DEFAULT NOW(),
//     updated_by VARCHAR(255)
// );
