package api

import (
	"caddy-proxy-manager/internal/auth"
	"caddy-proxy-manager/internal/caddy"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"caddy-proxy-manager/internal/websocket"

	"github.com/gin-gonic/gin"

	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		if err := db.DB.Exec("SELECT 1").Error; err != nil {
			health["status"] = "unhealthy"
			health["database"] = "error"
			c.JSON(http.StatusServiceUnavailable, health)
			return
		}
		health["database"] = "ok"

		resp, err := http.Get("http://localhost:2019/config/")
		if err != nil || resp.StatusCode != http.StatusOK {
			health["caddy"] = "error"
		} else {
			health["caddy"] = "ok"
			resp.Body.Close()
		}

		c.JSON(http.StatusOK, health)
	})

	api.POST("/auth/login", func(c *gin.Context) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		tokenPair, err := auth.LoginWithTokens(creds.Username, creds.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		setRefreshTokenCookie(c, tokenPair.RefreshToken)

		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenPair.AccessToken,
			"expires_in":   tokenPair.ExpiresIn,
			"token_type":   "Bearer",
		})
	})
	api.POST("/auth/refresh", func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
			return
		}

		tokenPair, err := auth.RefreshTokens(refreshToken)
		if err != nil {
			clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		setRefreshTokenCookie(c, tokenPair.RefreshToken)

		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenPair.AccessToken,
			"expires_in":   tokenPair.ExpiresIn,
			"token_type":   "Bearer",
		})
	})
	api.POST("/auth/logout", func(c *gin.Context) {
		refreshToken, err := c.Cookie("refresh_token")
		if err == nil && refreshToken != "" {
			auth.Logout(refreshToken)
		}
		clearRefreshTokenCookie(c)

		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			CheckAndBroadcastSSLStatus()
		}
	}()
	go func() {
		time.Sleep(1 * time.Minute)
		CheckAndAutoRenewCertificates()

		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			CheckAndAutoRenewCertificates()
		}
	}()

	api.GET("/ws", func(c *gin.Context) {
		websocket.HandleWS(c)
	})

	api.GET("/setup", CheckSetup)
	api.POST("/setup", SetupFirstUser)

	authorized := api.Group("/")
	authorized.Use(AuthMiddleware())
	{
		authorized.GET("/users", GetUsers)
		authorized.POST("/users", CreateUser)
		authorized.PUT("/users/:id", UpdateUser)
		authorized.DELETE("/users/:id", DeleteUser)

		authorized.GET("/profile", GetProfile)
		authorized.PUT("/profile", UpdateProfile)

		authorized.GET("/roles", GetRoles)
		authorized.POST("/roles", CreateRole)
		authorized.PUT("/roles/:id", UpdateRole)
		authorized.DELETE("/roles/:id", DeleteRole)
		authorized.GET("/permissions", GetPermissions)

		authorized.POST("/auth/change-password", func(c *gin.Context) {
			var body struct {
				OldPassword string `json:"old_password"`
				NewPassword string `json:"new_password"`
			}
			if err := c.BindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			if len(body.NewPassword) < 8 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
				return
			}
			userID, exists := c.Get("userID")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}

			var user models.User
			if err := db.DB.First(&user, userID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
				return
			}
			if strings.HasPrefix(user.Password, "$argon2id$") {
				valid, err := auth.VerifyPassword(body.OldPassword, user.Password)
				if err != nil || !valid {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
					return
				}
			} else {
				if user.Password != body.OldPassword {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
					return
				}
			}
			hashedPassword, err := auth.HashPassword(body.NewPassword)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
				return
			}

			user.Password = hashedPassword
			db.DB.Save(&user)

			logAudit(c, "change_password", "User changed password")

			c.JSON(http.StatusOK, gin.H{"message": "password updated"})
		})

		authorized.GET("/hosts", func(c *gin.Context) {
			var hosts []models.Host
			db.DB.Preload("Locations.Upstreams").Preload("Upstreams").Find(&hosts)
			for i, host := range hosts {
				if host.SSLActualProvider == "selfsigned" && host.SSLStatus == "ready" {
					continue
				}
				if host.SSLStatus == "ready" && host.SSLActualProvider == "" && host.SSL && (host.SSLProvider == "letsencrypt" || host.SSLProvider == "zerossl" || host.SSLProvider == "auto") {
					sslReady, actualProvider := caddy.CheckSSLStatusWithProvider(host.Domain, host.SSLProvider)
					if sslReady && actualProvider != "" {
						host.SSLActualProvider = actualProvider
						updates := map[string]interface{}{
							"ssl_actual_provider": actualProvider,
						}
						if host.CertificateID == nil && (actualProvider == "letsencrypt" || actualProvider == "zerossl") {
							cert, err := CreateACMECertificateRecord(host.Domain, actualProvider)
							if err == nil {
								host.CertificateID = &cert.ID
								hosts[i].CertificateID = &cert.ID
								updates["certificate_id"] = cert.ID
								websocket.BroadcastMessage("cert_created", cert)
							}
						}
						db.DB.Model(&host).Updates(updates)
						hosts[i].SSLActualProvider = actualProvider
						websocket.BroadcastMessage("host_updated", hosts[i])
					}
				}
				if host.SSLStatus == "generating" && (host.SSLProvider == "letsencrypt" || host.SSLProvider == "zerossl" || host.SSLProvider == "auto") {
					sslReady, actualProvider := caddy.CheckSSLStatusWithProvider(host.Domain, host.SSLProvider)
					if sslReady {
						host.SSLStatus = "ready"
						host.SSLError = ""
						host.SSLActualProvider = actualProvider
						updates := map[string]interface{}{
							"ssl_status":          "ready",
							"ssl_error":           "",
							"ssl_actual_provider": actualProvider,
						}
						if host.CertificateID == nil && (actualProvider == "letsencrypt" || actualProvider == "zerossl") {
							cert, err := CreateACMECertificateRecord(host.Domain, actualProvider)
							if err == nil {
								host.CertificateID = &cert.ID
								hosts[i].CertificateID = &cert.ID
								updates["certificate_id"] = cert.ID
								websocket.BroadcastMessage("cert_created", cert)
							}
						}
						db.DB.Model(&host).Updates(updates)
						hosts[i].SSLStatus = "ready"
						hosts[i].SSLActualProvider = actualProvider
						websocket.BroadcastMessage("host_updated", hosts[i])
					} else {
						errMsg, found := caddy.GetSSLError(host.Domain, host.CreatedAt)
						if found || time.Since(host.UpdatedAt) > 5*time.Minute {
							if host.SSLProvider == "auto" {
								fallbackErr := fallbackToSelfSigned(&host)
								if fallbackErr == nil {
									host.SSLStatus = "ready"
									host.SSLActualProvider = "selfsigned"
									if errMsg != "" {
										host.SSLError = errMsg + " - Using self-signed certificate"
									} else {
										host.SSLError = "ACME timed out - Using self-signed certificate"
									}
									db.DB.Model(&host).Updates(map[string]interface{}{
										"ssl_status":          "ready",
										"ssl_error":           host.SSLError,
										"ssl_actual_provider": "selfsigned",
										"certificate_id":      host.CertificateID,
									})
									hosts[i].SSLStatus = "ready"
									hosts[i].SSLError = host.SSLError
									hosts[i].SSLActualProvider = "selfsigned"
									hosts[i].CertificateID = host.CertificateID
									go caddy.ApplyConfig()
									websocket.BroadcastMessage("host_updated", hosts[i])
								} else {
									host.SSLStatus = "failed"
									host.SSLError = fmt.Sprintf("ACME failed: %s, Self-signed fallback also failed: %s", errMsg, fallbackErr.Error())
									db.DB.Model(&host).Updates(map[string]interface{}{
										"ssl_status": "failed",
										"ssl_error":  host.SSLError,
									})
									hosts[i].SSLStatus = "failed"
									hosts[i].SSLError = host.SSLError
									websocket.BroadcastMessage("host_updated", hosts[i])
								}
							} else {
								host.SSLStatus = "failed"
								if errMsg != "" {
									host.SSLError = errMsg
								} else {
									host.SSLError = "SSL certificate generation timed out"
								}
								db.DB.Model(&host).Updates(map[string]interface{}{
									"ssl_status": "failed",
									"ssl_error":  host.SSLError,
								})
								hosts[i].SSLStatus = "failed"
								hosts[i].SSLError = host.SSLError
								websocket.BroadcastMessage("host_updated", hosts[i])
							}
						}
					}
				}
			}

			c.JSON(http.StatusOK, hosts)
		})

		authorized.POST("/hosts", func(c *gin.Context) {
			var input struct {
				models.Host
				Locations []models.Location `json:"locations"`
				Upstreams []models.Upstream `json:"upstreams"`
			}
			if err := c.BindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			host := input.Host
			for i := range input.Locations {
				input.Locations[i].ID = 0
				for j := range input.Locations[i].Upstreams {
					input.Locations[i].Upstreams[j].ID = 0
					input.Locations[i].Upstreams[j].HostID = nil
					input.Locations[i].Upstreams[j].LocationID = nil
				}
			}
			host.Locations = input.Locations
			for i := range input.Upstreams {
				input.Upstreams[i].ID = 0
				input.Upstreams[i].LocationID = nil
			}
			host.Upstreams = input.Upstreams
			if host.SSLProvider == "selfsigned" && host.CertificateID == nil {
				uploadDir := "/data/custom_ssl"
				certPath, keyPath, err := GenerateSelfSignedCert(host.Domain, uploadDir)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate self-signed cert: " + err.Error()})
					return
				}

				expiry, _ := GetCertExpiry(certPath)

				cert := models.Certificate{
					Domain:    host.Domain,
					CertFile:  certPath,
					KeyFile:   keyPath,
					Provider:  "selfsigned",
					ExpiresAt: expiry,
					CreatedAt: time.Now(),
				}
				if err := db.DB.Create(&cert).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save cert record: " + err.Error()})
					return
				}
				host.CertificateID = &cert.ID
				host.SSLStatus = "ready"
				host.SSLActualProvider = "selfsigned"
			} else if host.SSL && (host.SSLProvider == "letsencrypt" || host.SSLProvider == "zerossl" || host.SSLProvider == "auto") {
				host.SSLStatus = "generating"
				host.SSLActualProvider = ""
			} else if host.SSLProvider == "custom" {
				host.SSLStatus = "ready"
				host.SSLActualProvider = "custom"
			} else {
				host.SSLStatus = "ready"
				host.SSLActualProvider = ""
			}

			if err := db.DB.Create(&host).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			go func(h models.Host) {
				if err := caddy.ApplyConfig(); err != nil {
					h.SSLStatus = "failed"
					h.SSLError = err.Error()
					db.DB.Save(&h)
					websocket.BroadcastMessage("host_updated", h)
					return
				}
				if h.SSL && (h.SSLProvider == "letsencrypt" || h.SSLProvider == "zerossl" || h.SSLProvider == "auto") {
					time.Sleep(5 * time.Second)
					CheckAndBroadcastSSLStatus()
				}
			}(host)

			targetInfo := host.Target
			if len(host.Upstreams) > 0 {
				targetInfo = fmt.Sprintf("%d upstreams (LB: %s)", len(host.Upstreams), host.LoadBalancing)
			}
			logAudit(c, "create_host", fmt.Sprintf("Created host: %s (Target: %s, Type: %s)", host.Domain, targetInfo, host.Type))

			websocket.BroadcastMessage("host_created", host)
			c.JSON(http.StatusOK, host)
		})

		authorized.PUT("/hosts/:id", func(c *gin.Context) {
			id := c.Param("id")
			var host models.Host
			if err := db.DB.Preload("Locations").Preload("Upstreams").First(&host, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
				return
			}

			var input struct {
				models.Host
				Locations []models.Location `json:"locations"`
				Upstreams []models.Upstream `json:"upstreams"`
			}
			if err := c.BindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			host.Domain = input.Domain
			host.Target = input.Target
			host.Type = input.Type
			host.SSL = input.SSL
			host.SSLProvider = input.SSLProvider
			host.HSTSEnabled = input.HSTSEnabled
			host.BlockExploits = input.BlockExploits
			host.CacheAssets = input.CacheAssets
			host.Websockets = input.Websockets
			host.ForwardingCode = input.ForwardingCode
			host.UseDNSChallenge = input.UseDNSChallenge
			host.DNSProvider = input.DNSProvider
			host.DNSToken = input.DNSToken
			host.AccessListID = input.AccessListID
			host.CertificateID = input.CertificateID
			host.IsActive = input.IsActive
			host.LoadBalancing = input.LoadBalancing
			host.LBTryDuration = input.LBTryDuration
			host.LBTryInterval = input.LBTryInterval
			host.HealthCheck = input.HealthCheck
			host.HealthCheckPath = input.HealthCheckPath
			host.HealthCheckInterval = input.HealthCheckInterval
			var oldLocations []models.Location
			db.DB.Where("host_id = ?", host.ID).Find(&oldLocations)
			for _, loc := range oldLocations {
				db.DB.Delete(&models.Upstream{}, "location_id = ?", loc.ID)
			}
			db.DB.Delete(&models.Location{}, "host_id = ?", host.ID)
			for i := range input.Locations {
				input.Locations[i].HostID = host.ID
				input.Locations[i].ID = 0
				for j := range input.Locations[i].Upstreams {
					input.Locations[i].Upstreams[j].ID = 0
					input.Locations[i].Upstreams[j].HostID = nil
					input.Locations[i].Upstreams[j].LocationID = nil
				}
			}
			host.Locations = input.Locations
			db.DB.Delete(&models.Upstream{}, "host_id = ?", host.ID)
			for i := range input.Upstreams {
				input.Upstreams[i].ID = 0
				input.Upstreams[i].HostID = &host.ID
				input.Upstreams[i].LocationID = nil
			}
			host.Upstreams = input.Upstreams
			if host.SSLProvider == "selfsigned" && host.CertificateID == nil {
				uploadDir := "/data/custom_ssl"
				certPath, keyPath, err := GenerateSelfSignedCert(host.Domain, uploadDir)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate self-signed cert: " + err.Error()})
					return
				}

				expiry, _ := GetCertExpiry(certPath)

				cert := models.Certificate{
					Domain:    host.Domain,
					CertFile:  certPath,
					KeyFile:   keyPath,
					Provider:  "selfsigned",
					ExpiresAt: expiry,
					CreatedAt: time.Now(),
				}
				if err := db.DB.Create(&cert).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save cert record: " + err.Error()})
					return
				}
				host.CertificateID = &cert.ID
				host.SSLStatus = "ready"
				host.SSLActualProvider = "selfsigned"
			}
			if host.SSL != input.SSL || host.SSLProvider != input.SSLProvider || host.Domain != input.Domain {
				if input.SSLProvider == "selfsigned" {
					host.SSLActualProvider = "selfsigned"
				} else if input.SSL && (input.SSLProvider == "letsencrypt" || input.SSLProvider == "zerossl" || input.SSLProvider == "auto") {
					host.SSLStatus = "generating"
					host.SSLActualProvider = ""
				} else if input.SSLProvider == "custom" {
					host.SSLStatus = "ready"
					host.SSLActualProvider = "custom"
				} else {
					host.SSLStatus = "ready"
					host.SSLActualProvider = ""
				}
			}

			if err := db.DB.Save(&host).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			go func(h models.Host) {
				if err := caddy.ApplyConfig(); err != nil {
					h.SSLStatus = "failed"
					h.SSLError = err.Error()
					db.DB.Save(&h)
					websocket.BroadcastMessage("host_updated", h)
					return
				}
				if h.SSL && (h.SSLProvider == "letsencrypt" || h.SSLProvider == "zerossl" || h.SSLProvider == "auto") {
					time.Sleep(5 * time.Second)
					CheckAndBroadcastSSLStatus()
				}
			}(host)

			var changes []string
			if host.Target != input.Target {
				changes = append(changes, fmt.Sprintf("Target: %s -> %s", host.Target, input.Target))
			}
			if host.Type != input.Type {
				changes = append(changes, fmt.Sprintf("Type: %s -> %s", host.Type, input.Type))
			}
			if host.SSL != input.SSL {
				changes = append(changes, fmt.Sprintf("SSL: %v -> %v", host.SSL, input.SSL))
			}
			if host.IsActive != input.IsActive {
				changes = append(changes, fmt.Sprintf("Active: %v -> %v", host.IsActive, input.IsActive))
			}

			details := fmt.Sprintf("Updated host: %s", host.Domain)
			if len(changes) > 0 {
				details += ". Changes: " + strings.Join(changes, ", ")
			}

			logAudit(c, "update_host", details)

			websocket.BroadcastMessage("host_updated", host)
			c.JSON(http.StatusOK, host)
		})

		authorized.DELETE("/hosts/:id", func(c *gin.Context) {
			id := c.Param("id")
			var host models.Host
			if err := db.DB.Unscoped().First(&host, id).Error; err != nil {
			}
			if err := db.DB.Unscoped().Delete(&models.Host{}, id).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			go caddy.ApplyConfig()

			logAudit(c, "delete_host", "Deleted host: "+host.Domain)

			websocket.BroadcastMessage("host_deleted", gin.H{"id": id})
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})

		authorized.POST("/hosts/:id/retry-ssl", func(c *gin.Context) {
			id := c.Param("id")
			var host models.Host
			if err := db.DB.First(&host, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "host not found"})
				return
			}

			if !host.SSL {
				c.JSON(http.StatusBadRequest, gin.H{"error": "SSL is not enabled for this host"})
				return
			}

			if host.SSLProvider != "letsencrypt" && host.SSLProvider != "zerossl" && host.SSLProvider != "auto" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Retry is only available for Let's Encrypt, ZeroSSL, or Auto SSL providers"})
				return
			}

			host.SSLStatus = "generating"
			host.SSLError = ""
			host.SSLActualProvider = ""
			host.UpdatedAt = time.Now()

			if err := db.DB.Save(&host).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			websocket.BroadcastMessage("host_updated", host)

			go func(h models.Host) {
				if err := caddy.ApplyConfig(); err != nil {
					h.SSLStatus = "failed"
					h.SSLError = err.Error()
					db.DB.Save(&h)
					websocket.BroadcastMessage("host_updated", h)
				}
				time.Sleep(5 * time.Second)
				CheckAndBroadcastSSLStatus()
			}(host)

			logAudit(c, "retry_ssl", fmt.Sprintf("Retrying SSL certificate generation for: %s", host.Domain))

			c.JSON(http.StatusOK, gin.H{"message": "SSL certificate generation restarted", "host": host})
		})

		authorized.POST("/caddy/apply", func(c *gin.Context) {
			if err := caddy.ApplyConfig(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "config applied"})
		})

		authorized.GET("/ssl/cert-info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "SSL Management is handled automatically by Caddy. Check Caddy logs for details.",
				"status":  "active",
			})
		})

		authorized.GET("/certificates", func(c *gin.Context) {
			var certs []models.Certificate
			db.DB.Find(&certs)
			for i, cert := range certs {
				if cert.ExpiresAt.IsZero() {
					expiry, err := GetCertExpiry(cert.CertFile)
					if err == nil {
						cert.ExpiresAt = expiry
						db.DB.Model(&cert).Update("expires_at", expiry)
						certs[i].ExpiresAt = expiry
					}
				}
			}

			c.JSON(http.StatusOK, certs)
		})

		authorized.POST("/certificates", func(c *gin.Context) {
			domain := c.PostForm("domain")
			certType := c.PostForm("type")

			if domain == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "domain is required"})
				return
			}
			if certType == "letsencrypt" || certType == "zerossl" {
				email := c.PostForm("email")
				useDNS := c.PostForm("use_dns_challenge") == "true"
				dnsProvider := c.PostForm("dns_provider")
				dnsToken := c.PostForm("dns_token")

				cert := models.Certificate{
					Domain:    domain,
					Provider:  certType,
					Status:    "generating",
					CreatedAt: time.Now(),
				}

				if err := db.DB.Create(&cert).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				go func(certID uint, d, ct, em, dp, dt string, ud bool) {
					cPath, kPath, err := GenerateACMECert(d, ct, em, ud, dp, dt)
					var c models.Certificate
					if dbErr := db.DB.First(&c, certID).Error; dbErr != nil {
						return
					}

					if err != nil {
						c.Status = "failed"
						c.Error = err.Error()
					} else {
						c.CertFile = cPath
						c.KeyFile = kPath
						c.Status = "ready"
						c.Error = ""
						expiry, _ := GetCertExpiry(cPath)
						c.ExpiresAt = expiry
					}
					db.DB.Save(&c)
					caddy.ApplyConfig()
					websocket.BroadcastMessage("cert_updated", c)
				}(cert.ID, domain, certType, email, dnsProvider, dnsToken, useDNS)

				logAudit(c, "upload_cert", fmt.Sprintf("Requested certificate for: %s (Type: %s)", domain, certType))
				c.JSON(http.StatusOK, cert)
				return
			}
			uploadDir := "/data/custom_ssl"
			var certPath, keyPath string
			var err error

			if certType == "selfsigned" {
				certPath, keyPath, err = GenerateSelfSignedCert(domain, uploadDir)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate cert: " + err.Error()})
					return
				}
			} else {
				certFile, err := c.FormFile("cert_file")
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "cert_file is required"})
					return
				}
				keyFile, err := c.FormFile("key_file")
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "key_file is required"})
					return
				}
				if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
					os.MkdirAll(uploadDir, 0755)
				}

				certPath = filepath.Join(uploadDir, domain+".crt")
				keyPath = filepath.Join(uploadDir, domain+".key")

				if err := c.SaveUploadedFile(certFile, certPath); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save cert file"})
					return
				}
				if err := c.SaveUploadedFile(keyFile, keyPath); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save key file"})
					return
				}
			}

			expiry, _ := GetCertExpiry(certPath)

			provider := "custom"
			if certType == "selfsigned" {
				provider = "selfsigned"
			}

			cert := models.Certificate{
				Domain:    domain,
				CertFile:  certPath,
				KeyFile:   keyPath,
				Provider:  provider,
				Status:    "ready",
				ExpiresAt: expiry,
				CreatedAt: time.Now(),
			}

			if err := db.DB.Create(&cert).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			go caddy.ApplyConfig()

			logAudit(c, "upload_cert", fmt.Sprintf("Added certificate for: %s (Type: %s)", domain, certType))

			websocket.BroadcastMessage("cert_created", cert)
			c.JSON(http.StatusOK, cert)
		})

		authorized.DELETE("/certificates/:id", func(c *gin.Context) {
			id := c.Param("id")
			var cert models.Certificate
			if err := db.DB.First(&cert, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "certificate not found"})
				return
			}
			var count int64
			db.DB.Model(&models.Host{}).Where("certificate_id = ?", id).Count(&count)
			if count > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Cannot delete certificate: in use by %d hosts", count)})
				return
			}
			os.Remove(cert.CertFile)
			os.Remove(cert.KeyFile)

			db.DB.Delete(&cert)
			go caddy.ApplyConfig()

			logAudit(c, "delete_cert", "Deleted certificate: "+cert.Domain)

			websocket.BroadcastMessage("cert_deleted", gin.H{"id": id})
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})

		authorized.PUT("/certificates/:id/auto-renew", func(c *gin.Context) {
			id := c.Param("id")
			var body struct {
				AutoRenew bool `json:"auto_renew"`
			}
			if err := c.BindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			var cert models.Certificate
			if err := db.DB.First(&cert, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "certificate not found"})
				return
			}

			cert.AutoRenew = body.AutoRenew
			db.DB.Save(&cert)

			status := "disabled"
			if body.AutoRenew {
				status = "enabled"
			}
			logAudit(c, "toggle_auto_renew", fmt.Sprintf("Auto-renew %s for certificate: %s", status, cert.Domain))

			websocket.BroadcastMessage("cert_updated", cert)
			c.JSON(http.StatusOK, cert)
		})

		authorized.GET("/certificates/:id/download", func(c *gin.Context) {
			id := c.Param("id")
			fileType := c.Query("type")

			var cert models.Certificate
			if err := db.DB.First(&cert, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "certificate not found"})
				return
			}

			var filePath string
			if fileType == "key" {
				filePath = cert.KeyFile
			} else {
				filePath = cert.CertFile
			}
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
				return
			}

			c.File(filePath)
		})

		RegisterAccessListRoutes(authorized)

		authorized.GET("/settings/default-page", GetDefaultPage)
		authorized.PUT("/settings/default-page", UpdateDefaultPage)
		authorized.GET("/settings/zerossl-eab", GetZeroSSLEAB)
		authorized.PUT("/settings/zerossl-eab", UpdateZeroSSLEAB)

		authorized.GET("/audit-logs", GetAuditLogs)
		authorized.GET("/streams", GetStreams)
		authorized.POST("/streams", CreateStream)
		authorized.PUT("/streams/:id", UpdateStream)
		authorized.DELETE("/streams/:id", DeleteStream)
	}
}
func GetDefaultPage(c *gin.Context) {
	var setting models.Setting
	if err := db.DB.Where("key = ?", "default_page_html").First(&setting).Error; err != nil {
		content, err := os.ReadFile("/data/default/index.html")
		if err != nil {
			content, err = os.ReadFile("/app/default/index.html")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read default page"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"html": string(content)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"html": setting.Value})
}
func UpdateDefaultPage(c *gin.Context) {
	var input struct {
		HTML string `json:"html" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setting := models.Setting{
		Key:   "default_page_html",
		Value: input.HTML,
	}
	if err := db.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
		return
	}
	if err := os.MkdirAll("/data/default", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
		return
	}

	if err := os.WriteFile("/data/default/index.html", []byte(input.HTML), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}

	logAudit(c, "update_default_page", "Updated default page HTML")

	c.JSON(http.StatusOK, gin.H{"message": "default page updated"})
}

func GetZeroSSLEAB(c *gin.Context) {
	var kidSetting, hmacSetting models.Setting
	kid := ""
	hmacKey := ""
	if err := db.DB.Where("key = ?", "zerossl_eab_kid").First(&kidSetting).Error; err == nil {
		kid = kidSetting.Value
	}
	if err := db.DB.Where("key = ?", "zerossl_eab_hmac_key").First(&hmacSetting).Error; err == nil {
		if len(hmacSetting.Value) > 8 {
			hmacKey = hmacSetting.Value[:4] + "****" + hmacSetting.Value[len(hmacSetting.Value)-4:]
		} else if hmacSetting.Value != "" {
			hmacKey = "****"
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"kid":        kid,
		"hmac_key":   hmacKey,
		"configured": kid != "" && hmacSetting.Value != "",
	})
}

func UpdateZeroSSLEAB(c *gin.Context) {
	var input struct {
		KID     string `json:"kid"`
		HMACKey string `json:"hmac_key"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kidSetting := models.Setting{Key: "zerossl_eab_kid", Value: input.KID}
	if err := db.DB.Where("key = ?", "zerossl_eab_kid").Assign(kidSetting).FirstOrCreate(&kidSetting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save EAB KID"})
		return
	}

	hmacSetting := models.Setting{Key: "zerossl_eab_hmac_key", Value: input.HMACKey}
	if err := db.DB.Where("key = ?", "zerossl_eab_hmac_key").Assign(hmacSetting).FirstOrCreate(&hmacSetting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save EAB HMAC Key"})
		return
	}

	logAudit(c, "update_zerossl_eab", "Updated ZeroSSL EAB credentials")
	c.JSON(http.StatusOK, gin.H{"message": "ZeroSSL EAB credentials updated"})
}

func CheckAndAutoRenewCertificates() {
	var certs []models.Certificate
	if err := db.DB.Where("auto_renew = ?", true).Find(&certs).Error; err != nil {
		return
	}
	renewThreshold := time.Now().Add(30 * 24 * time.Hour)

	for _, cert := range certs {
		expiresAt, err := GetCertExpiry(cert.CertFile)
		if err != nil {
			continue
		}
		if !expiresAt.Equal(cert.ExpiresAt) {
			cert.ExpiresAt = expiresAt
			db.DB.Save(&cert)
		}
		if expiresAt.Before(renewThreshold) {
			if strings.Contains(cert.CertFile, "selfsigned") {
				newCertPath, newKeyPath, err := GenerateSelfSignedCert(cert.Domain, "/data/custom_ssl")
				if err != nil {
					cert.Error = "Auto-renew failed: " + err.Error()
					db.DB.Save(&cert)
					continue
				}
				newExpiry, _ := GetCertExpiry(newCertPath)
				cert.CertFile = newCertPath
				cert.KeyFile = newKeyPath
				cert.ExpiresAt = newExpiry
				cert.Status = "ready"
				cert.Error = ""
				db.DB.Save(&cert)
				websocket.BroadcastMessage("certificate_renewed", cert)
			} else {
				caddy.ApplyConfig()
				newExpiry, err := GetCertExpiry(cert.CertFile)
				if err == nil && !newExpiry.Equal(cert.ExpiresAt) {
					cert.ExpiresAt = newExpiry
					cert.Status = "ready"
					cert.Error = ""
					db.DB.Save(&cert)
					websocket.BroadcastMessage("certificate_renewed", cert)
				}
			}
		}
	}
	var hosts []models.Host
	if err := db.DB.Where("ssl = ? AND ssl_status = ?", true, "ready").Find(&hosts).Error; err != nil {
		return
	}

	for _, host := range hosts {
		if !host.SSL {
			continue
		}
		if host.SSLProvider == "letsencrypt" || host.SSLProvider == "zerossl" || host.SSLProvider == "auto" {
			continue
		}
		if host.CertificateID != nil {
			var cert models.Certificate
			if err := db.DB.First(&cert, *host.CertificateID).Error; err == nil {
				if cert.AutoRenew && strings.Contains(cert.CertFile, "selfsigned") {
					expiresAt, _ := GetCertExpiry(cert.CertFile)
					if expiresAt.Before(renewThreshold) {
						newCertPath, newKeyPath, err := GenerateSelfSignedCert(cert.Domain, "/data/custom_ssl")
						if err == nil {
							newExpiry, _ := GetCertExpiry(newCertPath)
							cert.CertFile = newCertPath
							cert.KeyFile = newKeyPath
							cert.ExpiresAt = newExpiry
							cert.Status = "ready"
							cert.Error = ""
							db.DB.Save(&cert)
							caddy.ApplyConfig()
							websocket.BroadcastMessage("certificate_renewed", cert)
						}
					}
				}
			}
		}
	}
}
func CheckAndBroadcastSSLStatus() {
	var hosts []models.Host
	if err := db.DB.Where("ssl_status = ? AND ssl_provider IN ?", "generating", []string{"letsencrypt", "zerossl", "auto"}).Find(&hosts).Error; err != nil {
		return
	}

	for _, host := range hosts {
		changed := false
		sslReady, actualProvider := caddy.CheckSSLStatusWithProvider(host.Domain, host.SSLProvider)
		if sslReady {
			host.SSLStatus = "ready"
			host.SSLError = ""
			host.SSLActualProvider = actualProvider
			changed = true

			if host.CertificateID == nil && (actualProvider == "letsencrypt" || actualProvider == "zerossl") {
				cert, err := CreateACMECertificateRecord(host.Domain, actualProvider)
				if err == nil {
					host.CertificateID = &cert.ID
					websocket.BroadcastMessage("cert_created", cert)
				}
			}
		} else {
			errMsg, found := caddy.GetSSLError(host.Domain, host.CreatedAt)
			if found || time.Since(host.UpdatedAt) > 5*time.Minute {
				if host.SSLProvider == "auto" {
					fallbackErr := fallbackToSelfSigned(&host)
					if fallbackErr == nil {
						host.SSLStatus = "ready"
						host.SSLActualProvider = "selfsigned"
						if errMsg != "" {
							host.SSLError = errMsg + " - Using self-signed certificate"
						} else {
							host.SSLError = "ACME timed out - Using self-signed certificate"
						}
						go caddy.ApplyConfig()
						changed = true
					} else {
						host.SSLStatus = "failed"
						host.SSLError = fmt.Sprintf("ACME failed: %s, Self-signed fallback also failed: %s", errMsg, fallbackErr.Error())
						changed = true
					}
				} else {
					host.SSLStatus = "failed"
					if errMsg != "" {
						host.SSLError = errMsg
					} else {
						host.SSLError = "SSL generation timed out. Check logs."
					}
					changed = true
				}
			}
		}

		if changed {
			db.DB.Save(&host)
			websocket.BroadcastMessage("host_updated", host)
		}
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
func setRefreshTokenCookie(c *gin.Context, token string) {
	maxAge := 7 * 24 * 60 * 60

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		token,
		maxAge,
		"/api/auth",
		"",
		false,
		true,
	)
}
func clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/api/auth",
		"",
		false,
		true,
	)
}
