package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoURI          string        = "mongodb://127.0.0.1:27017"
	ConnectionTimeout time.Duration = 2 * time.Second
)

type MongoDBClient struct {
	client     *mongo.Client
	disconnect func()
}

func GetMongoClient() DBClient {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(MongoURI))
	if err != nil || client == nil {
		log.Fatal("mongo get client error", err)
		return nil
	}
	return &MongoDBClient{
		client: client,
		disconnect: func() {
			err := client.Disconnect(context.TODO())
			if err != nil {
				log.Println("mongo disconnect error")
			}
		},
	}
}

func (c *MongoDBClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	return c.client.Ping(ctx, nil)
}

func (c *MongoDBClient) SaveLogs(collection string, rawLogs [][]byte) error {
	coll := c.client.Database(DBName).Collection(collection)
	docs := make([]interface{}, len(rawLogs))
	for i := 0; i < len(rawLogs); i++ {
		var bdoc interface{}
		err := bson.UnmarshalExtJSON(rawLogs[i], true, &bdoc)
		if err != nil {
			log.Println("UnmarshalExtJSON failed", string(rawLogs[i]))
			continue
		}
		docs[i] = bdoc
	}
	start := time.Now()
	result, err := coll.InsertMany(context.TODO(), docs)
	if err != nil {
		log.Println("InsertMany failed", err)
		return err
	}
	log.Println("InsertMany success, inserted", len(result.InsertedIDs), "time taken:", time.Since(start))
	return nil
}

// DeleteOldLogs checks all collections and delete logs older than ts.
func (c *MongoDBClient) DeleteOldLogs(ts int64) error {
	colls, err := c.client.Database(DBName).ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		log.Println("Mongo failed to list all collections")
		return err
	}
	for _, collName := range colls {
		coll := c.client.Database(DBName).Collection(collName)
		_, err := coll.DeleteMany(context.TODO(), bson.M{"ts": bson.M{
			"$lte": ts,
		}})
		if err != nil {
			log.Println("DeleteMany failed for", collName, err)
			continue
		}
		log.Println("DeleteMany success, old logs deleted for", collName)
	}
	return nil
}

type MongoRecord struct {
	CollName   string `bson:"coll_name" json:"coll_name"`
	TS         int64  `bson:"ts" json:"ts"`
	SubmitType int32  `bson:"submit_type" json:"submit_type"`
	Success    bool   `bson:"success" json:"success"`
	ErrMsg     string `bson:"err_msg" json:"err_msg"`
	Count      int    `bson:"count" json:"count"`
}

func (c *MongoDBClient) SaveSubmissionRecord(collName string, submitType int32, success bool, errMsg string, count int) error {
	coll := c.client.Database(DBName).Collection(RecordsTableName)
	record := MongoRecord{
		CollName:   collName,
		TS:         time.Now().UnixNano(),
		SubmitType: submitType,
		Success:    success,
		ErrMsg:     errMsg,
		Count:      count,
	}
	_, err := coll.InsertOne(context.TODO(), record)
	if err != nil {
		log.Println("InsertMany failed", err)
		return err
	}
	return nil
}
