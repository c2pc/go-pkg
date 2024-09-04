package profile

import (
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	"github.com/gin-gonic/gin"
)

type Request[CreateInput ProfileCreateInput, UpdateInput ProfileUpdateInput, UpdateProfileInput ProfileUpdateProfileInput] struct {
}

func NewRequest[CreateInput ProfileCreateInput, UpdateInput ProfileUpdateInput, UpdateProfileInput ProfileUpdateProfileInput]() *Request[CreateInput, UpdateInput, UpdateProfileInput] {
	return &Request[CreateInput, UpdateInput, UpdateProfileInput]{}
}

type ProfileCreateRequest struct {
	Age     *int   `json:"age" binding:"omitempty,gte=0"`
	Height  *int   `json:"height" binding:"omitempty,gte=0"`
	Address string `json:"address" binding:"required,min=1,max=255"`
}

func (r Request[CreateInput, UpdateInput, UpdateProfileInput]) CreateRequest(c *gin.Context) (*CreateInput, error) {
	type Profile struct {
		Profile ProfileCreateRequest `json:"profile" binding:"required"`
	}

	cred, err := request2.BindJSON[Profile](c)
	if err != nil {
		return nil, err
	}

	input := CreateInput{
		Age:     cred.Profile.Age,
		Height:  cred.Profile.Height,
		Address: cred.Profile.Address,
	}

	data := CreateInput(input)

	return &data, nil
}

type ProfileUpdateRequest struct {
	Age     *int    `json:"age" binding:"omitempty,gte=0"`
	Height  *int    `json:"height" binding:"omitempty,gte=0"`
	Address *string `json:"address" binding:"omitempty,min=1,max=255"`
}

func (r Request[CreateInput, UpdateInput, UpdateProfileInput]) UpdateRequest(c *gin.Context) (*UpdateInput, error) {
	type Profile struct {
		Profile *ProfileUpdateRequest `json:"profile" binding:"omitempty"`
	}

	cred, err := request2.BindJSON[Profile](c)
	if err != nil {
		return nil, err
	}

	if cred.Profile == nil {
		return nil, nil
	}

	input := UpdateInput{
		Age:     cred.Profile.Age,
		Height:  cred.Profile.Height,
		Address: cred.Profile.Address,
	}

	data := UpdateInput(input)

	return &data, nil
}

type UpdateProfileRequest struct {
	Age     *int    `json:"age" binding:"omitempty,gte=0"`
	Height  *int    `json:"height" binding:"omitempty,gte=0"`
	Address *string `json:"address" binding:"omitempty,min=1,max=255"`
}

func (r Request[CreateInput, UpdateInput, UpdateProfileInput]) UpdateProfileRequest(c *gin.Context) (*UpdateProfileInput, error) {
	type Profile struct {
		Profile *UpdateProfileRequest `json:"profile" binding:"omitempty"`
	}

	cred, err := request2.BindJSON[Profile](c)
	if err != nil {
		return nil, err
	}

	if cred.Profile == nil {
		return nil, nil
	}

	input := UpdateProfileRequest{
		Age:     cred.Profile.Age,
		Height:  cred.Profile.Height,
		Address: cred.Profile.Address,
	}

	data := UpdateProfileInput(input)

	return &data, nil
}
