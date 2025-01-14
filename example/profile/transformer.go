package profile

type Transformer[Model Profile] struct {
}

func NewTransformer[Model Profile]() *Transformer[Model] {
	return &Transformer[Model]{}
}

type Transform struct {
	Age     *int   `json:"age"`
	Height  *int   `json:"height"`
	Address string `json:"address"`
}

func (r Transformer[Model]) Transform(m *Model) interface{} {
	if m == nil {
		return nil
	}

	prof := Profile(*m)

	return &Transform{
		Age:     prof.Age,
		Height:  prof.Height,
		Address: prof.Address,
	}
}

func (r Transformer[Model]) TransformList(models []Model) []interface{} {
	if models == nil {
		return nil
	}

	transformed := make([]interface{}, 0, len(models))
	for _, model := range models {
		prof := Profile(model)
		transformed = append(transformed, Transform{
			Age:     prof.Age,
			Height:  prof.Height,
			Address: prof.Address,
		})
	}

	return transformed
}

func (r Transformer[Model]) TransformProfile(m *Model) interface{} {
	if m == nil {
		return nil
	}

	prof := Profile(*m)

	return &Transform{
		Age:     prof.Age,
		Height:  prof.Height,
		Address: prof.Address,
	}
}
