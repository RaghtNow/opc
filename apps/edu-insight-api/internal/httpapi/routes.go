package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/config"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/httpapi/routes"
)

func RegisterRoutes(engine *gin.Engine, cfg config.Config) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	api := engine.Group("/api")
	{
		routes.RegisterMetaRoutes(api, cfg)
		routes.RegisterScoreRoutes(api)
	}
}
