package grpcapp

import (
	"context"
	ssov1 "github.com/LeeAntonV/Protos/gen/go/sso"
	"google.golang.org/grpc"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	panic("Login error")
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	panic("Register error")
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("IsAdmin error")
}
