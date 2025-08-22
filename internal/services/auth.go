package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"focusflow-be/internal/config"
	"focusflow-be/internal/models"
)

type AuthService struct {
	config *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

func (s *AuthService) CreateJWT(userSession *models.UserSession) (string, error) {
	claims := &Claims{
		UserID: userSession.UserID,
		Email:  userSession.Email,
		Name:   userSession.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *AuthService) VerifyJWT(tokenString string) (*models.UserSession, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &models.UserSession{
		UserID: claims.UserID,
		Email:  claims.Email,
		Name:   claims.Name,
	}, nil
}