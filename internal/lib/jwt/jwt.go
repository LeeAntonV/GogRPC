package jwt

import (
	"gRPC/internal/domain/models"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// CreateNewToken generates new token by HS256 signing algorithm
func CreateNewToken(user models.User, app models.App, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["user"] = user.Name
	claims["exp"] = time.Now().Add(tokenTTL).Unix()
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString(app.Secret)
	if err != nil {
		return "", status.Error(codes.Internal, "Internal error")
	}
	return tokenString, nil
}
