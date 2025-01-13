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

func (r Transformer[Model]) TransformList(m *Model) interface{} {
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
