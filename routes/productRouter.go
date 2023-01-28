package routes

import (
	"github.com/fredele20/golang-jwt-project/controllers"
	"github.com/fredele20/golang-jwt-project/middleware"
	"github.com/gin-gonic/gin"
)

func ProductRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.POST("/product/add", controllers.AddProduct())
}
