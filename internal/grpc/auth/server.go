package grpcapp

import (
	"context"
	ssov3 "github.com/LeeAntonV/Protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/mail"
	"strings"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
	ValidateCode(ctx context.Context, email string, code string) (bool, error)
}

type serverAPI struct {
	ssov3.UnimplementedAuthServer
	auth Auth
}

const emptyValue = 0

func Register(gRPC *grpc.Server, auth Auth) {
	ssov3.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov3.LoginRequest) (*ssov3.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, status.Error(codes.Internal, "Invalid credentials")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov3.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov3.RegisterRequest) (*ssov3.RegisterResponse, error) {
	if err := validateCredentials(req); err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov3.RegisterResponse{
		UserId: userId,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov3.IsAdminRequest) (*ssov3.IsAdminResponse, error) {
	if req.UserId == emptyValue {
		return nil, status.Error(codes.Internal, "You are not allowed to admin panel")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, int(req.UserId))

	if err != nil {
		return nil, status.Error(codes.Internal, "Invalid user id")
	}

	return &ssov3.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) ValidCode(ctx context.Context, req *ssov3.CodeRequest) (*ssov3.CodeResponse, error) {
	if len(strings.TrimSpace(req.Code)) == emptyValue {
		return nil, status.Error(codes.Internal, "Code field must not be empty")
	}

	validCode, err := s.auth.ValidateCode(ctx, req.Email, req.Code)
	if err != nil {
		return nil, status.Error(codes.Internal, "Wrong code")
	}

	return &ssov3.CodeResponse{
		ValidCode: validCode,
	}, nil

}

func validateLogin(req *ssov3.LoginRequest) error {
	if err := validateCredentials; err != nil {
		return status.Error(codes.InvalidArgument, "invalid credentials")
	}

	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "Invalid app id")
	}

	return nil
}

func validateCredentials(req *ssov3.RegisterRequest) error {
	_, err := mail.ParseAddress(req.GetEmail())
	if err != nil || len(strings.TrimSpace(req.GetPassword())) == emptyValue {
		return status.Error(codes.InvalidArgument, "invalid credentials")
	}

	return nil
}
