package auth

import (
	"crypto/subtle"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const adminTokenDuration = 24 * time.Hour

type AdminClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AdminService struct {
	secret   []byte
	username string
	password string
}

func NewAdminService(secret, username, password string) *AdminService {
	return &AdminService{secret: []byte(secret), username: username, password: password}
}

func (s *AdminService) Login(username, password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(username), []byte(s.username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) != 1 {
		return "", errors.New("账号或密码错误")
	}

	now := time.Now()
	claims := AdminClaims{
		Username: s.username,
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(adminTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *AdminService) Parse(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid || claims.Role != "admin" || claims.Username != s.username {
		return nil, errors.New("invalid admin token")
	}
	return claims, nil
}
