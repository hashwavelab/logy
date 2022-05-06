package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/hashwavelab/logy/core/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitWebService(server *Server) {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.GET("/list", server.GetAllCollectionNames)
	r.GET("/logs", server.GetLogs)
	r.Run("localhost:5004") // listen and serve on 5004
}

// func (s *Server) GetRecords(c *gin.Context, limit int64) {
// 	logs, err := s.dbClient.(*db.MongoDBClient).GetLogs("records", db.EmptyFilter, options.Find().SetLimit(limit), options.Find().SetSort(bson.D{{Key: "ts", Value: -1}}))
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, logs)
// }

func (s *Server) GetAllCollectionNames(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	collectionName, err := s.dbClient.(*db.MongoDBClient).GetAllCollectionNames()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, collectionName)
}

func (s *Server) GetLogs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	name := c.Query("name")
	jsonQuery := c.Query("query")
	jsonSort := c.Query("sort")
	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		log.Println("ParseInt Error: ", err)
		limit = 500
	}
	if limit == 0 || limit > 10000 {
		limit = 500
	}
	//
	optionsArray := make([]*options.FindOptions, 0)
	query := &bson.M{}
	if jsonQuery != "" {
		bson.UnmarshalExtJSON([]byte(jsonQuery), true, query)
	}
	sort := &bson.M{}
	if jsonSort != "" {
		bson.UnmarshalExtJSON([]byte(jsonSort), true, sort)
		optionsArray = append(optionsArray, options.Find().SetSort(sort))
	}
	optionsArray = append(optionsArray, options.Find().SetLimit(limit))
	//
	logs, err := s.dbClient.(*db.MongoDBClient).GetLogs(name, query, optionsArray...)
	if err != nil {
		log.Println("error: ", err, query, optionsArray)
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
