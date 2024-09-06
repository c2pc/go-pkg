package request

type MultipleCreateRequest[T any] []T
type MultipleUpdateRequest struct {
	ID int `json:"id" binding:"required,gt=0"`
}
type MultipleDeleteRequest []int
