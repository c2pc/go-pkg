package worker

import "time"

type Config struct {
	TableName         string
	TimeFieldName     string
	TimeThreshold     time.Duration
	RowCountThreshold int
	ArchiveBatchSize  int
	ArchivePath       string
	ArchiveFilePrefix string
	CheckInterval     time.Duration
}
