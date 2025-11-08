package profile

import "github.com/c2pc/go-pkg/v2/auth/profile"

type Transformer struct {
}

func NewTransformer() *Transformer {
	return &Transformer{}
}

type Transform struct {
	Age     *int   `json:"age"`
	Height  *int   `json:"height"`
	Address string `json:"address"`
}

func (r Transformer) Transform(m profile.IModel) interface{} {
	if m == nil {
		return nil
	}

	prof := (m).(*Profile)

	return &Transform{
		Age:     prof.Age,
		Height:  prof.Height,
		Address: prof.Address,
	}
}

func (r Transformer) TransformList(models []profile.IModel) []interface{} {
	if models == nil {
		return nil
	}

	transformed := make([]interface{}, 0, len(models))
	for _, model := range models {
		prof := (model).(Profile)
		transformed = append(transformed, Transform{
			Age:     prof.Age,
			Height:  prof.Height,
			Address: prof.Address,
		})
	}

	return transformed
}

func (r Transformer) TransformProfile(m profile.IModel) interface{} {
	if m == nil {
		return nil
	}

	prof := (m).(*Profile)

	return &Transform{
		Age:     prof.Age,
		Height:  prof.Height,
		Address: prof.Address,
	}
}
