package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "github.com/RaghtNow/opc/apps/edu-insight-api/internal/application/auth"
	domainauth "github.com/RaghtNow/opc/apps/edu-insight-api/internal/domain/auth"
)

func RegisterAuthRoutes(router *gin.RouterGroup, service appauth.Service) {
	router.POST("/auth/sms-code", func(c *gin.Context) {
		var req domainauth.SendSMSCodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		resp, err := service.SendSMSCode(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "send sms code failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	router.POST("/auth/login/sms", func(c *gin.Context) {
		var req domainauth.LoginWithSMSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request", "error": err.Error()})
			return
		}
		resp, err := service.LoginWithSMS(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "login failed", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/auth/me", func(c *gin.Context) {
		me, ok, err := service.CurrentUser(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "get current user failed", "error": err.Error()})
			return
		}
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		c.JSON(http.StatusOK, me)
	})
}
