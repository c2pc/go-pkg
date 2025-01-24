package worker

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"os"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/dbworker/internal/workererrors"

	logger2 "github.com/c2pc/go-pkg/v2/utils/logger"
	"gorm.io/gorm"
)

const FileTimeFormat = "20060102_150405"

type WorkerImpl struct {
	db  *gorm.DB
	cfg Config
}

func NewWorkerImpl(cfg Config, db *gorm.DB) *WorkerImpl {
	return &WorkerImpl{
		db:  db,
		cfg: cfg,
	}
}

func (w *WorkerImpl) Start(ctx context.Context) error {
	if err := os.MkdirAll(w.cfg.ArchivePath, 0755); err != nil {
		return workererrors.NewOpError("[DB WORKER] create archive directory", w.cfg.ArchivePath, "", err)
	}

	if w.cfg.CheckInterval <= 0 {
		w.cfg.CheckInterval = 14 * 24 * time.Hour
	}

	ticker := time.NewTicker(w.cfg.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger2.Info("[DB WORKER] shutting down")
			return nil
		case <-ticker.C:
			if err := w.processExpiredRecords(ctx); err != nil {
				return workererrors.NewOpError("[DB WORKER] error in processExpiredRecords", "", w.cfg.TableName, err)
			}
			if err := w.processOversizedTable(ctx); err != nil {
				return workererrors.NewOpError("[DB WORKER] error in processOversizedTable", "", w.cfg.TableName, err)
			}

		}
	}
}

func (w *WorkerImpl) processExpiredRecords(ctx context.Context) error {
	cutoffTime := time.Now().Add(-w.cfg.TimeThreshold)

	var records []map[string]interface{}
	if err := w.db.WithContext(ctx).
		Table(w.cfg.TableName).
		Where(fmt.Sprintf("%s < ?", w.cfg.TimeFieldName), cutoffTime).
		Find(&records).Error; err != nil {
		return workererrors.NewOpError("[DB WORKER] query expired records", "", w.cfg.TableName, err)
	}

	if len(records) == 0 {
		return nil
	}

	fileName := fmt.Sprintf(
		"%s_%s.json.gz",
		w.cfg.ArchiveFilePrefix,
		time.Now().Format(FileTimeFormat),
	)
	archivePath := filepath.Join(w.cfg.ArchivePath, fileName)

	if err := writeRecordsToGzipFile(archivePath, records); err != nil {
		return workererrors.NewOpError("[DB WORKER] write gzip file", archivePath, w.cfg.TableName, err)
	}

	if err := w.deleteRecords(ctx, records); err != nil {
		return workererrors.NewOpError("[DB WORKER] delete expired records", "", w.cfg.TableName, err)
	}

	logger2.Info(fmt.Sprintf("[DB WORKER] archived and deleted %d expired records => %s\n", len(records), archivePath))
	return nil
}

func (w *WorkerImpl) processOversizedTable(ctx context.Context) error {
	var totalCount int64
	if err := w.db.WithContext(ctx).
		Table(w.cfg.TableName).
		Count(&totalCount).Error; err != nil {
		return workererrors.NewOpError("[DB WORKER] count rows", "", w.cfg.TableName, err)
	}

	if totalCount <= int64(w.cfg.RowCountThreshold) {
		return nil
	}

	var records []map[string]interface{}
	if err := w.db.WithContext(ctx).
		Table(w.cfg.TableName).
		Order(fmt.Sprintf("%s ASC", w.cfg.TimeFieldName)).
		Limit(w.cfg.ArchiveBatchSize).
		Find(&records).Error; err != nil {
		return workererrors.NewOpError("[DB WORKER] query oldest records", "", w.cfg.TableName, err)
	}

	if len(records) == 0 {
		return nil
	}

	lastArchive, err := findLastArchive(w.cfg.ArchivePath, w.cfg.ArchiveFilePrefix)
	if err != nil {
		return workererrors.NewOpError("[DB WORKER] find last archive", w.cfg.ArchivePath, w.cfg.TableName, err)
	}

	if lastArchive == "" {
		fileName := fmt.Sprintf(
			"%s_%s.json.gz",
			w.cfg.ArchiveFilePrefix,
			time.Now().Format(FileTimeFormat),
		)
		lastArchive = filepath.Join(w.cfg.ArchivePath, fileName)
		if err := writeRecordsToGzipFile(lastArchive, records); err != nil {
			return workererrors.NewOpError("[DB WORKER] write gzip file (no existing archive)", lastArchive, w.cfg.TableName, err)
		}
	} else {
		existing, err := readRecordsFromGzipFile(lastArchive)
		if err != nil {
			return workererrors.NewOpError("[DB WORKER] read existing archive", lastArchive, w.cfg.TableName, err)
		}
		existing = append(existing, records...)

		if err := writeRecordsToGzipFile(lastArchive, existing); err != nil {
			return workererrors.NewOpError("[DB WORKER] update existing archive", lastArchive, w.cfg.TableName, err)
		}
	}

	if err := w.deleteRecords(ctx, records); err != nil {
		return workererrors.NewOpError("[DB WORKER] delete old records", "", w.cfg.TableName, err)
	}

	logger2.Info(fmt.Sprintf("[DB WORKER] archived and deleted %d oldest records => %s\n", len(records), lastArchive))
	return nil
}

func (w *WorkerImpl) deleteRecords(ctx context.Context, records []map[string]interface{}) error {
	var ids []interface{}
	for _, r := range records {
		if val, ok := r["id"]; ok {
			ids = append(ids, val)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	return w.db.WithContext(ctx).
		Table(w.cfg.TableName).
		Where("id IN ?", ids).
		Delete(nil).
		Error
}

func writeRecordsToGzipFile(filePath string, records []map[string]interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger2.Warning(err.Error())
		}
	}()

	gw := gzip.NewWriter(f)
	defer func() {
		if err := gw.Close(); err != nil {
			logger2.Warning(err.Error())
		}
	}()

	enc := json.NewEncoder(gw)
	enc.SetIndent("", "  ")
	if err := enc.Encode(records); err != nil {
		return err
	}
	return nil
}

func readRecordsFromGzipFile(filePath string) ([]map[string]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger2.Warning(err.Error())
		}
	}()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := gr.Close(); err != nil {
			logger2.Warning(err.Error())
		}
	}()

	var records []map[string]interface{}
	dec := json.NewDecoder(gr)
	if err := dec.Decode(&records); err != nil && err != io.EOF {
		return nil, err
	}
	return records, nil
}

func findLastArchive(dir, prefix string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var lastFile string
	var lastModTime time.Time

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()

		if prefix != "" && len(name) >= len(prefix) {
			if name[:len(prefix)] != prefix {
				continue
			}
		}
		if filepath.Ext(name) != ".gz" {
			continue
		}

		fullPath := filepath.Join(dir, name)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if info.ModTime().After(lastModTime) {
			lastFile = fullPath
			lastModTime = info.ModTime()
		}
	}

	return lastFile, nil
}
