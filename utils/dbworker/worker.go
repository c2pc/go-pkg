package dbworker

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/dbworker/internal/worker"
	"gorm.io/gorm"
)

type Worker interface {
	Start(ctx context.Context) error
}

type Config = worker.Config

func NewWorker(cfg Config, db *gorm.DB) Worker {
	return worker.NewWorkerImpl(cfg, db)
}
