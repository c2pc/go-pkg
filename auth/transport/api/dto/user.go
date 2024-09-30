package dto

import (
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/request"
)

func UserCreate(input *request.UserCreateRequest) service.UserCreateInput {
	return service.UserCreateInput{
		Login:      input.Login,
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		LastName:   input.LastName,
		Password:   input.Password,
		Email:      input.Email,
		Phone:      input.Phone,
		Roles:      input.Roles,
	}
}

func UserUpdate(input *request.UserUpdateRequest) service.UserUpdateInput {
	return service.UserUpdateInput{
		Login:      input.Login,
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		LastName:   input.LastName,
		Password:   input.Password,
		Email:      input.Email,
		Phone:      input.Phone,
		Roles:      input.Roles,
	}
}
