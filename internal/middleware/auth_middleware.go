package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"habit-streak-graphql/internal/utils"

	"github.com/gin-gonic/gin"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := utils.ValidateJWT(tokenString, os.Getenv("JWT_SECRET"))
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			ctx.Abort()
			return
		}

		c := context.WithValue(ctx.Request.Context(), userIDKey, userID)
		ctx.Request = ctx.Request.WithContext(c)
		ctx.Next()
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
