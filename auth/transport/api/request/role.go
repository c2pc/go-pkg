package request

type RoleCreateRequest struct {
	Name  string `json:"name" binding:"required,max=255,min=2,dot_underscore_hyphen"`
	Write []int  `json:"write" binding:"required,dive,gte=1"`
	Read  []int  `json:"read" binding:"required,dive,gte=1"`
	Exec  []int  `json:"exec" binding:"required,dive,gte=1"`
}

type RoleUpdateRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	Write []int   `json:"write" binding:"omitempty,dive,gte=1"`
	Read  []int   `json:"read" binding:"omitempty,dive,gte=1"`
	Exec  []int   `json:"exec" binding:"omitempty,dive,gte=1"`
}
