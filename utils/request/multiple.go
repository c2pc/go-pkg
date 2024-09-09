package request

type MultipleCreateRequest[T any] struct {
	Data []T `json:"data" binding:"required,min=1,max=100"`
}

type MultipleUpdateRequest[T any] struct {
	Data []T `json:"data" binding:"required,min=1,max=100"`
}

type MultipleDeleteRequest struct {
	Data []int `json:"data" binding:"required,min=1,max=100"`
}
