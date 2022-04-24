package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hashwavelab/logy/core/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitWebService(server *Server) {
	r := gin.Default()
	r.GET("/list", server.GetAllCollectionNames)
	r.GET("/collections", server.SimpleGetByName)
	r.Run("localhost:5004") // listen and serve on 5004
}

func (s *Server) GetRecords(c *gin.Context, limit int64) {
	logs, err := s.dbClient.(*db.MongoDBClient).GetLogs("records", db.EmptyFilter, options.Find().SetLimit(limit), options.Find().SetSort(bson.D{{Key: "ts", Value: -1}}))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

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

func (s *Server) SimpleGetByName(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	names := c.QueryArray("name[]")
	level := c.Query("level")
	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil {
		log.Println("ParseInt Error: ", err)
	}
	start, err := strconv.ParseInt(c.Query("start")+"000000", 10, 64)
	if err != nil {
		log.Println("ParseInt Error: ", err)
	}
	end, err := strconv.ParseInt(c.Query("end")+"000000", 10, 64)
	if err != nil {
		log.Println("ParseInt Error: ", err)
	}
	log.Println("Received", names, level, limit, start, end)

	if names[0] == "records" {
		s.GetRecords(c, limit)
		return
	}
	var pipeline *primitive.M
	if level == "warn" {
		pipeline = &bson.M{
			"ts": bson.M{
				"$gte": start,
				"$lte": end,
			},
			"$or": []interface{}{
				bson.M{"level": "warn"},
				bson.M{"level": "error"},
			},
		}
	} else if level == "error" {
		pipeline = &bson.M{
			"level": "error",
			"ts": bson.M{
				"$gte": start,
				"$lte": end,
			},
		}
	} else {
		pipeline = &bson.M{
			"ts": bson.M{
				"$gte": start,
				"$lte": end,
			},
		}
	}
	var returnArr []primitive.M
	for i := 0; i < len(names); i++ {
		logs, err := s.dbClient.(*db.MongoDBClient).GetLogs(names[i], pipeline, options.Find().SetLimit(limit), options.Find().SetSort(bson.D{{Key: "ts", Value: -1}}))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			continue
		}
		returnArr = append(returnArr, logs...)
	}
	c.JSON(http.StatusOK, returnArr)
}
