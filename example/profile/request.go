package profile

import (
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	"github.com/gin-gonic/gin"
)

type Request struct {
}

func NewRequest() *Request {
	return &Request{}
}

type ProfileCreateRequest struct {
	Age     *int   `json:"age" binding:"omitempty,gte=0"`
	Height  *int   `json:"height" binding:"omitempty,gte=0"`
	Address string `json:"address" binding:"omitempty,min=1,max=255"`
}

func (r Request) CreateRequest(c *gin.Context) (any, error) {
	type Profile struct {
		Profile ProfileCreateRequest `json:"profile" binding:"omitempty"`
	}

	cred, err := request2.BindJSON[Profile](c)
	if err != nil {
		return nil, err
	}

	input := ProfileCreateInput{
		Age:     cred.Profile.Age,
		Height:  cred.Profile.Height,
		Address: cred.Profile.Address,
	}

	return &input, nil
}

type ProfileUpdateRequest struct {
	Age     *int    `json:"age" binding:"omitempty,gte=0"`
	Height  *int    `json:"height" binding:"omitempty,gte=0"`
	Address *string `json:"address" binding:"omitempty,min=1,max=255"`
}

func (r Request) UpdateRequest(c *gin.Context) (any, error) {
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

	input := ProfileUpdateInput{
		Age:     cred.Profile.Age,
		Height:  cred.Profile.Height,
		Address: cred.Profile.Address,
	}

	return &input, nil
}
