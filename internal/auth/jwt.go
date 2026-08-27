package auth

import (
	"errors"
	"time"

	"github.com/7DeN4iK7/auth-service/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidClaims = errors.New("invalid claims")
	ErrInvalidToken  = errors.New("invalid token")
)

type JWTService struct {
	secret []byte
}

func NewJWT(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secret: []byte(cfg.Secret),
	}
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func (j *JWTService) GenerateToken(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secret)
}

func (j *JWTService) GetUserID(token *jwt.Token) (int, error) {
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, ErrInvalidClaims
	}

	return claims.UserID, nil
}

func (j *JWTService) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return j.secret, nil
		},
	)
}
