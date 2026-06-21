package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	appscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/score"
	domainscore "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/score"
)

func RegisterScoreRoutes(router *gin.RouterGroup, service appscore.Service, classroomService appclassroom.Service) {
	router.GET("/exams", func(c *gin.Context) {
		items, err := service.ListExams(classIDParam(c))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "list exams failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"items": items,
		})
	})

	router.GET("/exams/:id", func(c *gin.Context) {
		detail, ok, err := service.GetExamDetail(classIDParam(c), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "get exam detail failed", "error": err.Error()})
			return
		}
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

		detail, err := service.ImportExam(classIDParam(c), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "import exam failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, detail)
	})

	router.POST("/exams/import/validate", func(c *gin.Context) {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "missing score file", "error": err.Error()})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "open score file failed", "error": err.Error()})
			return
		}
		defer file.Close()

		content := make([]byte, fileHeader.Size)
		if _, err := file.Read(content); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "read score file failed", "error": err.Error()})
			return
		}

		subjects := splitSubjects(c.PostForm("subjects"))
		workspace, err := classroomService.GetWorkspace(classIDParam(c))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "load classroom workspace failed", "error": err.Error()})
			return
		}
		result, err := service.ValidateImport(fileHeader.Filename, content, subjects, workspace)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "validate score import failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	router.PATCH("/exams/:id/scores/:scoreId", func(c *gin.Context) {
		var req domainscore.UpdateScoreRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}

		detail, ok, err := service.UpdateScore(classIDParam(c), c.Param("id"), c.Param("scoreId"), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "update score failed", "error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "score not found"})
			return
		}

		c.JSON(http.StatusOK, detail)
	})
}

func splitSubjects(value string) []string {
	parts := strings.Split(value, ",")
	subjects := []string{}
	for _, part := range parts {
		subject := strings.TrimSpace(part)
		if subject != "" {
			subjects = append(subjects, subject)
		}
	}
	return subjects
}
