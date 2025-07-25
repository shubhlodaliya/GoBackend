package routes

import "github.com/gin-gonic/gin"

func Init(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/auth/login", AuthHandler)
	// api.Use(middleware.AuthMiddleware())

	api.POST("/device/add", AddDevice)
	api.POST("/device/status", UpdateDeviceStatus)
	api.GET("/devices/:user_id", GetDevicesByUserID)
}
