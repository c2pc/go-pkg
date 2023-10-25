package yaml

import (
	"fmt"
	"github.com/c2pc/go-pkg/config_migration/config"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
	"gopkg.in/yaml.v3"
	"io"
	nurl "net/url"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	y := Yaml{}
	config.Register("yaml", &y)
}

var (
	ErrNilConfig    = fmt.Errorf("no config")
	ErrNoConfigPath = fmt.Errorf("no config path")
)

type Config struct {
	ConfigPath string
}

type Yaml struct {
	file *lockedfile.File

	config *Config
}

func New(config *Config) (*Yaml, error) {
	if config == nil {
		return nil, ErrNilConfig
	}

	if config.ConfigPath == "" {
		return nil, ErrNoConfigPath
	}

	path, err := parseURL(config.ConfigPath)
	if err != nil {
		return nil, err
	}

	yml := &Yaml{
		config: &Config{
			ConfigPath: path,
		},
	}

	return yml, nil
}

func (y *Yaml) Open(configPath string) (config.Driver, error) {
	yml, err := New(&Config{ConfigPath: configPath})
	if err != nil {
		return nil, err
	}

	file, err := lockedfile.OpenFile(yml.config.ConfigPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	yml.file = file

	return yml, nil
}

func (y *Yaml) Close() error {
	if y.file != nil {
		if err := y.file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (y *Yaml) Lock() error {
	return nil
}

func (y *Yaml) Unlock() error {
	return y.Close()
}

func (y *Yaml) Run(migration io.Reader) error {
	migrFile, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := yaml.Unmarshal(migrFile, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = y.file.Seek(0, 0); err != nil {
		return err
	}

	configFile, err := io.ReadAll(y.file)
	if err != nil {
		return err
	}

	configMap := map[string]interface{}{}
	if err := yaml.Unmarshal(configFile, &configMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", y.config.ConfigPath)
	}

	base := mergeMaps(migrMap, configMap)
	base = clearMaps(base, migrMap)
	delete(base, "version")

	data, err := yaml.Marshal(base)
	if err != nil {
		return err
	}
	newData := strings.ReplaceAll(string(data), "'", "")
	newData = strings.ReplaceAll(newData, "null", "")

	newData = fmt.Sprintf("version: %v\n", migrMap["version"]) + newData

	err = y.file.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = y.file.Seek(0, 0); err != nil {
		return err
	}

	_, err = y.file.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

func (y *Yaml) Version() (int, error) {
	type version struct {
		Version int `yaml:"version"`
	}

	if _, err := y.file.Seek(0, 0); err != nil {
		return 0, err
	}

	r, err := io.ReadAll(y.file)
	if err != nil {
		return 0, err
	}

	v := new(version)
	if err := yaml.Unmarshal(r, v); err != nil {
		return 0, err
	}

	if v.Version == 0 {
		return config.NilVersion, nil
		/*if len(r) == 0 {
			return config.NilVersion, nil
		} else {
			return 0, errors.New("not found config version into file " + y.config.ConfigPath)
		}*/
	}

	return v.Version, nil
}

func (y *Yaml) Drop() error {
	return nil
}

func parseURL(url string) (string, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", err
	}
	// concat host and path to restore full path
	// host might be `.`
	p := u.Opaque
	if len(p) == 0 {
		p = u.Host + u.Path
	}

	if len(p) == 0 {
		// default to current directory if no path
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		p = wd

	} else if p[0:1] == "." || p[0:1] != "/" {
		// make path absolute if relative
		abs, err := filepath.Abs(p)
		if err != nil {
			return "", err
		}
		p = abs
	}
	return p, nil
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v2, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v2)
					continue
				}
			}
			out[k] = v2
			continue
		}

		out[k] = v
	}

	return out
}

func clearMaps(c, d map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range c {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := d[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = clearMaps(v, bv)
				}
			} else {
				out["#"+k] = clearMaps(v, map[string]interface{}{})
			}
			continue
		}

		if _, ok := d[k]; !ok {
			if _, ok := out["#"+k]; !ok {
				out["#"+k] = v
			}
		} else {
			out[k] = v
		}
	}

	return out
}
