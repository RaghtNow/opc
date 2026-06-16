package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/config"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/httpapi/routes"
)

func RegisterRoutes(engine *gin.Engine, cfg config.Config, classroomService appclassroom.Service, scoreService appscore.Service) {
	engine.Use(corsMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	api := engine.Group("/api")
	{
		routes.RegisterMetaRoutes(api, cfg)
		routes.RegisterClassroomRoutes(api, classroomService)
		routes.RegisterScoreRoutes(api, scoreService, classroomService)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
