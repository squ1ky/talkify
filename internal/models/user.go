package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserCreateRequest represents request for user creation
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// UserLoginRequest represents request for user login
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents user data in API responses
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// UserListResponse represents list of users in API responses
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

// HashPassword hashes the plain text password using bcrypt
func (u *User) HashPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedBytes)
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ToResponse converts User to UserResponse (without sensitive data)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
	}
}

// CreateUserFromRequest creates User from UserCreateRequest
func CreateUserFromRequest(req UserCreateRequest) (*User, error) {
	user := &User{
		Username:  req.Username,
		CreatedAt: time.Now(),
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, err
	}

	return user, nil
}

// IsValidUsername checks if username meets requirements
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}
