package main

import (
	"fmt"
	"log"
	"time"

	"github.com/codehard-labs/logy/core/client"
	"go.uber.org/zap"
)

func main() {
	client.AppName = "test-app"
	client.InstanceName = "test-instance"
	client.ServerAddress = "127.0.0.1:8878"
	c := client.NewClient("test-component")
	logger := c.DeafultZapLogger()

	ch := make(chan interface{}, 65537*2)
	c.SubscribeBroadcaster(ch)
	go func() {
		count := 0
		l := 0
		for m := range ch {
			count++
			msg := string(m.([]byte))
			l += len(msg)
			fmt.Print(msg)
			if count >= 65537*2 {
				log.Print("total bytes:", l)
			}
		}
	}()

	for i := 0; i < 65537; i++ {
		logger.Info("test", zap.Int("count", i))
	}
	//c.UnsubscribeBroadcaster(ch)
	for i := 0; i < 65537; i++ {
		logger.Info("test2", zap.Int("count2", i))
	}
	log.Println("all log done")
	time.Sleep(10 * time.Second)
}
