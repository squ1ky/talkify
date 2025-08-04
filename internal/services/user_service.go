package services

import (
	"errors"
	"github.com/squ1ky/talkify/internal/database"
	"github.com/squ1ky/talkify/internal/models"
)

var (
	ErrUserExists     = errors.New("user already exists")
	ErrBadCredentials = errors.New("invalid username or password")
	ErrNotFound       = errors.New("user not found")
)

// UserService manages user-related business logic
type UserService struct {
	users *database.UserRepository
}

// NewUserService creates new user service
func NewUserService(users *database.UserRepository) *UserService {
	return &UserService{users: users}
}

// Register new user, returns UserResponse or error
func (s *UserService) Register(req models.UserCreateRequest) (*models.UserResponse, error) {
	if !models.IsValidUsername(req.Username) {
		return nil, errors.New("invalid username format")
	}
	if exists, _ := s.users.Exists(req.Username); exists {
		return nil, ErrUserExists
	}

	user, err := models.CreateUserFromRequest(req)
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}

// Login user, returns User and error
func (s *UserService) Login(req models.UserLoginRequest) (*models.User, error) {
	user, err := s.users.GetByUsername(req.Username)
	if err != nil {
		return nil, ErrBadCredentials
	}

	if !user.CheckPassword(req.Password) {
		return nil, ErrBadCredentials
	}
	return user, nil
}

// List all users (for chat)
func (s *UserService) List(limit, offset int) ([]models.UserResponse, int, error) {
	users, err := s.users.List(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, u.ToResponse())
	}

	count, err := s.users.Count()
	if err != nil {
		return nil, 0, err
	}

	return resp, count, nil
}

// GetByID returns user without password
func (s *UserService) GetByID(userID int) (*models.UserResponse, error) {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return nil, ErrNotFound
	}

	resp := user.ToResponse()
	return &resp, nil
}
