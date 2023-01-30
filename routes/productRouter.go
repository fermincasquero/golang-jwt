package routes

import (
	"github.com/fredele20/golang-jwt-project/controllers"
	"github.com/fredele20/golang-jwt-project/middleware"
	"github.com/gin-gonic/gin"
)

func ProductRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/product/:product_id", controllers.GetProductById())
	incomingRoutes.GET("/product/owner/:ownerid", controllers.GetProductsByOwnerId())
}

func ProductProtectedRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.POST("/product/add", controllers.AddProduct())
	incomingRoutes.POST("/product/purchase", controllers.PurchaseProduct())
}
