package routes

import (
	"github.com/fredele20/golang-jwt-project/controllers"
	"github.com/fredele20/golang-jwt-project/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	// incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", middleware.Authenticate(), controllers.GetUsers())
	// incomingRoutes.GET("/users")
	incomingRoutes.GET("/users/:user_id", controllers.GetUserById())
}
