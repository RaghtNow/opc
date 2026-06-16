package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	domainclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
)

func RegisterClassroomRoutes(router *gin.RouterGroup) {
	service := appclassroom.NewService()

	router.GET("/classes/current", func(c *gin.Context) {
		c.JSON(http.StatusOK, service.GetWorkspace())
	})

	router.POST("/classes/current/students", func(c *gin.Context) {
		var req domainclassroom.SaveStudentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, service.CreateStudent(req))
	})

	router.PATCH("/classes/current/students/:id", func(c *gin.Context) {
		var req domainclassroom.SaveStudentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, ok := service.UpdateStudent(c.Param("id"), req)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "student not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.POST("/classes/current/teachers", func(c *gin.Context) {
		var req domainclassroom.SaveTeacherRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, service.CreateTeacher(req))
	})

	router.PATCH("/classes/current/teachers/:id", func(c *gin.Context) {
		var req domainclassroom.SaveTeacherRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, ok := service.UpdateTeacher(c.Param("id"), req)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "teacher not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.PATCH("/classes/current/policies/:id", func(c *gin.Context) {
		var req domainclassroom.SavePolicyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, ok := service.UpdatePolicy(c.Param("id"), req)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "policy not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})
}
