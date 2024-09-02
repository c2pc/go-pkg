package request

type RoleCreateRequest struct {
	Name  string `json:"name" binding:"required,max=255,min=2"`
	Write []int  `json:"write" binding:"required"`
	Read  []int  `json:"read" binding:"required"`
	Exec  []int  `json:"exec" binding:"required"`
}

type RoleUpdateRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=255,min=2"`
	Write []int   `json:"write" binding:"omitempty"`
	Read  []int   `json:"read" binding:"omitempty"`
	Exec  []int   `json:"exec" binding:"omitempty"`
}
