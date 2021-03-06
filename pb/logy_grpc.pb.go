// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LogyClient is the client API for Logy service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogyClient interface {
	SubmitLogs(ctx context.Context, opts ...grpc.CallOption) (Logy_SubmitLogsClient, error)
	SubmitLogsWithoutStream(ctx context.Context, in *Logs, opts ...grpc.CallOption) (*EmptyResponse, error)
}

type logyClient struct {
	cc grpc.ClientConnInterface
}

func NewLogyClient(cc grpc.ClientConnInterface) LogyClient {
	return &logyClient{cc}
}

func (c *logyClient) SubmitLogs(ctx context.Context, opts ...grpc.CallOption) (Logy_SubmitLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Logy_ServiceDesc.Streams[0], "/logy.Logy/SubmitLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &logySubmitLogsClient{stream}
	return x, nil
}

type Logy_SubmitLogsClient interface {
	Send(*Logs) error
	CloseAndRecv() (*EmptyResponse, error)
	grpc.ClientStream
}

type logySubmitLogsClient struct {
	grpc.ClientStream
}

func (x *logySubmitLogsClient) Send(m *Logs) error {
	return x.ClientStream.SendMsg(m)
}

func (x *logySubmitLogsClient) CloseAndRecv() (*EmptyResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(EmptyResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *logyClient) SubmitLogsWithoutStream(ctx context.Context, in *Logs, opts ...grpc.CallOption) (*EmptyResponse, error) {
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, "/logy.Logy/SubmitLogsWithoutStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogyServer is the server API for Logy service.
// All implementations must embed UnimplementedLogyServer
// for forward compatibility
type LogyServer interface {
	SubmitLogs(Logy_SubmitLogsServer) error
	SubmitLogsWithoutStream(context.Context, *Logs) (*EmptyResponse, error)
	mustEmbedUnimplementedLogyServer()
}

// UnimplementedLogyServer must be embedded to have forward compatible implementations.
type UnimplementedLogyServer struct {
}

func (UnimplementedLogyServer) SubmitLogs(Logy_SubmitLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method SubmitLogs not implemented")
}
func (UnimplementedLogyServer) SubmitLogsWithoutStream(context.Context, *Logs) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitLogsWithoutStream not implemented")
}
func (UnimplementedLogyServer) mustEmbedUnimplementedLogyServer() {}

// UnsafeLogyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogyServer will
// result in compilation errors.
type UnsafeLogyServer interface {
	mustEmbedUnimplementedLogyServer()
}

func RegisterLogyServer(s grpc.ServiceRegistrar, srv LogyServer) {
	s.RegisterService(&Logy_ServiceDesc, srv)
}

func _Logy_SubmitLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LogyServer).SubmitLogs(&logySubmitLogsServer{stream})
}

type Logy_SubmitLogsServer interface {
	SendAndClose(*EmptyResponse) error
	Recv() (*Logs, error)
	grpc.ServerStream
}

type logySubmitLogsServer struct {
	grpc.ServerStream
}

func (x *logySubmitLogsServer) SendAndClose(m *EmptyResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *logySubmitLogsServer) Recv() (*Logs, error) {
	m := new(Logs)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Logy_SubmitLogsWithoutStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Logs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogyServer).SubmitLogsWithoutStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logy.Logy/SubmitLogsWithoutStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogyServer).SubmitLogsWithoutStream(ctx, req.(*Logs))
	}
	return interceptor(ctx, in, info, handler)
}

// Logy_ServiceDesc is the grpc.ServiceDesc for Logy service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Logy_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logy.Logy",
	HandlerType: (*LogyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitLogsWithoutStream",
			Handler:    _Logy_SubmitLogsWithoutStream_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubmitLogs",
			Handler:       _Logy_SubmitLogs_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "pb/logy.proto",
}
