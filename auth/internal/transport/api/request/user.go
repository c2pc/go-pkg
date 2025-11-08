package request

type UserCreateRequest struct {
	Login      string  `json:"login" binding:"required,max=255,exclude_space"`
	FirstName  string  `json:"first_name" binding:"required,max=255"`
	SecondName *string `json:"second_name" binding:"omitempty,max=255"`
	LastName   *string `json:"last_name" binding:"omitempty,max=255"`
	Password   *string `json:"password" binding:"omitempty,max=60,min=8,exclude_space"`
	Email      *string `json:"email" binding:"omitempty,max=255"`
	Phone      *string `json:"phone" binding:"omitempty,max=255"`
	Roles      []int   `json:"roles" binding:"required,unique,dive,gte=1"`
	Blocked    bool    `json:"blocked" binding:"omitempty"`
	IsDomain   bool    `json:"is_domain" binding:"omitempty"`
}

type UserUpdateRequest struct {
	Login      *string `json:"login" binding:"omitempty,max=255,exclude_space"`
	FirstName  *string `json:"first_name" binding:"omitempty,max=255"`
	SecondName *string `json:"second_name" binding:"omitempty,max=255"`
	LastName   *string `json:"last_name" binding:"omitempty,max=255"`
	Password   *string `json:"password" binding:"omitempty,max=60,len=0|min=8,exclude_space"`
	Email      *string `json:"email" binding:"omitempty,max=255"`
	Phone      *string `json:"phone" binding:"omitempty,max=255"`
	Roles      []int   `json:"roles" binding:"omitempty,unique,dive,gte=1"`
	Blocked    *bool   `json:"blocked" binding:"omitempty"`
	IsDomain   *bool   `json:"is_domain" binding:"omitempty"`
}
