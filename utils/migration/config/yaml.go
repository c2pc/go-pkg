package config

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
	"sync"
)

func init() {
	y := Yaml{}
	database.Register("yaml", &y)
}

type YamlConfig struct {
	Path        string
	WithComment string
}

type Yaml struct {
	lockedfile *lockedfile.File
	mu         sync.Mutex
	config     *YamlConfig
}

func NewYaml(config *YamlConfig) (*Yaml, error) {
	if config == nil {
		return nil, ErrNilFile
	}

	if config.Path == "" {
		return nil, ErrNoPath
	}

	path, err := parseURL(config.Path)
	if err != nil {
		return nil, err
	}

	yml := &Yaml{
		config: &YamlConfig{
			Path:        path,
			WithComment: config.WithComment,
		},
	}

	return yml, nil
}

func (y *Yaml) Open(filePath string) (database.Driver, error) {
	js, err := NewYaml(&YamlConfig{Path: filePath})
	if err != nil {
		return nil, err
	}

	return js, nil
}

func (y *Yaml) Close() error {
	return y.lockedfile.Close()
}

func (y *Yaml) Lock() error {
	f, err := lockedfile.OpenFile(y.config.Path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	y.mu.Lock()

	y.lockedfile = f

	return nil
}

func (y *Yaml) Unlock() error {
	y.mu.Unlock()
	return y.Close()
}

func (y *Yaml) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := yaml.Unmarshal(migrData, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = y.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(y.lockedfile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := yaml.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", y.config.Path)
	}

	base := map[string]interface{}{}
	if y.config.WithComment == "" {
		base = Merge(migrMap, fileMap)
	} else {
		base = MergeWithComment(migrMap, fileMap, y.config.WithComment)
		delete(base, y.config.WithComment+"version")
		delete(base, y.config.WithComment+"force")
	}

	delete(base, "version")
	delete(base, "force")

	data, err := yaml.Marshal(base)
	if err != nil {
		return err
	}
	newData := strings.ReplaceAll(string(data), "'", "")
	newData = strings.ReplaceAll(newData, "null", "")

	err = y.lockedfile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	_, err = y.lockedfile.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

func (y *Yaml) SetVersion(version int, dirty bool) error {
	if _, err := y.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(y.lockedfile)
	if err != nil {
		return err
	}

	fileMap := map[string]interface{}{}
	if err := yaml.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", y.config.Path)
	}

	delete(fileMap, "version")
	delete(fileMap, "force")

	data, err := yaml.Marshal(fileMap)
	if err != nil {
		return err
	}

	newData := strings.ReplaceAll(string(data), "null", "")

	if len(fileMap) == 0 {
		newData = ""
	}

	newData = fmt.Sprintf("force: %v\n", dirty) + newData
	newData = fmt.Sprintf("version: %v\n", version) + newData

	err = y.lockedfile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	_, err = y.lockedfile.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

func (y *Yaml) Version() (int, bool, error) {
	type version struct {
		Version int  `yaml:"version"`
		Force   bool `json:"force"`
	}

	if _, err := y.lockedfile.Seek(0, 0); err != nil {
		return 0, false, err
	}

	r, err := io.ReadAll(y.lockedfile)
	if err != nil {
		return 0, false, err
	}

	if len(r) == 0 {
		return database.NilVersion, false, nil
	}

	v := new(version)
	if err := yaml.Unmarshal(r, v); err != nil {
		return 0, false, err
	}

	if v.Version == 0 {
		return database.NilVersion, false, nil
	}

	return v.Version, v.Force, nil
}

func (y *Yaml) Drop() error {
	err := y.lockedfile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	return nil
}
