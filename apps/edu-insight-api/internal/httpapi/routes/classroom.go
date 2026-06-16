package routes

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	appclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/classroom"
	domainclassroom "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/classroom"
)

func RegisterClassroomRoutes(router *gin.RouterGroup, service appclassroom.Service) {
	router.GET("/classes/current", func(c *gin.Context) {
		workspace, err := service.GetWorkspace()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "get classroom workspace failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.POST("/classes/current/students", func(c *gin.Context) {
		var req domainclassroom.SaveStudentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, err := service.CreateStudent(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "create student failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.PATCH("/classes/current/students/:id", func(c *gin.Context) {
		var req domainclassroom.SaveStudentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, ok, err := service.UpdateStudent(c.Param("id"), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "update student failed", "error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "student not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.POST("/classes/current/students/import", func(c *gin.Context) {
		fileName, content, ok := readImportFile(c)
		if !ok {
			return
		}
		workspace, summary, err := service.ImportStudents(fileName, content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "import students failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"workspace": workspace, "summary": summary})
	})

	router.POST("/classes/current/teachers", func(c *gin.Context) {
		var req domainclassroom.SaveTeacherRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, err := service.CreateTeacher(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "create teacher failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.PATCH("/classes/current/teachers/:id", func(c *gin.Context) {
		var req domainclassroom.SaveTeacherRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		workspace, ok, err := service.UpdateTeacher(c.Param("id"), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "update teacher failed", "error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "teacher not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.POST("/classes/current/teachers/import", func(c *gin.Context) {
		fileName, content, ok := readImportFile(c)
		if !ok {
			return
		}
		workspace, summary, err := service.ImportTeachers(fileName, content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "import teachers failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"workspace": workspace, "summary": summary})
	})

	router.POST("/classes/current/teachers/:id/bind-account", func(c *gin.Context) {
		workspace, ok, err := service.BindTeacherAccount(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "bind teacher account failed", "error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "teacher not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})

	router.POST("/classes/current/teachers/:id/sync-permission", func(c *gin.Context) {
		workspace, ok, err := service.SyncTeacherPermission(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "sync teacher permission failed", "error": err.Error()})
			return
		}
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
		workspace, ok, err := service.UpdatePolicy(c.Param("id"), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "update policy failed", "error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "policy not found"})
			return
		}
		c.JSON(http.StatusOK, workspace)
	})
}

func readImportFile(c *gin.Context) (string, []byte, bool) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing import file", "error": err.Error()})
		return "", nil, false
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "open import file failed", "error": err.Error()})
		return "", nil, false
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "read import file failed", "error": err.Error()})
		return "", nil, false
	}
	return fileHeader.Filename, content, true
}
