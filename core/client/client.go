package client

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/hashwavelab/logy/pb"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	AppName        = ""
	InstanceName   = ""
	ServerAddress  = ""
	BundleInterval = 3 * time.Minute
	BundleSize     = 65536
)

type Client struct {
	grpcClient  pb.LogyClient
	compName    string
	ch          chan []byte
	cache       [][]byte
	localLogger io.Writer
	//
	cacheCount int
	cacheTime  time.Time
}

func NewClient(compName string) *Client {
	if AppName == "" {
		log.Fatal("AppName is required to init new logy client")
	} else if InstanceName == "" {
		log.Fatal("InstanceName is required to init new logy client")
	} else if ServerAddress == "" {
		log.Fatal("ServerAddress is required to init new logy client")
	}

	c := &Client{
		grpcClient: getGRPCClient(),
		compName:   compName,
		ch:         make(chan []byte, 16*BundleSize),
		cache:      make([][]byte, 0, 2*BundleSize),
		localLogger: &lumberjack.Logger{
			Filename:   "logs/local.log",
			MaxSize:    10, // megabytes
			MaxBackups: 10,
			MaxAge:     1, // days
		},
		cacheTime: time.Now(),
	}

	go func() {
		for log := range c.ch {
			// reserve bytes of length 1 to be used as command
			if len(log) > 1 {
				c.cacheCount++
				c.cache = append(c.cache, log)
				if c.cacheCount >= BundleSize || time.Since(c.cacheTime) > BundleInterval {
					c.send(0)
					continue
				}
			} else if len(log) == 1 {
				// p: periodic ping
				if log[0] == 'p' && time.Since(c.cacheTime) > BundleInterval {
					c.send(0)
					continue
				}
				// e: emergent error
				if log[0] == 'e' {
					c.send(1)
					continue
				}
			}
		}
	}()

	go func() {
		for range time.NewTicker(BundleInterval).C {
			c.ch <- []byte{'p'}
		}
	}()

	return c
}

// Implement io.writer; copy & cache the input.
func (c *Client) Write(d []byte) (n int, err error) {
	length := len(d)
	data := make([]byte, length)
	copy(data, d)
	c.ch <- data
	return length, nil
}

func (c *Client) SendImmediate() {
	c.ch <- []byte{'e'}
}

func (c *Client) DeafultZapLogger() *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.EpochNanosTimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(c),
		zap.DebugLevel,
	)
	return zap.New(core)
}

func (c *Client) send(submitType int32) {
	if len(c.cache) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), SubmissionTimeout)
	defer cancel()
	_, err := c.grpcClient.SubmitLogsWithoutStream(ctx, &pb.Logs{
		App:        AppName,
		Component:  c.compName,
		Instance:   InstanceName,
		SubmitType: submitType,
		Logs:       c.cache,
	})
	if err != nil {
		// Important: this force local log can takn 10X longer than using it directly with zap.
		// So aviod local logging as much as possible.
		for _, log := range c.cache {
			c.localLogger.Write(log)
		}
		return
	}
	// clear cache
	c.cache = make([][]byte, 0, 2*BundleSize)
	c.cacheCount = 0
	c.cacheTime = time.Now()
}
