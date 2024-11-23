package api

import (
	"github.com/AbderraoufKhorchani/file-saver/broker-service/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)


func Routes() *gin.Engine {

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}
	config.ExposeHeaders = []string{"Link"}
	config.AllowCredentials = true
	config.MaxAge = 300
	r.Use(cors.New(config))

	r.POST("submit", handlers.MainHandler)
	return r
}
