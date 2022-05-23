package tracer

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hashwavelab/logy/core/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Tracer struct {
	dbClient        *db.MongoDBClient
	initDomain      string // e.g. "rpd_main_i0_172.31.44.215"
	initMatch       bson.M // e.g. {"msg":"evm recipe", "level": "info"}
	tsRangeInit     bson.M // e.g. {"ts":{"$gte":123456000 "$lte":234567000}}
	tsRangeTolerant bson.M // e.g. {"ts":{"$gte":123456000 "$lte":234567000 + 5*60*1000}} come with tolerance
	tracingMatch    string // e.g. "utid"
	originDBMapLock sync.RWMutex
	originDBMap     map[string][]bson.M // e.g. "rpd_main_i0_172.31.44.215" -> []string
	tracedDBMap     map[string][]bson.M // e.g. tracingMatch -> []string
}

type UtidTrace struct {
	Timestamp         time.Time `bson:"timestamp" json:"timestamp"`
	Utid              string    `bson:"utid" json:"utid"`
	AssetFrom         string    `bson:"assetFrom" json:"assetFrom"`
	AssetTo           string    `bson:"assetTo" json:"assetTo"`
	ChainName         string    `bson:"chainName" json:"chainName"`
	InitiatiedByChain bool      `bson:"initiatiedByChain" json:"initiatiedByChain"`
	LastMsg           string    `bson:"lastMsg" json:"lastMsg"`
	Steps             int       `bson:"steps" json:"steps"`
	TerminatedLevel   string    `bson:"terminatedLevel" json:"terminatedLevel"`
}

// e.g. "{"templateName":"recipeTracing","initDomain":"rpd_main_i0_172.31.44.215","initMatch":{"msg":"evm recipe"},"tracingMatch":"utid","domains":["swirl_seal_i0_172.31.44.215","swirl_ripple_aurora_150.109.148.233"],"tsRange":{"ts":"18:00-19:00"}}"
func InitTracer(initJSONbytes []byte, dbClient *db.MongoDBClient, start, end, endWithTolerance int64) *Tracer {
	t := &Tracer{
		dbClient:    dbClient,
		originDBMap: make(map[string][]bson.M),
		tracedDBMap: make(map[string][]bson.M),
	}
	// decode initJSONbytes to bson.M
	info := bson.M{}
	bson.UnmarshalExtJSON(initJSONbytes, true, info)
	// construct the tracer
	for key, value := range info {
		switch key {
		case "initDomain":
			t.initDomain = value.(string)
			t.originDBMap[t.initDomain] = make([]bson.M, 0)
		case "initMatch":
			t.initMatch = value.(bson.M)
		case "tsRangeInit":
			tsRangeInit := value.(bson.M)
			tsRangeInit["ts"].(bson.M)["$gte"] = start
			tsRangeInit["ts"].(bson.M)["$lte"] = end
			t.tsRangeInit = tsRangeInit
		case "tsRangeTolerant":
			tsRangeTolerant := value.(bson.M)
			tsRangeTolerant["ts"].(bson.M)["$gte"] = start
			tsRangeTolerant["ts"].(bson.M)["$lte"] = endWithTolerance
			t.tsRangeTolerant = tsRangeTolerant
		case "tracingMatch":
			t.tracingMatch = value.(string)
		case "domains":
			for _, domain := range value.(bson.A) {
				t.originDBMap[domain.(string)] = make([]bson.M, 0)
			}
		}
	}
	return t
}

