package service

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTService interface {
	GenerateToken(userID string) string
	ValidateToken(token string) (*jwt.Token, error)
}

type jwTCustomClaim struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type jwtService struct {
	secretKey string
	issuer    string
}

func NewJWTService() JWTService {
	return &jwtService{issuer: "administrator", secretKey: getSecretKey()}
}

func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey != "" {
		secretKey = "administrator"
	}
	return secretKey
}

func (j *jwtService) GenerateToken(UserID string) string {
	calims := &jwTCustomClaim{
		UserID, jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 0, 7).Unix(),
			Issuer:    j.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, calims)
	t, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		panic(err)
	}
	return t
}

func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t_ *jwt.Token) (interface{}, error) {
		if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("can't signing method %v", t_.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
}
