package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"online-judge-backend/login-flow-subset/internal/models"
	"online-judge-backend/login-flow-subset/internal/repositories"
)

type AuthService struct {
	users     *repositories.UserRepository
	jwtSecret []byte
}

type AuthUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
}

type AuthTokens struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type jwtClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(users *repositories.UserRepository, secret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(secret),
	}
}

func (s *AuthService) Register(input RegisterInput) (*AuthTokens, error) {
	if input.Username == "" || input.Password == "" {
		return nil, errors.New("username and password are required")
	}
	if len(input.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	if _, err := s.users.GetByUsername(input.Username); err == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: input.Username,
		Password: string(hash),
		Role:     models.UserRoleStudent,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return s.issueToken(user)
}

func (s *AuthService) Login(input LoginInput) (*AuthTokens, error) {
	if input.Username == "" || input.Password == "" {
		return nil, errors.New("username and password are required")
	}

	user, err := s.users.GetByUsername(input.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return s.issueToken(user)
}

func (s *AuthService) ParseToken(tokenString string) (*AuthUser, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &AuthUser{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func (s *AuthService) GetCurrentUser(id uint) (*AuthUser, error) {
	user, err := s.users.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

func (s *AuthService) issueToken(user *models.User) (*AuthTokens, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(72 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		Token: signed,
		User: AuthUser{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	}, nil
}
