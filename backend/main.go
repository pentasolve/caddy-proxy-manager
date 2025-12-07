package main

import (
	"caddy-proxy-manager/internal/api"
	"caddy-proxy-manager/internal/caddy"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func printBanner() {
	banner := `
   _____ ____  __  __ 
  / ____|  _ \|  \/  |
 | |    | |_) | \  / |
 | |    |  __/| |\/| |
 | |____| |   | |  | |
  \_____|_|   |_|  |_|

  Caddy Proxy Manager
`
	fmt.Print(banner)
}

func main() {
	printBanner()
	db.Init()

	var setting models.Setting
	var htmlContent string

	content, err := os.ReadFile("/app/default/index.html")
	if err == nil {
		htmlContent = string(content)
		var existing models.Setting
		if err := db.DB.Where("key = ?", "default_page_html").First(&existing).Error; err == nil {
			existing.Value = htmlContent
			db.DB.Save(&existing)
		} else {
			db.DB.Create(&models.Setting{Key: "default_page_html", Value: htmlContent})
		}
	} else {
		if err := db.DB.Where("key = ?", "default_page_html").First(&setting).Error; err != nil {
			htmlContent = "<h1>Welcome to Caddy Proxy Manager</h1>"
			db.DB.Create(&models.Setting{Key: "default_page_html", Value: htmlContent})
		} else {
			htmlContent = setting.Value
		}
	}

	os.MkdirAll("/data/default", 0755)
	os.WriteFile("/data/default/index.html", []byte(htmlContent), 0644)

	go func() {
		log.Println("Waiting for Caddy admin API...")
		if err := caddy.WaitForCaddy(30, 1*time.Second); err != nil {
			log.Printf("Warning: Caddy admin API not ready: %v", err)
			return
		}
		log.Println("Caddy admin API is ready")

		if err := caddy.ApplyConfig(); err != nil {
			log.Printf("Warning: Failed to apply initial Caddy configuration: %v", err)
			return
		}
		log.Println("Caddy configuration applied successfully on startup")
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api.RegisterRoutes(r)

	r.Run(":8080")
}

