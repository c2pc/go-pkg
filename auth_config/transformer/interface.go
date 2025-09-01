package transformer

type AuthConfigTransformers map[string]AuthConfigTransformer

type AuthConfigTransformer interface {
	Check(data []byte) error
	AfterUpdate(data []byte) error
	Init() ([]byte, error)
	Transform(data []byte) (any, error)
}
