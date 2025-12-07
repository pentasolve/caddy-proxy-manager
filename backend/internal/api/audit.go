package api

import (
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"caddy-proxy-manager/internal/websocket"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)
func logAudit(c *gin.Context, action string, details string) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	var uid uint
	var uname string

	if id, ok := userID.(uint); ok {
		uid = id
	}
	if name, ok := username.(string); ok {
		uname = name
	} else {
		uname = "system/unknown"
	}

	log := models.AuditLog{
		UserID:    uid,
		Username:  uname,
		Action:    action,
		Details:   details,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		CreatedAt: time.Now(),
	}
	db.DB.Create(&log)
	websocket.BroadcastMessage("audit_log_created", log)
}
func GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	if err := db.DB.Order("created_at desc").Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
