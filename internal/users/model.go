package users

import "time"

type User struct {
	Id           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUserParams struct {
	Username     string
	PasswordHash string
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
