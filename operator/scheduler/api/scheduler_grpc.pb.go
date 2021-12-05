// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package scheduler

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

// SchedulerClient is the client API for Scheduler service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SchedulerClient interface {
	ServerStatus(ctx context.Context, in *ServerReference, opts ...grpc.CallOption) (*ServerStatusResponse, error)
	LoadModel(ctx context.Context, in *LoadModelRequest, opts ...grpc.CallOption) (*LoadModelResponse, error)
	UnloadModel(ctx context.Context, in *ModelReference, opts ...grpc.CallOption) (*UnloadModelResponse, error)
	ModelStatus(ctx context.Context, in *ModelReference, opts ...grpc.CallOption) (*ModelStatusResponse, error)
	SubscribeModelEvents(ctx context.Context, in *ModelSubscriptionRequest, opts ...grpc.CallOption) (Scheduler_SubscribeModelEventsClient, error)
}

type schedulerClient struct {
	cc grpc.ClientConnInterface
}

func NewSchedulerClient(cc grpc.ClientConnInterface) SchedulerClient {
	return &schedulerClient{cc}
}

func (c *schedulerClient) ServerStatus(ctx context.Context, in *ServerReference, opts ...grpc.CallOption) (*ServerStatusResponse, error) {
	out := new(ServerStatusResponse)
	err := c.cc.Invoke(ctx, "/seldon.mlops.scheduler.Scheduler/ServerStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) LoadModel(ctx context.Context, in *LoadModelRequest, opts ...grpc.CallOption) (*LoadModelResponse, error) {
	out := new(LoadModelResponse)
	err := c.cc.Invoke(ctx, "/seldon.mlops.scheduler.Scheduler/LoadModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) UnloadModel(ctx context.Context, in *ModelReference, opts ...grpc.CallOption) (*UnloadModelResponse, error) {
	out := new(UnloadModelResponse)
	err := c.cc.Invoke(ctx, "/seldon.mlops.scheduler.Scheduler/UnloadModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) ModelStatus(ctx context.Context, in *ModelReference, opts ...grpc.CallOption) (*ModelStatusResponse, error) {
	out := new(ModelStatusResponse)
	err := c.cc.Invoke(ctx, "/seldon.mlops.scheduler.Scheduler/ModelStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerClient) SubscribeModelEvents(ctx context.Context, in *ModelSubscriptionRequest, opts ...grpc.CallOption) (Scheduler_SubscribeModelEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Scheduler_ServiceDesc.Streams[0], "/seldon.mlops.scheduler.Scheduler/SubscribeModelEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &schedulerSubscribeModelEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Scheduler_SubscribeModelEventsClient interface {
	Recv() (*ModelEventMessage, error)
	grpc.ClientStream
}

type schedulerSubscribeModelEventsClient struct {
	grpc.ClientStream
}

func (x *schedulerSubscribeModelEventsClient) Recv() (*ModelEventMessage, error) {
	m := new(ModelEventMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SchedulerServer is the server API for Scheduler service.
// All implementations must embed UnimplementedSchedulerServer
// for forward compatibility
type SchedulerServer interface {
	ServerStatus(context.Context, *ServerReference) (*ServerStatusResponse, error)
	LoadModel(context.Context, *LoadModelRequest) (*LoadModelResponse, error)
	UnloadModel(context.Context, *ModelReference) (*UnloadModelResponse, error)
	ModelStatus(context.Context, *ModelReference) (*ModelStatusResponse, error)
	SubscribeModelEvents(*ModelSubscriptionRequest, Scheduler_SubscribeModelEventsServer) error
	mustEmbedUnimplementedSchedulerServer()
}

// UnimplementedSchedulerServer must be embedded to have forward compatible implementations.
type UnimplementedSchedulerServer struct {
}

func (UnimplementedSchedulerServer) ServerStatus(context.Context, *ServerReference) (*ServerStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ServerStatus not implemented")
}
func (UnimplementedSchedulerServer) LoadModel(context.Context, *LoadModelRequest) (*LoadModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadModel not implemented")
}
func (UnimplementedSchedulerServer) UnloadModel(context.Context, *ModelReference) (*UnloadModelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnloadModel not implemented")
}
func (UnimplementedSchedulerServer) ModelStatus(context.Context, *ModelReference) (*ModelStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ModelStatus not implemented")
}
func (UnimplementedSchedulerServer) SubscribeModelEvents(*ModelSubscriptionRequest, Scheduler_SubscribeModelEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeModelEvents not implemented")
}
func (UnimplementedSchedulerServer) mustEmbedUnimplementedSchedulerServer() {}

// UnsafeSchedulerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SchedulerServer will
// result in compilation errors.
type UnsafeSchedulerServer interface {
	mustEmbedUnimplementedSchedulerServer()
}

func RegisterSchedulerServer(s grpc.ServiceRegistrar, srv SchedulerServer) {
	s.RegisterService(&Scheduler_ServiceDesc, srv)
}

func _Scheduler_ServerStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServerReference)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).ServerStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/seldon.mlops.scheduler.Scheduler/ServerStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).ServerStatus(ctx, req.(*ServerReference))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_LoadModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).LoadModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/seldon.mlops.scheduler.Scheduler/LoadModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).LoadModel(ctx, req.(*LoadModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_UnloadModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModelReference)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).UnloadModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/seldon.mlops.scheduler.Scheduler/UnloadModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).UnloadModel(ctx, req.(*ModelReference))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_ModelStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModelReference)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerServer).ModelStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/seldon.mlops.scheduler.Scheduler/ModelStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerServer).ModelStatus(ctx, req.(*ModelReference))
	}
	return interceptor(ctx, in, info, handler)
}

func _Scheduler_SubscribeModelEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ModelSubscriptionRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SchedulerServer).SubscribeModelEvents(m, &schedulerSubscribeModelEventsServer{stream})
}

type Scheduler_SubscribeModelEventsServer interface {
	Send(*ModelEventMessage) error
	grpc.ServerStream
}

type schedulerSubscribeModelEventsServer struct {
	grpc.ServerStream
}

func (x *schedulerSubscribeModelEventsServer) Send(m *ModelEventMessage) error {
	return x.ServerStream.SendMsg(m)
}

// Scheduler_ServiceDesc is the grpc.ServiceDesc for Scheduler service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Scheduler_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "seldon.mlops.scheduler.Scheduler",
	HandlerType: (*SchedulerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ServerStatus",
			Handler:    _Scheduler_ServerStatus_Handler,
		},
		{
			MethodName: "LoadModel",
			Handler:    _Scheduler_LoadModel_Handler,
		},
		{
			MethodName: "UnloadModel",
			Handler:    _Scheduler_UnloadModel_Handler,
		},
		{
			MethodName: "ModelStatus",
			Handler:    _Scheduler_ModelStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubscribeModelEvents",
			Handler:       _Scheduler_SubscribeModelEvents_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "scheduler.proto",
}
