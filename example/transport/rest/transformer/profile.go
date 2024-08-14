package transformer

import (
	"github.com/c2pc/go-pkg/v2/example/model"
)

type ProfileTransformer[Model model.Profile] struct {
}

func NewProfileTransformer[Model model.Profile]() *ProfileTransformer[Model] {
	return &ProfileTransformer[Model]{}
}

type ProfileTransform struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

func (r ProfileTransformer[Model]) Transform(m *Model) interface{} {
	if m == nil {
		return nil
	}

	prof := model.Profile(*m)

	return &ProfileTransform{
		Login: prof.Login,
		Name:  prof.Name,
	}
}

func (r ProfileTransformer[Model]) TransformList(m *Model) interface{} {
	if m == nil {
		return nil
	}

	prof := model.Profile(*m)

	return &ProfileTransform{
		Login: prof.Login,
		Name:  prof.Name,
	}
}

func (r ProfileTransformer[Model]) TransformProfile(m *Model) interface{} {
	if m == nil {
		return nil
	}

	prof := model.Profile(*m)

	return &ProfileTransform{
		Login: prof.Login,
		Name:  prof.Name,
	}
}
