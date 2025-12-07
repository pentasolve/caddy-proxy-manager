package api

import (
	"caddy-proxy-manager/internal/caddy"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"caddy-proxy-manager/internal/websocket"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)
func hasProtocolConflict(stream1, stream2 models.Stream) bool {
	if stream1.TCPEnabled && stream2.TCPEnabled {
		return true
	}
	if stream1.UDPEnabled && stream2.UDPEnabled {
		return true
	}
	return false
}
func GetStreams(c *gin.Context) {
	var streams []models.Stream
	if err := db.DB.Find(&streams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, streams)
}
func getProtocolString(stream models.Stream) string {
	if stream.TCPEnabled && stream.UDPEnabled {
		return "tcp+udp"
	} else if stream.TCPEnabled {
		return "tcp"
	} else if stream.UDPEnabled {
		return "udp"
	}
	return "none"
}
func CreateStream(c *gin.Context) {
	var stream models.Stream
	if err := c.BindJSON(&stream); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if stream.ListenPort == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listen_port is required"})
		return
	}
	if stream.Target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target is required"})
		return
	}
	if !stream.TCPEnabled && !stream.UDPEnabled {
		stream.TCPEnabled = true
	}
	if stream.SSL {
		validSSLProviders := map[string]bool{"auto": true, "letsencrypt": true, "zerossl": true, "selfsigned": true, "custom": true}
		if stream.SSLProvider == "" {
			stream.SSLProvider = "auto"
		}
		if !validSSLProviders[stream.SSLProvider] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ssl_provider must be 'auto', 'letsencrypt', 'zerossl', 'selfsigned', or 'custom'"})
			return
		}
		if stream.SSLProvider == "custom" && stream.CertificateID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "certificate_id is required when using custom SSL"})
			return
		}
	}
	var existingStreams []models.Stream
	db.DB.Where("listen_port = ?", stream.ListenPort).Find(&existingStreams)
	for _, existing := range existingStreams {
		if hasProtocolConflict(stream, existing) {
			conflictProto := ""
			if stream.TCPEnabled && existing.TCPEnabled {
				conflictProto = "TCP"
			} else if stream.UDPEnabled && existing.UDPEnabled {
				conflictProto = "UDP"
			}
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Port %d already has conflicting %s configuration", stream.ListenPort, conflictProto)})
			return
		}
	}
	reservedPorts := []int{80, 81, 443, 2019}
	for _, rp := range reservedPorts {
		if stream.ListenPort == rp {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Port %d is reserved for system use", rp)})
			return
		}
	}

	if err := db.DB.Create(&stream).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go func() {
		if err := caddy.ApplyConfig(); err != nil {
			fmt.Printf("Failed to apply config after stream creation: %v\n", err)
		}
	}()

	logAudit(c, "create_stream", fmt.Sprintf("Created stream: %s (Port: %d/%s -> %s)", stream.Name, stream.ListenPort, getProtocolString(stream), stream.Target))

	websocket.BroadcastMessage("stream_created", stream)
	c.JSON(http.StatusOK, stream)
}
func UpdateStream(c *gin.Context) {
	id := c.Param("id")
	var stream models.Stream
	if err := db.DB.First(&stream, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		return
	}

	var input models.Stream
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.ListenPort == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listen_port is required"})
		return
	}
	if input.Target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target is required"})
		return
	}
	if !input.TCPEnabled && !input.UDPEnabled {
		input.TCPEnabled = true
	}
	if input.SSL {
		validSSLProviders := map[string]bool{"auto": true, "letsencrypt": true, "zerossl": true, "selfsigned": true, "custom": true}
		if input.SSLProvider == "" {
			input.SSLProvider = "auto"
		}
		if !validSSLProviders[input.SSLProvider] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ssl_provider must be 'auto', 'letsencrypt', 'zerossl', 'selfsigned', or 'custom'"})
			return
		}
		if input.SSLProvider == "custom" && input.CertificateID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "certificate_id is required when using custom SSL"})
			return
		}
	}
	var existingStreams []models.Stream
	db.DB.Where("listen_port = ? AND id != ?", input.ListenPort, stream.ID).Find(&existingStreams)
	for _, existing := range existingStreams {
		if hasProtocolConflict(input, existing) {
			conflictProto := ""
			if input.TCPEnabled && existing.TCPEnabled {
				conflictProto = "TCP"
			} else if input.UDPEnabled && existing.UDPEnabled {
				conflictProto = "UDP"
			}
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Port %d already has conflicting %s configuration", input.ListenPort, conflictProto)})
			return
		}
	}
	reservedPorts := []int{80, 81, 443, 2019}
	for _, rp := range reservedPorts {
		if input.ListenPort == rp {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Port %d is reserved for system use", rp)})
			return
		}
	}
	stream.Name = input.Name
	stream.ListenPort = input.ListenPort
	stream.TCPEnabled = input.TCPEnabled
	stream.UDPEnabled = input.UDPEnabled
	stream.Target = input.Target
	stream.SSL = input.SSL
	stream.SSLProvider = input.SSLProvider
	stream.CertificateID = input.CertificateID
	stream.IsActive = input.IsActive

	if err := db.DB.Save(&stream).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go func() {
		if err := caddy.ApplyConfig(); err != nil {
			fmt.Printf("Failed to apply config after stream update: %v\n", err)
		}
	}()

	logAudit(c, "update_stream", fmt.Sprintf("Updated stream: %s (Port: %d/%s -> %s)", stream.Name, stream.ListenPort, getProtocolString(stream), stream.Target))

	websocket.BroadcastMessage("stream_updated", stream)
	c.JSON(http.StatusOK, stream)
}
func DeleteStream(c *gin.Context) {
	id := c.Param("id")
	var stream models.Stream
	if err := db.DB.First(&stream, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
		return
	}

	if err := db.DB.Delete(&stream).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go func() {
		if err := caddy.ApplyConfig(); err != nil {
			fmt.Printf("Failed to apply config after stream deletion: %v\n", err)
		}
	}()

	logAudit(c, "delete_stream", fmt.Sprintf("Deleted stream: %s (Port: %d/%s)", stream.Name, stream.ListenPort, getProtocolString(stream)))

	websocket.BroadcastMessage("stream_deleted", gin.H{"id": id})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
