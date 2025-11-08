package request

type FilterCreateRequest struct {
	Name     string `json:"name" binding:"required,max=255"`
	Endpoint string `json:"endpoint" binding:"required,max=255"`
	Value    string `json:"value" binding:"required,min=1"`
}

type FilterUpdateRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=255"`
	Value *string `json:"value" binding:"omitempty,min=1"`
}
