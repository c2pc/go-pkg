package configurator

type File interface {
	GetKey() string
	GetData() []byte
	GetName() string
}

type Config interface {
	GetFiles() []File
	GetKey() string
	GetValue() []byte
}

type Configs map[string]Configurator

type Configurator interface {
	Unmarshal(cfg Config) (any, error)
	Check(newCfg, lastCfg Config) ([]byte, error)
	AfterUpdate(cfg Config) error
	Init() ([]byte, error)
	Transform(cfg Config) ([]byte, error)
	Action() string
}
