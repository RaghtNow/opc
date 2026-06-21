package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appinsight "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/insight"
)

func RegisterInsightRoutes(router *gin.RouterGroup, service appinsight.Service) {
	router.GET("/insights/dashboard", func(c *gin.Context) {
		dashboard, err := service.GetDashboard(appinsight.DashboardQuery{
			ClassID:  classIDParam(c),
			Scope:    scopeParam(c),
			ClassIDs: classIDsParam(c),
			ExamID:   c.Query("examId"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "get insight dashboard failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, dashboard)
	})

	router.POST("/sync/publish-latest", func(c *gin.Context) {
		dashboard, err := service.PublishLatestExam(classIDParam(c))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "publish latest exam failed", "error": err.Error(), "dashboard": dashboard})
			return
		}
		c.JSON(http.StatusOK, dashboard)
	})
}
