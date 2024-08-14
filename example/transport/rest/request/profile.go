package request

import (
	"github.com/c2pc/go-pkg/v2/example/service"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	"github.com/gin-gonic/gin"
)

type ProfileRequest[CreateInput service.ProfileCreateInput, UpdateInput service.ProfileUpdateInput, UpdateProfileInput service.ProfileUpdateProfileInput] struct {
}

func NewProfileRequest[CreateInput service.ProfileCreateInput, UpdateInput service.ProfileUpdateInput, UpdateProfileInput service.ProfileUpdateProfileInput]() *ProfileRequest[CreateInput, UpdateInput, UpdateProfileInput] {
	return &ProfileRequest[CreateInput, UpdateInput, UpdateProfileInput]{}
}

type ProfileCreateRequest struct {
	Login string `json:"login" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

func (r ProfileRequest[CreateInput, UpdateInput, UpdateProfileInput]) CreateRequest(c *gin.Context) (*CreateInput, error) {
	type Profile struct {
		Profile ProfileCreateRequest `json:"profile" binding:"required"`
	}

	cred, err := request2.BindJSON[Profile](c)
	if err != nil {
		return nil, err
	}

	input := service.ProfileCreateInput{
		Login: cred.Profile.Login,
		Name:  cred.Profile.Name,
	}

	data := CreateInput(input)

	return &data, nil
}

type ProfileUpdateRequest struct {
	Login *string `json:"login" binding:"omitempty"`
	Name  *string `json:"name" binding:"omitempty"`
}

func (r ProfileRequest[CreateInput, UpdateInput, UpdateProfileInput]) UpdateRequest(c *gin.Context) (*UpdateInput, error) {
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

	input := service.ProfileUpdateInput{
		Login: cred.Profile.Login,
		Name:  cred.Profile.Name,
	}

	data := UpdateInput(input)

	return &data, nil
}

type ProfileUpdateProfileInput struct {
	Login *string `json:"login" binding:"omitempty"`
	Name  *string `json:"name" binding:"omitempty"`
}

func (r ProfileRequest[CreateInput, UpdateInput, UpdateProfileInput]) UpdateProfileRequest(c *gin.Context) (*UpdateProfileInput, error) {
	type Profile struct {
		Profile *ProfileUpdateProfileInput `json:"profile" binding:"omitempty"`
	}

	cred, err := request2.BindJSON[Profile](c)
	if err != nil {
		return nil, err
	}

	if cred.Profile == nil {
		return nil, nil
	}

	input := service.ProfileUpdateProfileInput{
		Login: cred.Profile.Login,
		Name:  cred.Profile.Name,
	}

	data := UpdateProfileInput(input)

	return &data, nil
}
