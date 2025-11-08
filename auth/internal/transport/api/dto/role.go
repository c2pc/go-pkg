package dto

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
)

func RoleCreate(input *request.RoleCreateRequest) service.RoleCreateInput {
	return service.RoleCreateInput{
		Name:  input.Name,
		Write: input.Write,
		Read:  input.Read,
		Exec:  input.Exec,
	}
}

func RoleUpdate(input *request.RoleUpdateRequest) service.RoleUpdateInput {
	return service.RoleUpdateInput{
		Name:  input.Name,
		Write: input.Write,
		Read:  input.Read,
		Exec:  input.Exec,
	}
}
