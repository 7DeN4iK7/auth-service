package users

import (
	"context"
	"errors"
	"strconv"

	"github.com/7DeN4iK7/auth-service/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoPassword        = errors.New("No password entered")
	ErrNoUsername        = errors.New("No username entered")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrWrongPassword     = errors.New("Wrong password")
	ErrGeneratingJWT     = errors.New("Error generating JWT")
)

type UserRepository interface {
	CreateUser(ctx context.Context, r CreateUserParams) (int, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
}

type Service struct {
	Repository UserRepository
	JWT        *auth.JWTService
}

func NewService(r UserRepository, jwt *auth.JWTService) *Service {
	return &Service{Repository: r, JWT: jwt}
}

func (s *Service) CreateUser(ctx context.Context, req RegisterRequest) (int, error) {
	if req.Password == "" {
		return 0, ErrNoPassword
	}

	if req.Username == "" {
		return 0, ErrNoUsername
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	params := CreateUserParams{
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	id, err := s.Repository.CreateUser(ctx, params)
	if err != nil {
		if errors.Is(err, ErrUniqueViolation) {
			return 0, ErrUserAlreadyExists
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (s *Service) LoginUser(ctx context.Context, r LoginRequest) (string, error) {
	if r.Password == "" {
		return "", ErrNoPassword
	}

	if r.Username == "" {
		return "", ErrNoUsername
	}

	user, err := s.Repository.GetUserByUsername(ctx, r.Username)

	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.Password))
	if err != nil {
		return "", ErrWrongPassword
	}

	id, _ := strconv.Atoi(user.Id)
	jwt, err := s.JWT.GenerateToken(id)
	if err != nil {
		return "", ErrGeneratingJWT
	}

	return jwt, nil
}

func (s *Service) GetUserByID(ctx context.Context, id int) (User, error) {
	return s.Repository.GetUserByID(ctx, id)
}
