package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fredele20/golang-jwt-project/helpers"
	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")
		if clientToken == "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			ctx.Abort()
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
			ctx.Abort()
			return
		}

		ctx.Set("email", claims.Email)
		ctx.Set("first_name", claims.First_name)
		ctx.Set("last_name", claims.Last_name)
		ctx.Set("uid", claims.Uid)
		ctx.Set("user_type", claims.User_type)
		ctx.Next()

	}
}

func GetCurrentUser(id string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.GetString("user_id")
		if userId != id {
			err := errors.New("user can not perform this operation")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err})
			return
		}
		ctx.Next()
	}
}
