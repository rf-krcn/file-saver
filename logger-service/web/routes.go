package api

import (
	"github.com/AbderraoufKhorchani/file-saver/logger-service/internal/helpers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Routes defines the API routes and returns a Gin engine.
func Routes() *gin.Engine {

	// Create a new Gin router
	r := gin.Default()

	// Enable CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"POST"}
	config.AllowHeaders = []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}
	config.ExposeHeaders = []string{"Link"}
	config.AllowCredentials = true
	config.MaxAge = 300
	r.Use(cors.New(config))

	// Define the "/log" route
	r.POST("/log", helpers.WriteLog)
	return r
}
