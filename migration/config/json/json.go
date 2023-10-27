package json

import (
	"encoding/json"
	"fmt"
	"github.com/c2pc/go-pkg/migration/config"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/lockedfile"
	"io"
	nurl "net/url"
	"os"
	"path/filepath"
	"sync"
)

func init() {
	j := Json{}
	database.Register("json", &j)
}

var (
	ErrNilFile = fmt.Errorf("no file")
	ErrNoPath  = fmt.Errorf("no file path")
)

type Config struct {
	Path        string
	WithComment string
}

type Json struct {
	lockedfile *lockedfile.File
	mu         sync.Mutex
	config     *Config
}

func New(config *Config) (*Json, error) {
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

	js := &Json{
		config: &Config{
			Path:        path,
			WithComment: config.WithComment,
		},
	}

	return js, nil
}

func (j *Json) Open(filePath string) (database.Driver, error) {
	js, err := New(&Config{Path: filePath})
	if err != nil {
		return nil, err
	}

	return js, nil
}

func (j *Json) Close() error {
	return j.lockedfile.Close()
}

func (j *Json) Lock() error {
	f, err := lockedfile.OpenFile(j.config.Path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	j.mu.Lock()

	j.lockedfile = f

	return nil
}

func (j *Json) Unlock() error {
	j.mu.Unlock()
	return j.Close()
}

func (j *Json) Run(migration io.Reader) error {
	migrData, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrMap := map[string]interface{}{}
	if err := json.Unmarshal(migrData, &migrMap); err != nil {
		return errors.Wrapf(err, "failed to parse migration file")
	}

	if _, err = j.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(j.lockedfile)
	if err != nil {
		return err
	}

	if len(fileData) == 0 {
		fileData = []byte("{}")
	}

	fileMap := map[string]interface{}{}
	if err := json.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", j.config.Path)
	}

	base := map[string]interface{}{}
	if j.config.WithComment == "" {
		base = config.Merge(migrMap, fileMap)
	} else {
		base = config.MergeWithComment(migrMap, fileMap, j.config.WithComment)
		delete(base, j.config.WithComment+"version")
		delete(base, j.config.WithComment+"force")
	}

	delete(base, "version")
	delete(base, "force")

	data, err := json.MarshalIndent(base, "", "    ")
	if err != nil {
		return err
	}

	err = j.lockedfile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = j.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	_, err = j.lockedfile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (j *Json) SetVersion(version int, dirty bool) error {
	if _, err := j.lockedfile.Seek(0, 0); err != nil {
		return err
	}

	fileData, err := io.ReadAll(j.lockedfile)
	if err != nil {
		return err
	}

	if len(fileData) == 0 {
		fileData = []byte("{}")
	}

	fileMap := map[string]interface{}{}
	if err := json.Unmarshal(fileData, &fileMap); err != nil {
		return errors.Wrapf(err, "failed to parse %s", j.config.Path)
	}

	if version >= 0 || (version == database.NilVersion && dirty) {
		delete(fileMap, "version")
		delete(fileMap, "force")

		data, err := json.MarshalIndent(fileMap, "", "    ")
		if err != nil {
			return err
		}

		newData := string(data)
		if len(fileMap) == 0 {
			newData = fmt.Sprintf(`{
    "version": %v,
    "force": %v
    `, version, dirty) + newData[1:]
		} else {
			newData = fmt.Sprintf(`{
    "version": %v,
    "force": %v,`, version, dirty) + newData[1:]
		}

		err = j.lockedfile.Truncate(0)
		if err != nil {
			return err
		}

		if _, err = j.lockedfile.Seek(0, 0); err != nil {
			return err
		}

		_, err = j.lockedfile.Write([]byte(newData))
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *Json) Version() (int, bool, error) {
	type version struct {
		Version int  `json:"version"`
		Force   bool `json:"force"`
	}

	if _, err := j.lockedfile.Seek(0, 0); err != nil {
		return 0, false, err
	}

	r, err := io.ReadAll(j.lockedfile)
	if err != nil {
		return 0, false, err
	}

	if len(r) == 0 {
		return database.NilVersion, false, nil
	}

	v := new(version)
	if err := json.Unmarshal(r, v); err != nil {
		return 0, false, err
	}

	if v.Version == 0 {
		return database.NilVersion, false, nil
	}

	return v.Version, v.Force, nil
}

func (j *Json) Drop() error {
	err := j.lockedfile.Truncate(0)
	if err != nil {
		return err
	}

	if _, err = j.lockedfile.Seek(0, 0); err != nil {
		return err
	}

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
