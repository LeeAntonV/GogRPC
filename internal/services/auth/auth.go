package auth

import (
	"context"
	"errors"
	"fmt"
	"gRPC/internal/domain/models"
	codesender "gRPC/internal/lib/email"
	"gRPC/internal/lib/jwt"
	"gRPC/internal/lib/sl"
	"gRPC/internal/storage"
	"gRPC/internal/storage/postgres"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"strconv"
	"time"
)

var (
	ErrUserExists = errors.New("user already exists")
)

type Auth struct {
	log          *slog.Logger
	usrSaver     UserSaver
	usrProvider  UserProvider
	appProvider  AppProvider
	codeProvider CodeProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte, verCode []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	IsAdmin(ctx context.Context, userID int64) (status bool, err error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

type CodeProvider interface {
	ValidateCode(ctx context.Context, email string) (code string, err error)
}

var (
	InvalidCredentials = errors.New("invalid credentials")
)

// New returns new instance of Auth service.
func New(log *slog.Logger, usrSaver UserSaver, usrProvider UserProvider, appProvider AppProvider, codeProvider CodeProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		usrSaver:     usrSaver,
		usrProvider:  usrProvider,
		log:          log,
		appProvider:  appProvider,
		codeProvider: codeProvider,
		tokenTTL:     tokenTTL,
	}
}

// Login verifies if given credentials exist in the system
//
// If user exists, but password is incorrect returns error
// If user do not exist, returns error
func (a *Auth) Login(
	ctx context.Context, email string, password string, appID int,
) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("logging into user account")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		log.Error("failed to login into user account", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, InvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged successfully")

	token, err := jwt.CreateNewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("Failed to create token", sl.Err(err))
		return "", fmt.Errorf("%s :%w", op, err)
	}

	log.Info("Successful logging")
	return token, nil
}

// RegisterNewUser verifies if user with this email do not exist
//
// If user exists, returns error
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "Auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		return 2, fmt.Errorf("%s: %w", op, err)
	}

	verificationCode, err := codesender.SendEmail(email)
	fmt.Println(err)
	if err != nil {
		fmt.Println(err)
		log.Error("Failed to send code")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	stringCode := strconv.Itoa(verificationCode)
	hashedCode, err := bcrypt.GenerateFromPassword([]byte(stringCode), bcrypt.DefaultCost)

	id, err := a.usrSaver.SaveUser(ctx, email, passHash, hashedCode)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))
		return 1, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

// ValidateCode verifies if confirmation code provided by user is valid
//
// If so return true, else false
func (a *Auth) ValidateCode(
	ctx context.Context, email string, code string,
) (bool, error) {
	const op = "Auth.ValidateCode"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("Trying to validate confirmation code")

	dbCode, err := a.codeProvider.ValidateCode(ctx, email)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbCode), []byte(code)); err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	err = postgres.AcceptCode(email)
	if err != nil {
		fmt.Println(err)
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Code is valid")

	return true, nil
}

// IsAdmin verifies if user is admin
//
// If user is not admin, returns false, else true
func (a *Auth) IsAdmin(
	ctx context.Context, userID int,
) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("uid", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, int64(userID))
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
