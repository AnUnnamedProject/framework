// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

type (
	// GRPCServices is the gRPC interface used for service registration.
	GRPCServices interface{}

	// GRPCServer is the structure that contains the gRPC server.
	GRPCServer struct {
		*grpc.Server
	}

	// GRPCService implements the gRPC Service.
	// All user defined services must use this structure.
	GRPCService struct{}
)

// NewGRPC instantiates a new gRPC server.
func NewGRPC(opt ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{grpc.NewServer(opt...)}
}

// RegisterGRPCService registers a new gRPC service.
// Registered services are called via Init(s *framework.GRPCServer) where you finally register
// your protos servers before Serving the gRPC.
func RegisterGRPCService(s GRPCServices) {
	App.grpcServices = append(App.grpcServices, s)
}

// UseGRPCUnaryInterceptor appends a new global gRPC unary interceptor.
func UseGRPCUnaryInterceptor(filter grpc.UnaryServerInterceptor) {
	App.grpcUnaryInterceptors = append(App.grpcUnaryInterceptors, filter)
}

// UseGRPCStreamInterceptor appends a new global gRPC stream interceptor.
func UseGRPCStreamInterceptor(filter grpc.StreamServerInterceptor) {
	App.grpcStreamInterceptors = append(App.grpcStreamInterceptors, filter)
}

// ServerUnaryInterceptor creates a single unary interceptor from a list of interceptors.
// Use UseGRPCUnaryInterceptor to add a new global unary interceptor.
func ServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		interceptor := func(current grpc.UnaryServerInterceptor, next grpc.UnaryHandler) grpc.UnaryHandler {
			return func(curCtx context.Context, curReq interface{}) (interface{}, error) {
				return current(curCtx, curReq, info, next)
			}
		}

		for i := len(App.grpcUnaryInterceptors) - 1; i >= 0; i-- {
			handler = interceptor(App.grpcUnaryInterceptors[i], handler)
		}
		return handler(ctx, req)
	}
}

// ServerStreamInterceptor creates a single stream interceptor from a list of interceptors.
// Use UseGRPCStreamInterceptor to add a new global stream interceptor.
func ServerStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		interceptor := func(current grpc.StreamServerInterceptor, next grpc.StreamHandler) grpc.StreamHandler {
			return func(curSrv interface{}, curStream grpc.ServerStream) error {
				return current(curSrv, curStream, info, next)
			}
		}

		for i := len(App.grpcStreamInterceptors) - 1; i >= 0; i-- {
			handler = interceptor(App.grpcStreamInterceptors[i], handler)
		}
		return handler(srv, stream)
	}
}
