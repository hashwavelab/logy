package main

import (
	"log"

	"github.com/hashwavelab/logy/core/server"
)

func main() {
	logyServer := server.NewServer().UseMongo()
	err := logyServer.PingDB()
	if err != nil {
		log.Fatal("Cannot connect to DB", err)
	}
	log.Println("connected to DB")
	server.InitGRPCServer(logyServer, ":8878")
	select {}
}
