package request

type UserCreateRequest struct {
	Login      string  `json:"login" binding:"required,max=255,min=2,dot_underscore_hyphen"`
	FirstName  string  `json:"first_name" binding:"required,max=255,min=2,dot_underscore_hyphen"`
	SecondName *string `json:"second_name" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	LastName   *string `json:"last_name" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	Password   string  `json:"password" binding:"required,max=255,min=8,spec_chars"`
	Email      *string `json:"email" binding:"omitempty,email,max=255"`
	Phone      *string `json:"phone" binding:"omitempty,max=255,min=1"`
	Roles      []int   `json:"roles" binding:"required,dive,gte=1"`
}

type UserUpdateRequest struct {
	Login      *string `json:"login" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	FirstName  *string `json:"first_name" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	SecondName *string `json:"second_name" binding:"omitempty,len=0|min=2,max=255,dot_underscore_hyphen"`
	LastName   *string `json:"last_name" binding:"omitempty,len=0|min=2,max=255,dot_underscore_hyphen"`
	Password   *string `json:"password" binding:"omitempty,max=255,min=8,spec_chars"`
	Email      *string `json:"email" binding:"omitempty,len=0|email,max=255"`
	Phone      *string `json:"phone" binding:"omitempty,len=0|min=1,max=255"`
	Roles      []int   `json:"roles" binding:"omitempty,dive,gte=1"`
}
