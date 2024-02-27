package grpcapp

import (
	"context"
	ssov1 "github.com/LeeAntonV/Protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int, err error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

const emptyValue = 0

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, status.Error(codes.Internal, "Invalid credentials")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegisterNewUser(req); err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov1.RegisterResponse{
		UserId: int64(userId),
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if req.UserId == emptyValue {
		return nil, status.Error(codes.Internal, "You are not allowed to admin panel")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, int(req.UserId))

	if err != nil {
		return nil, status.Error(codes.Internal, "Invalid user id")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if len(strings.TrimSpace(req.GetEmail())) == emptyValue ||
		len(strings.TrimSpace(req.GetPassword())) == emptyValue {

		return status.Error(codes.InvalidArgument, "Invalid email or password")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "Invalid app id")
	}

	return nil
}

func validateRegisterNewUser(req *ssov1.RegisterRequest) error {
	if len(strings.TrimSpace(req.GetEmail())) == emptyValue ||
		len(strings.TrimSpace(req.GetPassword())) == emptyValue {

		return status.Error(codes.InvalidArgument, "Invalid email or password")
	}

	return nil
}
