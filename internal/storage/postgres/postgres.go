package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gRPC/internal/domain/models"
	"gRPC/internal/storage"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser saves user to db
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte, hashCode []byte) (uid int64, err error) {
	const op = "storage.postgres.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO user_profile(email, hash, code) VALUES($1,$2,$3)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, email, passHash, hashCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64
	stmt, err = s.db.Prepare("SELECT id FROM user_profile WHERE email = $1")
	row := stmt.QueryRowContext(ctx, email)

	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// User returns user by email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	stmt, err := s.db.Prepare("SELECT id, email, hash FROM user_profile WHERE email = $1")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.PassHash, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// ValidateCode returns confirmation code by email
func (s *Storage) ValidateCode(ctx context.Context, email string) (string, error) {
	const op = "storage.postgres.ValidateCode"

	stmt, err := s.db.Prepare("SELECT code FROM user_profile WHERE email = $1")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var code string
	err = row.Scan(&code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return code, nil
}

// IsAdmin returns admin status by user id
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	stmt, err := s.db.Prepare("SELECT isadmin FROM user_profile WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

// App returns app by id
func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.postgres.App"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = $1")
	if err != nil {
		return models.App{}, err
	}

	row := stmt.QueryRowContext(ctx, id)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)

	}

	return app, nil
}

func AcceptCode(email string) error {
	const op = "storage.postgres.AcceptCode"
	err := godotenv.Load()

	db, err := sql.Open("postgres", os.Getenv("StoragePath"))
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("UPDATE user_profile SET verified = true WHERE email = $1")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}
