package ffm

import "time"

type FileInfo struct {
	Path     string    `json:"path"`
	FullPath string    `json:"full_path"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	ModTime  time.Time `json:"mod_time"`
}
