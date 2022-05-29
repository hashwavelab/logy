package client

import (
	"log"
	"time"

	"github.com/hashwavelab/logy/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	SubmissionTimeout = 10 * time.Second
)

func getGRPCClient() pb.LogyClient {
	conn, err := grpc.Dial(ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("grpc dial signa failed:", err)
	}
	return pb.NewLogyClient(conn)
}
