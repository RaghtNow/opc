package routes

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func classIDParam(c *gin.Context) string {
	return c.Query("classId")
}

func scopeParam(c *gin.Context) string {
	return c.Query("scope")
}

func classIDsParam(c *gin.Context) []string {
	raw := c.Query("classIds")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ids := []string{}
	for _, part := range parts {
		id := strings.TrimSpace(part)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
