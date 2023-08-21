// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: service.proto

package proto

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

// CacheClient is the client API for Cache service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CacheClient interface {
	Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (Cache_SearchClient, error)
}

type cacheClient struct {
	cc grpc.ClientConnInterface
}

func NewCacheClient(cc grpc.ClientConnInterface) CacheClient {
	return &cacheClient{cc}
}

func (c *cacheClient) Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error) {
	out := new(PutResponse)
	err := c.cc.Invoke(ctx, "/Cache/Put", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/Cache/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (Cache_SearchClient, error) {
	stream, err := c.cc.NewStream(ctx, &Cache_ServiceDesc.Streams[0], "/Cache/Search", opts...)
	if err != nil {
		return nil, err
	}
	x := &cacheSearchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Cache_SearchClient interface {
	Recv() (*SearchResponse, error)
	grpc.ClientStream
}

type cacheSearchClient struct {
	grpc.ClientStream
}

func (x *cacheSearchClient) Recv() (*SearchResponse, error) {
	m := new(SearchResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CacheServer is the server API for Cache service.
// All implementations must embed UnimplementedCacheServer
// for forward compatibility
type CacheServer interface {
	Put(context.Context, *PutRequest) (*PutResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Search(*SearchRequest, Cache_SearchServer) error
	mustEmbedUnimplementedCacheServer()
}

// UnimplementedCacheServer must be embedded to have forward compatible implementations.
type UnimplementedCacheServer struct {
}

func (UnimplementedCacheServer) Put(context.Context, *PutRequest) (*PutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Put not implemented")
}
func (UnimplementedCacheServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedCacheServer) Search(*SearchRequest, Cache_SearchServer) error {
	return status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedCacheServer) mustEmbedUnimplementedCacheServer() {}

// UnsafeCacheServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CacheServer will
// result in compilation errors.
type UnsafeCacheServer interface {
	mustEmbedUnimplementedCacheServer()
}

func RegisterCacheServer(s grpc.ServiceRegistrar, srv CacheServer) {
	s.RegisterService(&Cache_ServiceDesc, srv)
}

func _Cache_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Cache/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).Put(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Cache/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cache_Search_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SearchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CacheServer).Search(m, &cacheSearchServer{stream})
}

type Cache_SearchServer interface {
	Send(*SearchResponse) error
	grpc.ServerStream
}

type cacheSearchServer struct {
	grpc.ServerStream
}

func (x *cacheSearchServer) Send(m *SearchResponse) error {
	return x.ServerStream.SendMsg(m)
}

// Cache_ServiceDesc is the grpc.ServiceDesc for Cache service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Cache_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Cache",
	HandlerType: (*CacheServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Put",
			Handler:    _Cache_Put_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Cache_Get_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Search",
			Handler:       _Cache_Search_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "service.proto",
}
