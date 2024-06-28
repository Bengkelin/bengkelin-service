package crypto

import (
	"fmt"
	"time"

	"github.com/Bengkelin/bengkelin-service/internal/pkg/config"
	"github.com/golang-jwt/jwt"
)

var jwtHelper *jwtCryptoHelper

// Contract fot JWT Crypto Helper
type JWTCryptoHelper interface {
	GenerateToken(UserId string) (string, error)
	GenerateTokenMitra(MitraId string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	ValidateTokenMitra(tokenString string) (*jwt.Token, error)
}

// Struct for jwt custom claim
type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type jwtCustomClaimMitra struct {
	MitraId string `json:"mitra_id"`
	jwt.StandardClaims
}

// Struct for JWTHelper
type jwtCryptoHelper struct {
}

// Func to initialize new jwt crypto helper
func GetJWTCrypto() JWTCryptoHelper {
	if jwtHelper == nil {
		jwtHelper = &jwtCryptoHelper{}
	}
	return jwtHelper
}

// Func to Generate Token with User ID as main issuer
func (helper *jwtCryptoHelper) GenerateToken(UserID string) (string, error) {
	serverConfiguration := config.GetConfig().Server
	claims := &jwtCustomClaim{
		UserID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(serverConfiguration.ExpiresHour)).Unix(),
			Issuer:    serverConfiguration.Name,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	t, err := token.SignedString([]byte(serverConfiguration.Secret))
	if err != nil {
		return err.Error(), err
	}
	return t, nil
}

func (helper *jwtCryptoHelper) GenerateTokenMitra(MitraId string) (string, error) {
	serverConfiguration := config.GetConfig().Server
	claims := &jwtCustomClaimMitra{
		MitraId,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(serverConfiguration.ExpiresHour)).Unix(),
			Issuer:    serverConfiguration.Name,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	t, err := token.SignedString([]byte(serverConfiguration.Secret2))
	if err != nil {
		return err.Error(), err
	}
	return t, nil
}

// Func to validate token
func (helper *jwtCryptoHelper) ValidateToken(tokenString string) (*jwt.Token, error) {
	serverConfiguration := config.GetConfig().Server
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error")
		}
		return []byte(serverConfiguration.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (helper *jwtCryptoHelper) ValidateTokenMitra(tokenString string) (*jwt.Token, error) {
	serverConfiguration := config.GetConfig().Server
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error")
		}
		return []byte(serverConfiguration.Secret2), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Func to get user id from token
func (helper *jwtCryptoHelper) GetUserIDFromToken(tokenString string) (string, error) {
	token, err := helper.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("token is invalid")
	}
	return claim["user_id"].(string), nil
}
