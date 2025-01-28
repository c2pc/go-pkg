package worker

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/dbworker/internal/workererrors"

	logger2 "github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/robfig/cron/v3"
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

	c := cron.New(
		cron.WithSeconds(),
	)

	_, err := c.AddFunc(w.cfg.Frequency, func() {
		if err := w.processOversizedTable(ctx); err != nil {
			logger2.Warning(fmt.Sprintf("[DB WORKER] failed to processOversizedTable: %v", err))
		}
	})
	if err != nil {
		return fmt.Errorf("[DB WORKER] failed to schedule cron job: %w", err)
	}

	c.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-sigCh:
	}

	c.Stop()
	logger2.Info("[DB WORKER] shutting down cron")

	return nil
}

func (w *WorkerImpl) processOversizedTable(ctx context.Context) error {
	var totalCount int64
	if err := w.db.WithContext(ctx).
		Table(w.cfg.TableName).
		Count(&totalCount).Error; err != nil {
		return workererrors.NewOpError("[DB WORKER] count rows", "", w.cfg.TableName, err)
	}

	if totalCount <= int64(w.cfg.ArchiveBatchSize) {
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

	w.unzipFields(records)

	fileName := fmt.Sprintf(
		"%s_%s.json.gz",
		w.cfg.TableName,
		time.Now().Format(FileTimeFormat),
	)

	filePath := filepath.Join(w.cfg.ArchivePath, fileName)

	if err := writeRecordsToGzipFile(filePath, records); err != nil {
		return workererrors.NewOpError("[DB WORKER] write gzip file (no existing archive)", filePath, w.cfg.TableName, err)
	}

	if err := w.deleteRecords(ctx, records); err != nil {
		return workererrors.NewOpError("[DB WORKER] delete old records", "", w.cfg.TableName, err)
	}

	logger2.Info(fmt.Sprintf("[DB WORKER] archived and deleted %d oldest records => %s\n", len(records), fileName))
	return nil
}

func (w *WorkerImpl) unzipFields(records []map[string]interface{}) {
	for i := range records {
		for _, fieldName := range w.cfg.UnzipNames {
			if val, ok := records[i][fieldName]; ok && val != nil {
				switch raw := val.(type) {
				case []byte:
					decomp, err := decompressGZIP(raw)
					if err != nil {
						logger2.Warning(fmt.Sprintf(
							"failed decompress field %q, id=%v: %v",
							fieldName, records[i]["id"], err,
						))
						continue
					}
					records[i][fieldName] = decomp

				case string:
					decomp, err := decompressGZIP([]byte(raw))
					if err != nil {
						logger2.Warning(fmt.Sprintf(
							"failed to decompress field %q (string), id=%v: %v",
							fieldName, records[i]["id"], err,
						))
						continue
					}
					records[i][fieldName] = decomp
				}
			}
		}
	}
}

func decompressGZIP(data []byte) (string, error) {
	b := bytes.NewReader(data)
	gr, err := gzip.NewReader(b)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	unzipped, err := io.ReadAll(gr)
	if err != nil {
		return "", err
	}
	return string(unzipped), nil
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
