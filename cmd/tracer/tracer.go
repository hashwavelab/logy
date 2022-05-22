package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/hashwavelab/logy/core/db"
	"github.com/hashwavelab/logy/core/tracer"
)

func main() {
	// s := gocron.NewScheduler(time.UTC)
	// s.Cron("*/10 * * * *").Do(func() {
	// 	endWithTolerences := time.Now().UnixNano()
	// 	end := endWithTolerences - 3*60*1000*1000*1000
	// 	start := end - 10*60*1000*1000*1000
	// 	msg := `{"templateName":"recipeTracing","initDomain":"rpd_main_i0_172.31.44.215","initMatch":{"msg":"evm recipe"},"tracingMatch":"utid","domains":["swirl_ripple_celo_150.109.148.233","swirl_ripple_cronos_150.109.148.233","swirl_ripple_polygon_3.83.159.18","swirl_seal_i0_172.31.44.215","swirl_ripple_gnosis_34.217.50.229","swirl_ripple_avalanchec_150.109.148.233","swirl_ripple_bsc_34.217.50.229","rpd_main_i0_172.31.44.215","swirl_ripple_harmony_3.83.159.18","swirl_ripple_aurora_150.109.148.233","swirl_ripple_moonbeam_150.109.148.233","swirl_ripple_fantom_34.217.50.229","swirl_ripple_emerald_150.109.148.233"],"tsRangeInit":{"ts":{"$gte":` + strconv.FormatInt(start, 10) + `,"$lte":` + strconv.FormatInt(end, 10) + `}},"tsRangeTolerant":{"ts":{"$gte":` + strconv.FormatInt(start, 10) + `,"$lte":` + strconv.FormatInt(endWithTolerences, 10) + `}}`
	// 	tracer := tracer.InitTracer(msg, db.GetMongoClient(db.DBName))
	// 	res := tracer.OperateTracing()
	// 	atlas := db.GetMongoClient(db.GetDotEnvVariable("MONGO_ATLAS_URI"))
	// 	atlas.SaveDocs("logy", "utid_trace", res)
	// })
	localMongoCli := db.GetMongoClient(db.MongoURI)
	remoteMongoClient := db.GetMongoClient(db.GetDotEnvVariable("MONGO_ATLAS_URI"))
	jsonBytes := getJsonBytes("./tracing_recipes/utid_trace.json")

	endWithTolerance := time.Now().UnixNano()
	end := endWithTolerance - 3*60*1000*1000*1000
	start := end - 10*60*1000*1000*1000
	tracer := tracer.InitTracer(jsonBytes, localMongoCli, start, end, endWithTolerance)
	res := tracer.ExecuteTracing()
	remoteMongoClient.SaveDocs("logy", "utid_trace", res)
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
