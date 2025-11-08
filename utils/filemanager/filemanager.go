package file_manager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileManager interface {
	Name() string
	GetWorkDir() string
	WalkDir(root string, fn filepath.WalkFunc) error
	RelativePath(path string) string
	Match(name string, filter string) (bool, error)
	CheckDirPath(path string) error
	CheckFilePath(path string) error
	CheckFileExt(path string, extensions ...string) error
	GetFullPath(path string) string
	Info(path string) (fs.FileInfo, error)
	MkDir(path string) error
	TouchFile(path string, name string, file multipart.FileHeader, append bool) (string, error)
	Remove(path string) error
	RemoveDir(path string, onlyChildren bool) error
	CreateFilePermissions() bool
	CreateDirPermissions() bool
	RemoveFilePermissions() bool
	RemoveDirPermissions() bool
	DownloadPermissions() bool
	UntarFile(src, dest string) error
	UntarGzFile(src, dest string) error
	UnzipFile(zipPath, destDir string) error
}

type File struct {
	cfg Config
}

func New(cfg Config) (FileManager, error) {
	f := &File{cfg: cfg}

	if err := f.cfg.validate(); err != nil {
		return nil, err
	}

	if err := f.checkWorkDir(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *File) Name() string {
	return f.cfg.Name
}

func (f *File) GetWorkDir() string {
	return f.cfg.WorkDir
}

func (f *File) CreateFilePermissions() bool {
	return f.cfg.CreateFile
}

func (f *File) CreateDirPermissions() bool {
	return f.cfg.CreateDir
}

func (f *File) RemoveFilePermissions() bool {
	return f.cfg.RemoveFile
}

func (f *File) RemoveDirPermissions() bool {
	return f.cfg.RemoveDir
}

func (f *File) DownloadPermissions() bool {
	return f.cfg.Download
}

func (f *File) checkWorkDir() error {
	dir, err := os.Stat(f.GetWorkDir())
	if err != nil {
		return errors.New("work dir is not exists")
	}

	if !dir.IsDir() {
		return errors.New("work dir is not directory")
	}

	if err := f.IsWritable(f.GetWorkDir()); err != nil {
		return err
	}

	return nil
}

func (f *File) IsWritable(p string) error {
	tmpFile := "tmp"

	file, err := os.CreateTemp(p, tmpFile)
	if err != nil {
		if os.IsPermission(err) {
			return errors.New("work directory is not writable")
		}
		return errors.New("work dir is not writable")
	}

	defer os.Remove(file.Name())
	defer file.Close()

	return nil
}

func (f *File) GetFullPath(path string) string {
	if path == "" {
		return f.GetWorkDir()
	}

	if strings.Contains(path, "..") {
		return ""
	}

	if strings.Contains(path, "./") {
		return ""
	}

	if strings.Contains(path, "~/") {
		return ""
	}

	path = strings.ReplaceAll(path, "//", "/")

	if path == "." {
		return f.GetWorkDir()
	}

	if path != "/" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	if path == "" {
		return f.GetWorkDir()
	}

	if path[0] == '/' {
		return filepath.Join(f.GetWorkDir(), path)
	}

	return filepath.Join(f.GetWorkDir(), path)
}

func (f *File) CheckDirPath(path string) error {
	dir, err := os.Stat(path)
	if err != nil {
		return ErrObjectIsNotExists
	}

	if !dir.IsDir() {
		return ErrObjectIsNotDirectory
	}

	return nil
}

func (f *File) Info(path string) (fs.FileInfo, error) {
	s, err := os.Stat(path)
	if err != nil {
		return nil, ErrObjectIsNotExists
	}

	return s, nil
}

func (f *File) CheckFilePath(path string) error {
	file, err := os.Stat(path)
	if err != nil {
		return ErrObjectIsNotExists
	}

	if file.IsDir() {
		return ErrObjectIsNotFile
	}

	return nil
}

func (f *File) CheckFileExt(path string, extensions ...string) error {
	ext := filepath.Ext(path)

	for _, e := range extensions {
		if e == ext {
			return nil
		}
	}

	return ErrObjectHasInvalidExtension
}

func (f *File) Remove(path string) error {
	err := os.Remove(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrObjectIsNotExists
		}
		if os.IsPermission(err) {
			return ErrObjectIsNotWritable
		}
		return err
	}

	return nil
}

