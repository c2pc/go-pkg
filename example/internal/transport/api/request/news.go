package request

type NewsCreateRequest struct {
	Title   string  `json:"title" binding:"required,max=255,min=2"`
	Content *string `json:"content" binding:"omitempty,max=1024,min=1"`
}

type NewsUpdateRequest struct {
	Title   *string `json:"title" binding:"omitempty,max=255,min=2"`
	Content *string `json:"content" binding:"omitempty,max=1024,min=1"`
}

type NewsMassUpdateRequest struct {
	Content *string `json:"content" binding:"omitempty,max=1024,min=1"`
}

type NewsImportRequest struct {
	Title   string  `json:"title" binding:"required,max=255,min=2" csv:"title"`
	Content *string `json:"content" binding:"omitempty,max=1024,min=1" csv:"content"`
}
