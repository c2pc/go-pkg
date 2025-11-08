package model

import (
	"fmt"
	"time"
)

type User struct {
	ID         int64   `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Password   *string `json:"password"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Blocked    bool    `json:"blocked"`
	IsDomain   bool    `json:"is_domain"`

	Roles   []Role      `json:"roles" gorm:"many2many:auth_user_roles;"`
	Profile interface{} `json:"profile" gorm:"-" db:"-"`
}

func (m User) TableName() string {
	return "auth_users"
}

func GeneratePassword(password string, id int64) string {
	return fmt.Sprintf("%s_%d", password, id)
}

type UserRole struct {
	UserID int64 `json:"user_id"`
	RoleID int   `json:"role_id"`

	User *User `json:"user"`
}

func (m UserRole) TableName() string {
	return "auth_user_roles"
}

type UserBlocked struct {
	UserID    int64     `json:"user_id"`
	BlockedAt time.Time `json:"blocked_at"`

	User *User `json:"user"`
}

func (m UserBlocked) TableName() string {
	return "auth_users_blocked"
}
