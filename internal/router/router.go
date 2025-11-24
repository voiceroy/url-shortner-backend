package router

import (
	"net/http"

	"go-shorten/internal/handler"
	"go-shorten/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))
	r.Use(gin.Recovery())

	r.TrustedPlatform = gin.PlatformFlyIO

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/shorten", middleware.RateLimit(), handler.ShortenURLHandler)
	r.GET("/:code", handler.RetrieveMappingHandler)

	return r
}
