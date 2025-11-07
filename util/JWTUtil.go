package util

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"GoGin/model"
)

func GenerateToken(userID string, username string) (string, error) {
	jwtConfig := model.DefaultJWTConfig

	claims := jwt.MapClaims{
		"username": username,
		"user_id":  userID,
		"iss":      jwtConfig.Issuer,
		"sub":      userID,
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Unix(),
		"exp":      time.Now().Add(jwtConfig.ExpirationTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(model.DefaultJWTConfig.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func ExtractClaims(token *jwt.Token) (jwt.MapClaims, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, jwt.ErrTokenInvalidClaims
	}
}
