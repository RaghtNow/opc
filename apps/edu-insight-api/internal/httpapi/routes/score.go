package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
	domainscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
)

func RegisterScoreRoutes(router *gin.RouterGroup) {
	service := appscore.NewService()

	router.GET("/exams", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"items": service.ListExams(),
		})
	})

	router.GET("/exams/:id", func(c *gin.Context) {
		detail, ok := service.GetExamDetail(c.Param("id"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "exam not found"})
			return
		}
		c.JSON(http.StatusOK, detail)
	})

	router.POST("/exams/import", func(c *gin.Context) {
		var req domainscore.ImportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, service.ImportExam(req))
	})

	router.PATCH("/exams/:id/scores/:scoreId", func(c *gin.Context) {
		var req domainscore.UpdateScoreRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}

		detail, ok := service.UpdateScore(c.Param("id"), c.Param("scoreId"), req)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "score not found"})
			return
		}

		c.JSON(http.StatusOK, detail)
	})
}
