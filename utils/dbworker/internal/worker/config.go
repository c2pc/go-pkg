package worker

type Config struct {
	TableName        string
	TimeFieldName    string
	ArchiveBatchSize int
	ArchivePath      string
	UnzipNames       []string
	Frequency        string
}
