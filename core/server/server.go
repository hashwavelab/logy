package server

import (
	"github.com/hashwavelab/logy/core/db"
)

var (
	LNName = "LN"
)

type Server struct {
	dbClient db.DBClient
}

func NewServer() *Server {
	return &Server{}
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
