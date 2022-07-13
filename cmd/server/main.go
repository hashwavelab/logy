package main

import (
	"log"

	"github.com/codehard-labs/logy/core/server"
)

func main() {
	//server.MaxAgeOfLogsInNanoSeconds = (6 * time.Hour).Nanoseconds()
	logyServer := server.NewServer().UseMongo()
	err := logyServer.PingDB()
	if err != nil {
		log.Fatal("Cannot connect to DB", err)
	}
	log.Println("connected to DB")
	server.InitGRPCServer(logyServer, ":8878")
	server.InitWebService(logyServer)
	select {}
}
