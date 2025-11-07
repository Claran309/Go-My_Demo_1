package middleware

import (
	"errors"
	"net/http"
	"strings"

	"GoGin/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Authorization header is empty",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authorizationHeader, " ", 2)
		if len(parts) != 2 && parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Authorization header is invalid",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := util.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  http.StatusUnauthorized,
					"message": "Token is expired",
				})
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Token is invalid",
			})
			c.Abort()
			return
		}

		claims, err := util.ExtractClaims(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to extract claims",
			})
			c.Abort()
			return
		}

		c.Set("username", claims["username"])
		c.Set("user_id", claims["user_id"])

		c.Next()
	}
}
