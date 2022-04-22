package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hashwavelab/logy/core/db"
)

func InitWebService(server *Server) {
	r := gin.Default()
	// r.GET("/", server.SimpleGet)
	r.GET("/collections", server.SimpleGetByName)

	r.Run() // listen and serve on 0.0.0.0:8080
}

func (s *Server) SimpleGet(c *gin.Context) {
	logs, err := s.dbClient.(*db.MongoDBClient).GetLogs("testapp_testcomp_i0_local", db.EmptyFilter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, logs)
}

func (s *Server) SimpleGetByName(c *gin.Context) {
	name := c.Query("name")
	log.Println("Received", name)
	// name := c.Param("name")
	logs, err := s.dbClient.(*db.MongoDBClient).GetLogs(name, db.EmptyFilter)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