func (f *File) RemoveDir(path string, onlyChildren bool) error {
	if onlyChildren {
		d, err := os.Open(path)
		if err != nil {
			return err
		}
		defer d.Close()
		names, err := d.Readdirnames(-1)
		if err != nil {
			return err
		}
		var err2 error
		for _, name := range names {
			err := os.RemoveAll(filepath.Join(path, name))
			if err != nil {
				if os.IsNotExist(err) {
					return ErrObjectIsNotExists
				}
				if os.IsPermission(err) {
					return ErrObjectIsNotWritable
				}
				err2 = err
				continue
			}
		}
		return err2
	} else {
		err := os.RemoveAll(path)
		if err != nil {
			if os.IsNotExist(err) {
				return ErrObjectIsNotExists
			}
			if os.IsPermission(err) {
				return ErrObjectIsNotWritable
			}
			return err
		}
	}

	return nil
}

func (f *File) WalkDir(root string, fn filepath.WalkFunc) error {
	p := f.GetFullPath(root)
	if p == "" {
		return ErrObjectIsNotExists
	}

	if err := f.CheckDirPath(p); err != nil {
		return err
	}

	return filepath.Walk(p, fn)
}

func (f *File) RelativePath(path string) string {
	if len(path) == len(f.GetWorkDir()) {
		return ""
	}

	return path[len(f.GetWorkDir())+1:]
}

func (f *File) Match(name string, filter string) (bool, error) {
	return regexp.Match(name, []byte(filter))
}

func (f *File) MkDir(path string) error {
	var mode os.FileMode
	dir, err := os.Stat(f.GetWorkDir())
	if err != nil {
		mode = 0755
	} else {
		mode = dir.Mode()
	}

	if err := os.MkdirAll(path, mode); err != nil {
		if os.IsExist(err) {
			return ErrObjectIsAlreadyExists
		}
		if os.IsPermission(err) {
			return ErrObjectIsNotWritable
		}
		return err
	}

	return nil
}

func (f *File) TouchFile(path string, name string, file multipart.FileHeader, append bool) (string, error) {
	fi, err := file.Open()
	if err != nil {
		return "", err
	}
	defer fi.Close()

	var mode os.FileMode
	dir, err := os.Stat(f.GetWorkDir())
	if err != nil {
		mode = 0755
	} else {
		mode = dir.Mode()
	}

	p := filepath.Join(path, name)

	var fo *os.File
	if append {
		fo, err = os.OpenFile(p, os.O_APPEND|os.O_WRONLY|os.O_CREATE, mode)
		if err != nil {
			if os.IsExist(err) {
				return "", ErrObjectIsAlreadyExists
			}
			if os.IsPermission(err) {
				return "", ErrObjectIsNotWritable
			}
			return "", err
		}
		defer fo.Close()
	} else {
		fo, err = os.OpenFile(p, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, mode)
		if err != nil {
			if os.IsExist(err) {
				return "", ErrObjectIsAlreadyExists
			}
			if os.IsPermission(err) {
				return "", ErrObjectIsNotWritable
			}
			return "", err
		}
		defer fo.Close()
	}

	buf := make([]byte, 1024)
	for {
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}

		if _, err := fo.Write(buf[:n]); err != nil {
			return "", err
		}
	}

	return p, nil
}

func (f *File) UntarFile(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	tr := tar.NewReader(file)
	var targetPath string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath = filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				_ = outFile.Close()
				return err
			}
			if err := outFile.Close(); err != nil {
				_ = outFile.Close()
			}
		default:
		}
	}
	return nil
}

func (f *File) UntarGzFile(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var targetPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath = filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}

		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				_ = outFile.Close()
				return err
			}
			_ = outFile.Close()
		default:
		}
	}
	return nil
}

func (f *File) UnzipFile(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		fpath := filepath.Join(destDir, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		inFile, err := file.Open()
		if err != nil {
			return err
		}
		defer inFile.Close()

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		_, err = io.Copy(outFile, inFile)
		_ = outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil

}
