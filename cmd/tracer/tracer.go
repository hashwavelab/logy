package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/hashwavelab/logy/core/db"
	"github.com/hashwavelab/logy/core/server"
	"github.com/hashwavelab/logy/core/tracer"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	LocalMongoCli  *db.MongoDBClient
	RemoteMongoCli *db.MongoDBClient
)

func init() {
	LocalMongoCli = db.GetMongoClient(db.MongoURI)
	RemoteMongoCli = db.GetMongoClient(db.GetDotEnvVariable("MONGO_ATLAS_URI"))
}

func main() {
	c := cron.New()
	task0JSONByte := getJsonBytes("./tracing_recipes/utid_trace.json")
	c.AddFunc("@every 10m", func() { traceTask(task0JSONByte) })
	c.Start()
	select {}
}

func traceTask(jsonBytes []byte) {
	log.Println("tracing...")
	now := time.Now().UnixNano()
	endWithTolerance := now
	end := endWithTolerance - 3*60*1000*1000*1000
	start := end - 10*60*1000*1000*1000
	tracer := tracer.InitTracer(jsonBytes, LocalMongoCli, start, end, endWithTolerance)
	res := tracer.ExecuteTracing()
	RemoteMongoCli.SaveDocs("logy", "utid_trace", res)
	RemoteMongoCli.DeleteDocs("logy", "utid_trace", bson.M{"ts": bson.M{"$lte": now - server.MaxAgeOfLogsInNanoSeconds}})
	log.Println("tracing done")
}

func getJsonBytes(path string) []byte {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	jsonBytes, _ := ioutil.ReadAll(jsonFile)
	return jsonBytes
}

// func saveTestLogs() {
// 	client.AppName = "testapp2"
// 	client.InstanceName = "i0"
// 	client.ServerAddress = "localhost:8878"
// 	client.BundleSize = 50
// 	c := client.NewClient("testcomp")
// 	logger := c.DeafultZapLogger()
// 	time.Sleep(0 * time.Second)
// 	begin := time.Now()
// 	for i := 0; i < 100; i++ {
// 		logger.Error("hi", zap.Int("count", i))
// 	}
// 	for i := 0; i < 100; i++ {
// 		logger.Warn("hey", zap.Int("count", i))
// 	}
// 	fmt.Println("time taken", time.Since(begin))
// 	time.Sleep(10 * time.Second)
// }
