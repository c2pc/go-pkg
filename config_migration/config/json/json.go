package json

import (
	"encoding/json"
	"fmt"
	"github.com/c2pc/go-pkg/config_migration/config"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
	"io"
	nurl "net/url"
	"os"
	"path/filepath"
)

func init() {
	y := Json{}
	config.Register("json", &y)
}

var (
	ErrNilConfig    = fmt.Errorf("no config")
	ErrNoConfigPath = fmt.Errorf("no config path")
)

type Config struct {
	ConfigPath string
}

type Json struct {
	file *lockedfile.File

	config *Config
}

func New(config *Config) (*Json, error) {
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

	js := &Json{
		config: &Config{
			ConfigPath: path,
		},
	}

	return js, nil
}

func (j *Json) Open(configPath string) (config.Driver, error) {
	js, err := New(&Config{ConfigPath: configPath})
	if err != nil {
		return nil, err
	}

	file, err := lockedfile.OpenFile(js.config.ConfigPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	js.file = file

	return js, nil
}

func (j *Json) Close() error {
	if j.file != nil {
		if err := j.file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (j *Json) Lock() error {
	return nil
}

func (j *Json) Unlock() error {
	return j.Close()
}

func (j *Json) Run(migration io.Reader) error {
	migrFile, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := json.Unmarshal(migrFile, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = j.file.Seek(0, 0); err != nil {
		return err
	}

	configFile, err := io.ReadAll(j.file)
	if err != nil {
		return err
	}

	if len(configFile) == 0 {
		configFile = []byte("{}")
	}

	configMap := map[string]interface{}{}
	if err := json.Unmarshal(configFile, &configMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", j.config.ConfigPath)
	}

	base := mergeMaps(migrMap, configMap)
	base = clearMaps(base, migrMap)
	delete(base, "version")

	data, err := json.MarshalIndent(base, "", "    ")
	if err != nil {
		return err
	}

	newData := string(data)
	newData = fmt.Sprintf(`{
    "version": %v,`, migrMap["version"]) + newData[1:]

	err = j.file.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = j.file.Seek(0, 0); err != nil {
		return err
	}

	_, err = j.file.Write([]byte(newData))
	if err != nil {
		return err
	}

	return nil
}

func (j *Json) Version() (int, error) {
	type version struct {
		Version int `json:"version"`
	}

	if _, err := j.file.Seek(0, 0); err != nil {
		return 0, err
	}

	r, err := io.ReadAll(j.file)
	if err != nil {
		return 0, err
	}

	if len(r) == 0 {
		return config.NilVersion, nil
	}

	v := new(version)
	if err := json.Unmarshal(r, v); err != nil {
		return 0, err
	}

	if v.Version == 0 {
		return 0, errors.New("not found config version into file " + j.config.ConfigPath)
	}

	return v.Version, nil
}

func (j *Json) Drop() error {
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
		if k[:1] == "_" {
			continue
		}
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := d[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = clearMaps(v, bv)
				}
			} else {
				out["_"+k] = clearMaps(v, map[string]interface{}{})
			}
			continue
		}

		if _, ok := d[k]; !ok {
			if _, ok := out["_"+k]; !ok {
				out["_"+k] = v
			}
		} else {
			out[k] = v
		}
	}

	return out
}
