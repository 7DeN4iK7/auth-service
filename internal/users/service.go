package users

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/7DeN4iK7/auth-service/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxUsernameLength = 32
	maxPasswordLength = 72
)

var (
	ErrNoPassword        = errors.New("no password entered")
	ErrNoUsername        = errors.New("no username entered")
	ErrTooLongPassword   = errors.New("password is too long")
	ErrTooLongUsername   = errors.New("username is too long")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrWrongPassword     = errors.New("wrong password")
	ErrGeneratingJWT     = errors.New("error generating JWT")
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

	if utf8.RuneCountInString(req.Password) > maxPasswordLength {
		return 0, ErrTooLongPassword
	}

	if req.Username == "" {
		return 0, ErrNoUsername
	}

	if utf8.RuneCountInString(req.Username) > maxUsernameLength {
		return 0, ErrTooLongUsername
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

	jwt, err := s.JWT.GenerateToken(user.ID)
	if err != nil {
		return "", ErrGeneratingJWT
	}

	return jwt, nil
}

func (s *Service) GetUserByID(ctx context.Context, id int) (User, error) {
	return s.Repository.GetUserByID(ctx, id)
}
