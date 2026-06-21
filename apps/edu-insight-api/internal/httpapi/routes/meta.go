package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/config"
)

func RegisterMetaRoutes(router *gin.RouterGroup, cfg config.Config) {
	router.GET("/meta", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name": "edu-insight-api",
			"env":  cfg.AppEnv,
		})
	})
}
