package file_manager

import (
	"errors"
	"path"
	"strings"
)

type Config struct {
	Name       string
	WorkDir    string
	RemoveFile bool
	RemoveDir  bool
	CreateFile bool
	CreateDir  bool
	Download   bool
}

func (c *Config) validate() error {
	if err := c.checkName(); err != nil {
		return err
	}

	if err := c.checkWorkDir(); err != nil {
		return err
	}

	return nil
}

func (c *Config) checkName() error {
	if c.Name == "" {
		return errors.New("name is empty")
	}

	return nil
}

func (c *Config) checkWorkDir() error {
	if c.WorkDir == "" {
		return errors.New("work dir is empty")
	}

	if c.WorkDir == "/" {
		return errors.New("work dir is root")
	}

	if strings.Contains(c.WorkDir, "../") || strings.Contains(c.WorkDir, "/..") {
		return errors.New("work dir has invalid symbols")
	}

	if c.WorkDir[len(c.WorkDir)-1] == '/' {
		c.WorkDir = c.WorkDir[:len(c.WorkDir)-1]
	}

	if !path.IsAbs(c.WorkDir) {
		return errors.New("work dir is not absolute")
	}

	return nil
}
