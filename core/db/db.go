package db

var (
	DBName           = "LOGY"
	RecordsTableName = "RECORDS"
)

type DBClient interface {
	Ping() error
	SaveLogs(string, [][]byte) error
	SaveSubmissionRecord(string, int32, bool, string, int) error
}
