package client

import (
	"context"
	"time"

	"github.com/hashwavelab/logy/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	GetStreanTimeout     = 1 * time.Second
	MinReconnectInterval = 1 * time.Minute
)

func (c *Client) getStream() {
	if time.Since(c.lastConTime) < MinReconnectInterval {
		return
	}
	defer func() { c.lastConTime = time.Now() }()
	con, err := grpc.Dial(ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		//log.Println("grpc dial error", err)
		c.stream = nil
		return
	}
	client := pb.NewLogyClient(con)
	connected := false
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if <-time.NewTimer(GetStreanTimeout).C; !connected {
			cancel()
		}
	}()
	stream, err := client.SubmitLogs(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		//log.Println("grpc submit error", err)
		c.stream = nil
		return
	}
	connected = true
	c.stream = stream
}
