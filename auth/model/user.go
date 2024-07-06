package model

type User struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
	Password   string  `json:"password"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`

	Roles []Role `json:"roles" gorm:"many2many:user_roles;"`
}

func (m User) TableName() string {
	return "users"
}

type UserRole struct {
	UserID int `json:"user_id"`
	RoleID int `json:"role_id"`
}

func (m UserRole) TableName() string {
	return "user_roles"
}
