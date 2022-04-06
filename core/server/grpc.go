package server

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/hashwavelab/logy/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/peer"
)

type GRPCServer struct {
	server *Server
	pb.UnimplementedLogyServer
}

func (s *GRPCServer) SubmitLogs(stream pb.Logy_SubmitLogsServer) error {
	for {
		logs, err := stream.Recv()
		p, _ := peer.FromContext(stream.Context())
		ip, ok := getIPFromAddress(p.Addr.String())
		if !ok {
			return errors.New("cannot get ip")
		}
		if err == io.EOF {
			return stream.SendAndClose(&pb.EmptyResponse{})
		} else if err != nil {
			log.Println("stream is closed", err)
			return nil
		}
		s.server.saveLogsToDB(logs.App+"_"+logs.Component+"_"+logs.Instance+"_"+ip, logs.Logs, logs.SubmitType)
	}
}

func InitGRPCServer(server *Server, port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	encoding.RegisterCompressor(encoding.GetCompressor(gzip.Name))
	s := grpc.NewServer(grpc.MaxRecvMsgSize(50 * 1024 * 1024))
	pb.RegisterLogyServer(s, &GRPCServer{server: server})
	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func getIPFromAddress(addr string) (string, bool) {
	for i := len(addr) - 1; i > 0; i-- {
		if addr[i:i+1] == ":" {
			res := addr[:i]
			if res == "[::1]" || res == "127.0.0.1" || res == "localhost" {
				res = "LOCAL"
			}
			return res, true
		}
	}
	return "", false
}
