package db

var (
	DBName           = "logy"
	RecordsTableName = "records"
)

type DBClient interface {
	Ping() error
	SaveLogs(string, [][]byte) error
	DeleteOldLogs(int64) error
	SaveSubmissionRecord(string, int32, bool, string, int) error
}
