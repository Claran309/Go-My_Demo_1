package model

import (
	"time"
)

type JWTConfig struct {
	Issuer         string
	SecretKey      string
	ExpirationTime time.Duration
}

var DefaultJWTConfig = JWTConfig{
	SecretKey:      "MySuperSecureJWTKey@2025!DoNotShare!",
	Issuer:         "Login_And_Register_Demo",
	ExpirationTime: time.Hour * 24 * 7,
}
