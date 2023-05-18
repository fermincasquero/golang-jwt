package main

import (
	"fmt"

	"github.com/fredele20/golang-jwt-project/config"
	"github.com/fredele20/golang-jwt-project/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	secrets := config.GetSecrets()

	address := fmt.Sprintf("127.0.0.1:%s", secrets.Port)

	router := gin.New()
	router.Use(gin.Logger())

	routes.ProductRoutes(router)
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	// routes.ProductProtectedRoutes(router)

	router.GET("/api-1", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": "Access granted for api-1"})
	})
	router.GET("/api-2", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	router.Run(address)

}
