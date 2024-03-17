package grpcapp

import (
	"context"
	ssov1 "github.com/LeeAntonV/Protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/mail"
	"strings"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int, err error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
	ValidateCode(ctx context.Context, code string) (bool, error)
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
	if err := validateCredentials(req); err != nil {
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

func (s *serverAPI) ValidateCode(ctx context.Context, req *ssov1.CodeRequest) (*ssov1.CodeResponse, error) {
	if len(strings.TrimSpace(req.Code)) == emptyvalue {
		return nil, statuse.Error(codes.Internal, "Code field must not be empty")
	}

	validCode, err := s.auth.ValidateCode(ctx, req.Code)
	if err != nil {
		return nil, status.Error(codes.Internal, "Wrong code")
	}

	return &ssov1.CodeResponse{
		ValidCode: validCode,
	}, nil

}

func validateLogin(req *ssov1.LoginRequest) error {
	if err := validateCredentials; err != nil {
		return status.Error(codes.InvalidArgument, "invalid credentials")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "Invalid app id")
	}

	return nil
}

func validateCredentials(req *ssov1.RegisterRequest) error {
	_, err := mail.ParseAddress(req.GetEmail())
	if err != nil || len(strings.TrimSpace(req.GetPassword())) == emptyValue {
		return status.Error(codes.InvalidArgument, "invalid credentials")
	}

	return nil
}
