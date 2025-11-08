package middleware

import (
	"errors"
	"strings"

	"GoGin/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			util.Error(c, 401, "Authorization header is empty")
			c.Abort()
			return
		}

		parts := strings.SplitN(authorizationHeader, " ", 2)
		if len(parts) != 2 && parts[0] != "Bearer" {
			util.Error(c, 401, "Authorization header is invalid")
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := util.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				util.Error(c, 401, "Token is expired")
				c.Abort()
				return
			}
			util.Error(c, 401, "Token is invalid")
			c.Abort()
			return
		}

		claims, err := util.ExtractClaims(token)
		if err != nil {
			util.Error(c, 500, "Failed to extract claims")
			c.Abort()
			return
		}

		c.Set("username", claims["username"])
		c.Set("user_id", claims["user_id"])

		c.Next()
	}
}
