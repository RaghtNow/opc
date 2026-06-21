package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/auth"
	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appinsight "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/insight"
	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/config"
	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/httpapi/routes"
)

func RegisterRoutes(engine *gin.Engine, cfg config.Config, authService appauth.Service, classroomService appclassroom.Service, scoreService appscore.Service, insightService appinsight.Service) {
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
		routes.RegisterAuthRoutes(api, authService)
		routes.RegisterClassroomRoutes(api, classroomService)
		routes.RegisterScoreRoutes(api, scoreService, classroomService)
		routes.RegisterInsightRoutes(api, insightService)
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
