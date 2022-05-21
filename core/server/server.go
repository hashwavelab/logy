package server

import (
	"time"

	"github.com/hashwavelab/logy/core/db"
)

var (
	MaxAgeOfLogsCheckInterval = 1 * time.Hour
	MaxAgeOfLogsInNanoSeconds = (2 * 24 * time.Hour).Nanoseconds()
)

type Server struct {
	dbClient db.DBClient
}

func NewServer() *Server {
	s := &Server{}

	go func() {
		time.Sleep(3 * time.Second)
		s.deleteOldLogs()
		for range time.NewTicker(MaxAgeOfLogsCheckInterval).C {
			s.deleteOldLogs()
		}
	}()

	return s
}

func (s *Server) UseMongo() *Server {
	s.dbClient = db.GetMongoClient()
	return s
}

func (s *Server) PingDB() error {
	return s.dbClient.Ping()
}

func (s *Server) saveLogsToDB(collection string, rawLogs [][]byte, submitType int32) error {
	err := s.dbClient.SaveLogs(collection, rawLogs)
	if err != nil {
		return s.dbClient.SaveSubmissionRecord(collection, submitType, false, err.Error(), len(rawLogs))
	}
	return s.dbClient.SaveSubmissionRecord(collection, submitType, true, "", len(rawLogs))
}

func (s *Server) deleteOldLogs() {
	s.dbClient.DeleteOldLogs(time.Now().UnixNano() - MaxAgeOfLogsInNanoSeconds)
}
