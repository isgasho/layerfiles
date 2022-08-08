// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package filewatcher_model

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

// FuseFSClient is the client API for FuseFS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FuseFSClient interface {
	NotifyAccess(ctx context.Context, in *NotifyAccessReq, opts ...grpc.CallOption) (*NotifyAccessResp, error)
	NotifyRead(ctx context.Context, in *NotifyReadReq, opts ...grpc.CallOption) (*NotifyReadResp, error)
	ReadFile(ctx context.Context, in *ReadFileReq, opts ...grpc.CallOption) (FuseFS_ReadFileClient, error)
	ReadDir(ctx context.Context, in *ReadDirReq, opts ...grpc.CallOption) (*ReadDirResp, error)
}

type fuseFSClient struct {
	cc grpc.ClientConnInterface
}

func NewFuseFSClient(cc grpc.ClientConnInterface) FuseFSClient {
	return &fuseFSClient{cc}
}

func (c *fuseFSClient) NotifyAccess(ctx context.Context, in *NotifyAccessReq, opts ...grpc.CallOption) (*NotifyAccessResp, error) {
	out := new(NotifyAccessResp)
	err := c.cc.Invoke(ctx, "/filewatcher_model.FuseFS/NotifyAccess", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fuseFSClient) NotifyRead(ctx context.Context, in *NotifyReadReq, opts ...grpc.CallOption) (*NotifyReadResp, error) {
	out := new(NotifyReadResp)
	err := c.cc.Invoke(ctx, "/filewatcher_model.FuseFS/NotifyRead", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fuseFSClient) ReadFile(ctx context.Context, in *ReadFileReq, opts ...grpc.CallOption) (FuseFS_ReadFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &FuseFS_ServiceDesc.Streams[0], "/filewatcher_model.FuseFS/ReadFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &fuseFSReadFileClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FuseFS_ReadFileClient interface {
	Recv() (*ReadFileResp, error)
	grpc.ClientStream
}

type fuseFSReadFileClient struct {
	grpc.ClientStream
}

func (x *fuseFSReadFileClient) Recv() (*ReadFileResp, error) {
	m := new(ReadFileResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fuseFSClient) ReadDir(ctx context.Context, in *ReadDirReq, opts ...grpc.CallOption) (*ReadDirResp, error) {
	out := new(ReadDirResp)
	err := c.cc.Invoke(ctx, "/filewatcher_model.FuseFS/ReadDir", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FuseFSServer is the server API for FuseFS service.
// All implementations must embed UnimplementedFuseFSServer
// for forward compatibility
type FuseFSServer interface {
	NotifyAccess(context.Context, *NotifyAccessReq) (*NotifyAccessResp, error)
	NotifyRead(context.Context, *NotifyReadReq) (*NotifyReadResp, error)
	ReadFile(*ReadFileReq, FuseFS_ReadFileServer) error
	ReadDir(context.Context, *ReadDirReq) (*ReadDirResp, error)
	mustEmbedUnimplementedFuseFSServer()
}

// UnimplementedFuseFSServer must be embedded to have forward compatible implementations.
type UnimplementedFuseFSServer struct {
}

func (UnimplementedFuseFSServer) NotifyAccess(context.Context, *NotifyAccessReq) (*NotifyAccessResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyAccess not implemented")
}
func (UnimplementedFuseFSServer) NotifyRead(context.Context, *NotifyReadReq) (*NotifyReadResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyRead not implemented")
}
func (UnimplementedFuseFSServer) ReadFile(*ReadFileReq, FuseFS_ReadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method ReadFile not implemented")
}
func (UnimplementedFuseFSServer) ReadDir(context.Context, *ReadDirReq) (*ReadDirResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadDir not implemented")
}
func (UnimplementedFuseFSServer) mustEmbedUnimplementedFuseFSServer() {}

// UnsafeFuseFSServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FuseFSServer will
// result in compilation errors.
type UnsafeFuseFSServer interface {
	mustEmbedUnimplementedFuseFSServer()
}

func RegisterFuseFSServer(s grpc.ServiceRegistrar, srv FuseFSServer) {
	s.RegisterService(&FuseFS_ServiceDesc, srv)
}

func _FuseFS_NotifyAccess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyAccessReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FuseFSServer).NotifyAccess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/filewatcher_model.FuseFS/NotifyAccess",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FuseFSServer).NotifyAccess(ctx, req.(*NotifyAccessReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _FuseFS_NotifyRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyReadReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FuseFSServer).NotifyRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/filewatcher_model.FuseFS/NotifyRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FuseFSServer).NotifyRead(ctx, req.(*NotifyReadReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _FuseFS_ReadFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ReadFileReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FuseFSServer).ReadFile(m, &fuseFSReadFileServer{stream})
}

type FuseFS_ReadFileServer interface {
	Send(*ReadFileResp) error
	grpc.ServerStream
}

type fuseFSReadFileServer struct {
	grpc.ServerStream
}

func (x *fuseFSReadFileServer) Send(m *ReadFileResp) error {
	return x.ServerStream.SendMsg(m)
}

func _FuseFS_ReadDir_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadDirReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FuseFSServer).ReadDir(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/filewatcher_model.FuseFS/ReadDir",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FuseFSServer).ReadDir(ctx, req.(*ReadDirReq))
	}
	return interceptor(ctx, in, info, handler)
}

// FuseFS_ServiceDesc is the grpc.ServiceDesc for FuseFS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FuseFS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "filewatcher_model.FuseFS",
	HandlerType: (*FuseFSServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NotifyAccess",
			Handler:    _FuseFS_NotifyAccess_Handler,
		},
		{
			MethodName: "NotifyRead",
			Handler:    _FuseFS_NotifyRead_Handler,
		},
		{
			MethodName: "ReadDir",
			Handler:    _FuseFS_ReadDir_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ReadFile",
			Handler:       _FuseFS_ReadFile_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "FuseMessage.proto",
}