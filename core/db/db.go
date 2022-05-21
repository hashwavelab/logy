package db

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

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

func GetDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
