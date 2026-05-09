package service

import (
	"errors"
	"strings"
	"time"

	"github.com/chatagent/server/config"
	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrWrongPassword      = errors.New("wrong password")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrSystemInitialized  = errors.New("system already initialized")
	ErrSamePassword       = errors.New("new password must be different from old password")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	repo *repository.UserRepository
	cfg  *config.Config
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, cfg: cfg}
}

func (s *AuthService) Status() (bool, error) {
	count, err := s.repo.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *AuthService) InitRoot(email, password, name string) (*model.User, string, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	name = strings.TrimSpace(name)

	if email == "" || password == "" || name == "" {
		return nil, "", errors.New("email, password and name are required")
	}
	if !strings.Contains(email, "@") {
		return nil, "", errors.New("invalid email format")
	}
	if len(password) < 6 {
		return nil, "", errors.New("password must be at least 6 characters")
	}

	count, err := s.repo.Count()
	if err != nil {
		return nil, "", err
	}
	if count > 0 {
		return nil, "", ErrSystemInitialized
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Email:              email,
		Name:               name,
		PasswordHash:       string(hash),
		Role:               "super_admin",
		Status:             "active",
		MustChangePassword: false,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(email, password string) (*model.User, string, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return nil, "", errors.New("email and password are required")
	}

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, "", ErrUserNotFound
	}

	if user.Status != "active" {
		return nil, "", ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrWrongPassword
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if oldPassword == newPassword {
		return ErrSamePassword
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrWrongPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(userID, string(hash))
}

func (s *AuthService) ValidateToken(tokenString string) (interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) GenerateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWT.ExpireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}
