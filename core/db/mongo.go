package db

import (
	"context"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoURI          string        = "mongodb://127.0.0.1:27017"
	ConnectionTimeout time.Duration = 2 * time.Second
	MongoQueryTimeout time.Duration = 30 * time.Second
	EmptyFilter                     = &bson.M{}
)

type MongoDBClient struct {
	sync.RWMutex
	client     *mongo.Client
	colls      map[string]bool
	disconnect func()
}

func GetMongoClient() *MongoDBClient {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(MongoURI))
	if err != nil || client == nil {
		log.Fatal("mongo get client error", err)
		return nil
	}
	c := &MongoDBClient{
		client: client,
		colls:  make(map[string]bool),
		disconnect: func() {
			err := client.Disconnect(context.TODO())
			if err != nil {
				log.Println("mongo disconnect error")
			}
		},
	}
	colls, err := c.GetAllCollectionNames()
	if err != nil {
		log.Fatal(err)
	}
	for _, collName := range colls {
		c.addCollection(collName)
	}
	return c
}

func (c *MongoDBClient) addCollection(cn string) {
	c.Lock()
	defer c.Unlock()
	c.colls[cn] = true
}

func (c *MongoDBClient) hasCollection(cn string) bool {
	c.RLock()
	defer c.RUnlock()
	_, has := c.colls[cn]
	return has
}

func (c *MongoDBClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	return c.client.Ping(ctx, nil)
}

func (c *MongoDBClient) getCollection(collection string) *mongo.Collection {
	if !c.hasCollection(collection) {
		ctx, cancel := context.WithTimeout(context.Background(), MongoQueryTimeout)
		defer cancel()
		err := c.client.Database(DBName).CreateCollection(ctx, collection)
		if err != nil {
			c.addCollection(collection)
		}
		coll := c.client.Database(DBName).Collection(collection)
		mod := mongo.IndexModel{
			Keys: bson.M{
				"ts": -1,
			}, Options: nil,
		}
		ctx1, cancel1 := context.WithTimeout(context.Background(), MongoQueryTimeout)
		defer cancel1()
		index, err1 := coll.Indexes().CreateOne(ctx1, mod)
		log.Println("new collection created:", collection, err, index, err1)
	}
	return c.client.Database(DBName).Collection(collection)
}

func (c *MongoDBClient) DeleteOldLogs(ts int64) error {
	colls, err := c.client.Database(DBName).ListCollectionNames(context.TODO(), EmptyFilter)
	if err != nil {
		log.Println("Mongo failed to list all collections")
		return err
	}
	for _, collName := range colls {
		coll := c.client.Database(DBName).Collection(collName)
		result, err := coll.DeleteMany(context.TODO(), bson.M{"ts": bson.M{
			"$lte": ts,
		}})
		if err != nil {
			log.Println("DeleteMany failed for", collName, err)
			continue
		}
		log.Println("DeleteMany success, old logs deleted for", collName, result.DeletedCount)
	}
	return nil
}

func (c *MongoDBClient) GetLogs(collection string, filter interface{}, opts ...*options.FindOptions) ([]bson.M, error) {
	coll := c.getCollection(collection)
	var docs []bson.M
	ctx, cancel := context.WithTimeout(context.Background(), MongoQueryTimeout)
	defer cancel()
	r, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	r.All(context.TODO(), &docs)
	return docs, nil
}

func (c *MongoDBClient) GetAllCollectionNames() ([]string, error) {
	colls, err := c.client.Database(DBName).ListCollectionNames(context.TODO(), EmptyFilter)
	if err != nil {
		log.Println("Mongo failed to list all collections")
		return nil, err
	}
	return colls, nil
}

func (c *MongoDBClient) SaveLogs(collection string, rawLogs [][]byte) error {
	coll := c.getCollection(collection)
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