func (t *Tracer) ExecuteTracing() []interface{} {
	var initMatch bson.M = t.initMatch
	var tracingMatch string = t.tracingMatch
	// Step 1:
	// query all db data from all domains and save into originDBMap
	wg := &sync.WaitGroup{}
	for key := range t.originDBMap {
		wg.Add(1)
		go t.step1(key, wg)
	}
	wg.Wait()
	log.Println("Step1 Finished")
	// Step 2:
	// trace out logs from initDomain which satisfy init match and save into tracedDBMap from tracingMatch to array of logs
	for domainName, domain := range t.originDBMap {
		switch domainName {
		case t.initDomain:
			for _, log := range domain {
				if _, ok := log[tracingMatch]; ok {
					matchCount := 0
					for key, value := range initMatch {
						if _, ok := log[key]; ok {
							if log[key] == value {
								matchCount++
							} else {
								break
							}
						} else {
							break
						}
					}
					if matchCount == len(initMatch) {
						// ingnore logs with new tracingMatch after the init ending timestamp range
						if log["ts"].(int64) <= t.tsRangeInit["ts"].(bson.M)["$lte"].(int64) {
							t.tracedDBMap[log[tracingMatch].(string)] = make([]bson.M, 0)
							t.tracedDBMap[log[tracingMatch].(string)] = append(t.tracedDBMap[log[tracingMatch].(string)], log)
						}
					} else {
						if _, ok := t.tracedDBMap[log[tracingMatch].(string)]; ok {
							t.tracedDBMap[log[tracingMatch].(string)] = append(t.tracedDBMap[log[tracingMatch].(string)], log)
						}
					}
				}
			}
		default:
			continue
		}
	}
	log.Println("Step2 Finished")
	// Step 3:
	// loop data from all domains to find out logs with same tracingMatch and save into the tracedDBMap
	for domainName, domain := range t.originDBMap {
		switch domainName {
		case t.initDomain:
			continue
		default:
			for _, log := range domain {
				// if the log has a key that matches the tracingMatch
				if _, ok := log[tracingMatch]; ok {
					// and the tracedDBMap has a key that match the tracingMatch, append into array
					if _, ok := t.tracedDBMap[log[tracingMatch].(string)]; ok {
						t.tracedDBMap[log[tracingMatch].(string)] = append(t.tracedDBMap[log[tracingMatch].(string)], log)
					}
				}
			}
		}
	}
	log.Println("Step3 Finished")
	// Step 4:
	// sort logs in each tracedDBMap by timestamp and export
	uts := make([]interface{}, len(t.tracedDBMap))
	i := 0
	for key, domain := range t.tracedDBMap {
		sort.Slice(domain, func(i, j int) bool {
			return domain[i]["ts"].(int64) < domain[j]["ts"].(int64)
		})
		// fmt.Println(key)
		// fmt.Println(domain)
		ut := UtidTrace{
			Timestamp:         time.UnixMilli(domain[0]["ts"].(int64) / 1000).UTC(),
			Utid:              key,
			AssetFrom:         domain[0]["assetFrom"].(string),
			AssetTo:           domain[0]["assetTo"].(string),
			ChainName:         domain[0]["chainName"].(string),
			InitiatiedByChain: domain[0]["chainName"].(string) == domain[0]["initiativeSource"].(string),
			LastMsg:           domain[len(domain)-1]["msg"].(string),
			Steps:             len(domain),
			TerminatedLevel:   domain[len(domain)-1]["level"].(string),
		}
		uts[i] = ut
		i++
	}
	log.Println("Step4 Finished")
	return uts
}

func (t *Tracer) step1(key string, wg *sync.WaitGroup) {
	defer wg.Done()
	// fmt.Println(key, t.tsRangeTolerant)
	logs, err := t.dbClient.GetLogs(key, t.tsRangeTolerant, options.Find().SetSort(&bson.M{"ts": 1}))
	if err != nil {
		log.Println("error: ", err)
	}
	// fmt.Println(key, logs)
	t.originDBMapLock.Lock()
	defer t.originDBMapLock.Unlock()
	t.originDBMap[key] = logs
}

func (t *Tracer) StartTracingPeriodically(timeInterval uint64) {
	s := gocron.NewScheduler(time.UTC)
	s.Cron("*/10 * * * *").Do(func() {

	})
}
