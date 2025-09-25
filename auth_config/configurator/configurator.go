package configurator

type Configs map[string]Configurator

type Configurator interface {
	Unmarshal(data []byte) (any, error)
	Check(newData, lastData []byte) ([]byte, error)
	AfterUpdate(data []byte) error
	Init() ([]byte, error)
	Transform(data []byte) ([]byte, error)
	Action() string
}
