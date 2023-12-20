// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.3
// source: index.proto

package index

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Service_Create_FullMethodName               = "/index.v1.Service/Create"
	Service_CreateWithStreamData_FullMethodName = "/index.v1.Service/CreateWithStreamData"
	Service_UpdateNode_FullMethodName           = "/index.v1.Service/UpdateNode"
	Service_DeleteNodes_FullMethodName          = "/index.v1.Service/DeleteNodes"
	Service_ListNodes_FullMethodName            = "/index.v1.Service/ListNodes"
	Service_PatchRecords_FullMethodName         = "/index.v1.Service/PatchRecords"
	Service_ListRecords_FullMethodName          = "/index.v1.Service/ListRecords"
	Service_Search_FullMethodName               = "/index.v1.Service/Search"
)

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceClient interface {
	// Create allows to create the new records for a search.
	Create(ctx context.Context, in *CreateRecordsRequest, opts ...grpc.CallOption) (*CreateRecordsResult, error)
	// CreateWithStreamData allows to create new index records by streaming the records.
	CreateWithStreamData(ctx context.Context, opts ...grpc.CallOption) (Service_CreateWithStreamDataClient, error)
	// UpdateNode allows to update Node data, e.g. tags.
	UpdateNode(ctx context.Context, in *UpdateNodeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// DeleteNode allows to delete nodes according to the request provided
	DeleteNodes(ctx context.Context, in *DeleteNodesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// ListNodes returns all known children for the Path provided
	ListNodes(ctx context.Context, in *Path, opts ...grpc.CallOption) (*Nodes, error)
	// Patch allows to insert, update or delete an index's records
	PatchRecords(ctx context.Context, in *PatchRecordsRequest, opts ...grpc.CallOption) (*PatchRecordsResult, error)
	// ListRecords returns list of records for a path associated with it
	ListRecords(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListRecordsResult, error)
	// Search runs the search across all the index records matching the query. Result will
	// be ordered by the ranks for the request.
	Search(ctx context.Context, in *SearchRecordsRequest, opts ...grpc.CallOption) (*SearchRecordsResult, error)
}

type serviceClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceClient(cc grpc.ClientConnInterface) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) Create(ctx context.Context, in *CreateRecordsRequest, opts ...grpc.CallOption) (*CreateRecordsResult, error) {
	out := new(CreateRecordsResult)
	err := c.cc.Invoke(ctx, Service_Create_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) CreateWithStreamData(ctx context.Context, opts ...grpc.CallOption) (Service_CreateWithStreamDataClient, error) {
	stream, err := c.cc.NewStream(ctx, &Service_ServiceDesc.Streams[0], Service_CreateWithStreamData_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceCreateWithStreamDataClient{stream}
	return x, nil
}

type Service_CreateWithStreamDataClient interface {
	Send(*CreateIndexStreamRequest) error
	CloseAndRecv() (*CreateRecordsResult, error)
	grpc.ClientStream
}

type serviceCreateWithStreamDataClient struct {
	grpc.ClientStream
}

func (x *serviceCreateWithStreamDataClient) Send(m *CreateIndexStreamRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *serviceCreateWithStreamDataClient) CloseAndRecv() (*CreateRecordsResult, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(CreateRecordsResult)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *serviceClient) UpdateNode(ctx context.Context, in *UpdateNodeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Service_UpdateNode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) DeleteNodes(ctx context.Context, in *DeleteNodesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, Service_DeleteNodes_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ListNodes(ctx context.Context, in *Path, opts ...grpc.CallOption) (*Nodes, error) {
	out := new(Nodes)
	err := c.cc.Invoke(ctx, Service_ListNodes_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) PatchRecords(ctx context.Context, in *PatchRecordsRequest, opts ...grpc.CallOption) (*PatchRecordsResult, error) {
	out := new(PatchRecordsResult)
	err := c.cc.Invoke(ctx, Service_PatchRecords_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) ListRecords(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListRecordsResult, error) {
	out := new(ListRecordsResult)
	err := c.cc.Invoke(ctx, Service_ListRecords_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) Search(ctx context.Context, in *SearchRecordsRequest, opts ...grpc.CallOption) (*SearchRecordsResult, error) {
	out := new(SearchRecordsResult)
	err := c.cc.Invoke(ctx, Service_Search_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceServer is the server API for Service service.
// All implementations must embed UnimplementedServiceServer
// for forward compatibility
type ServiceServer interface {
	// Create allows to create the new records for a search.
	Create(context.Context, *CreateRecordsRequest) (*CreateRecordsResult, error)
	// CreateWithStreamData allows to create new index records by streaming the records.
	CreateWithStreamData(Service_CreateWithStreamDataServer) error
	// UpdateNode allows to update Node data, e.g. tags.
	UpdateNode(context.Context, *UpdateNodeRequest) (*emptypb.Empty, error)
	// DeleteNode allows to delete nodes according to the request provided
	DeleteNodes(context.Context, *DeleteNodesRequest) (*emptypb.Empty, error)
	// ListNodes returns all known children for the Path provided
	ListNodes(context.Context, *Path) (*Nodes, error)
	// Patch allows to insert, update or delete an index's records
	PatchRecords(context.Context, *PatchRecordsRequest) (*PatchRecordsResult, error)
	// ListRecords returns list of records for a path associated with it
	ListRecords(context.Context, *ListRequest) (*ListRecordsResult, error)
	// Search runs the search across all the index records matching the query. Result will
	// be ordered by the ranks for the request.
	Search(context.Context, *SearchRecordsRequest) (*SearchRecordsResult, error)
	mustEmbedUnimplementedServiceServer()
}

// UnimplementedServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (UnimplementedServiceServer) Create(context.Context, *CreateRecordsRequest) (*CreateRecordsResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedServiceServer) CreateWithStreamData(Service_CreateWithStreamDataServer) error {
	return status.Errorf(codes.Unimplemented, "method CreateWithStreamData not implemented")
}
func (UnimplementedServiceServer) UpdateNode(context.Context, *UpdateNodeRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNode not implemented")
}
func (UnimplementedServiceServer) DeleteNodes(context.Context, *DeleteNodesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNodes not implemented")
}
func (UnimplementedServiceServer) ListNodes(context.Context, *Path) (*Nodes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListNodes not implemented")
}
func (UnimplementedServiceServer) PatchRecords(context.Context, *PatchRecordsRequest) (*PatchRecordsResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PatchRecords not implemented")
}
func (UnimplementedServiceServer) ListRecords(context.Context, *ListRequest) (*ListRecordsResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRecords not implemented")
}
func (UnimplementedServiceServer) Search(context.Context, *SearchRecordsRequest) (*SearchRecordsResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedServiceServer) mustEmbedUnimplementedServiceServer() {}

// UnsafeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceServer will
// result in compilation errors.
type UnsafeServiceServer interface {
	mustEmbedUnimplementedServiceServer()
}

func RegisterServiceServer(s grpc.ServiceRegistrar, srv ServiceServer) {
	s.RegisterService(&Service_ServiceDesc, srv)
}

func _Service_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_Create_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Create(ctx, req.(*CreateRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_CreateWithStreamData_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ServiceServer).CreateWithStreamData(&serviceCreateWithStreamDataServer{stream})
}

type Service_CreateWithStreamDataServer interface {
	SendAndClose(*CreateRecordsResult) error
	Recv() (*CreateIndexStreamRequest, error)
	grpc.ServerStream
}

type serviceCreateWithStreamDataServer struct {
	grpc.ServerStream
}

func (x *serviceCreateWithStreamDataServer) SendAndClose(m *CreateRecordsResult) error {
	return x.ServerStream.SendMsg(m)
}

func (x *serviceCreateWithStreamDataServer) Recv() (*CreateIndexStreamRequest, error) {
	m := new(CreateIndexStreamRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Service_UpdateNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).UpdateNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_UpdateNode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).UpdateNode(ctx, req.(*UpdateNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_DeleteNodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteNodesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).DeleteNodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_DeleteNodes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).DeleteNodes(ctx, req.(*DeleteNodesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_ListNodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Path)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).ListNodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_ListNodes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).ListNodes(ctx, req.(*Path))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_PatchRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PatchRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).PatchRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_PatchRecords_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).PatchRecords(ctx, req.(*PatchRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_ListRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).ListRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_ListRecords_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).ListRecords(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Service_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Service_Search_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceServer).Search(ctx, req.(*SearchRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Service_ServiceDesc is the grpc.ServiceDesc for Service service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Service_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "index.v1.Service",
	HandlerType: (*ServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _Service_Create_Handler,
		},
		{
			MethodName: "UpdateNode",
			Handler:    _Service_UpdateNode_Handler,
		},
		{
			MethodName: "DeleteNodes",
			Handler:    _Service_DeleteNodes_Handler,
		},
		{
			MethodName: "ListNodes",
			Handler:    _Service_ListNodes_Handler,
		},
		{
			MethodName: "PatchRecords",
			Handler:    _Service_PatchRecords_Handler,
		},
		{
			MethodName: "ListRecords",
			Handler:    _Service_ListRecords_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _Service_Search_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CreateWithStreamData",
			Handler:       _Service_CreateWithStreamData_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "index.proto",
}
