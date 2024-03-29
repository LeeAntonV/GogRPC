package grpcapp

import (
	"context"
	"fmt"
	ssov5 "github.com/LeeAntonV/Protos/gen/go/sso"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/mail"
	"strings"
	"time"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
	ValidateCode(ctx context.Context, email string, code string) (bool, error)
}

type serverAPI struct {
	ssov5.UnimplementedAuthServer
	auth Auth
}

const emptyValue = 0

func Register(gRPC *grpc.Server, auth Auth) {
	ssov5.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov5.LoginRequest) (*ssov5.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ssov5.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov5.RegisterRequest) (*ssov5.RegisterResponse, error) {
	if err := validateCredentials(req); err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &ssov5.RegisterResponse{
		UserId: userId,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov5.IsAdminRequest) (*ssov5.IsAdminResponse, error) {
	if req.UserId == emptyValue {
		return nil, status.Error(codes.Internal, "You are not allowed to admin panel")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, int(req.UserId))

	if err != nil {
		return nil, status.Error(codes.Internal, "Invalid user id")
	}

	return &ssov5.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) ValidCode(ctx context.Context, req *ssov5.CodeRequest) (*ssov5.CodeResponse, error) {
	if len(strings.TrimSpace(req.Code)) == emptyValue {
		return nil, status.Error(codes.Internal, "Code field must not be empty")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	val, err := rdb.Get(ctx, req.Email).Result()
	if val == req.Code {
		return &ssov5.CodeResponse{ValidCode: false}, status.Error(codes.Internal, "Wrong code")
	}

	validCode, err := s.auth.ValidateCode(ctx, req.Email, req.Code)
	if err != nil {
		if isUnique := UniqueValue(ctx, rdb, req.Email, req.Code, 1*time.Minute); !isUnique {
			rdb.Set(ctx, req.Email, req.Code, 1*time.Minute)
		}

		return nil, status.Error(codes.Internal, "Wrong code")
	}

	return &ssov5.CodeResponse{
		ValidCode: validCode,
	}, nil
}

func validateCredentials(req *ssov5.RegisterRequest) error {
	_, err := mail.ParseAddress(req.GetEmail())
	if err != nil || len(strings.TrimSpace(req.GetPassword())) == emptyValue {
		return status.Error(codes.InvalidArgument, "invalid credentials")
	}

	return nil
}

func UniqueValue(ctx context.Context, client *redis.Client, key string, value string, exp time.Duration) bool {
	isNew, err := client.SetNX(ctx, key, value, exp).Result()

	if err != nil {
		return false
	}

	return isNew
}
