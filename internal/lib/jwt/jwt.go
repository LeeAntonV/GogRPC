package jwt

import (
	"fmt"
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
	claims["user"] = user.Email
	claims["exp"] = time.Now().Add(tokenTTL).Unix()
	claims["app_id"] = app.ID

	tokenString, err := token.SignedString(app.Secret)
	if err != nil {
		return "", status.Error(codes.Internal, "Internal error")
	}
	return tokenString, nil
}

// CheckTokenValidity checks if jwt token is valid for system
//
// If token is valid returns true, if not, false
func CheckTokenValidity(tokenString string, app models.App) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return app.Secret, nil
	})

	if err != nil || !token.Valid {
		return fmt.Errorf("%s: %w", "Invalid token", err)
	}

	return nil
}
