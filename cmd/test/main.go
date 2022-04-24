package main

import (
	"fmt"
	"time"

	"github.com/hashwavelab/logy/core/client"
	"go.uber.org/zap"
)

func main() {
	client.AppName = "testapp"
	client.InstanceName = "i1"
	client.ServerAddress = "localhost:8878"
	client.BundleSize = 50
	c := client.NewClient("testcomp")
	logger := c.DeafultZapLogger()
	time.Sleep(0 * time.Second)
	begin := time.Now()
	for i := 0; i < 100; i++ {
		logger.Info("hi", zap.Int("count", i))
	}
	for i := 0; i < 100; i++ {
		logger.Info("hey", zap.Int("count", i))
	}
	fmt.Println("time taken", time.Since(begin))
	time.Sleep(10 * time.Second)
}