package api

import (
	"caddy-proxy-manager/internal/caddy"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"caddy-proxy-manager/internal/websocket"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func RegisterAccessListRoutes(r *gin.RouterGroup) {
	r.GET("/access-lists", getAccessLists)
	r.POST("/access-lists", createAccessList)
	r.PUT("/access-lists/:id", updateAccessList)
	r.DELETE("/access-lists/:id", deleteAccessList)
}

func getAccessLists(c *gin.Context) {
	var lists []models.AccessList
	if err := db.DB.Preload("Clients").Preload("Rules").Find(&lists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lists)
}

func createAccessList(c *gin.Context) {
	var input struct {
		Name    string `json:"name"`
		Clients []struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"clients"`
		Rules []struct {
			IP     string `json:"ip"`
			Action string `json:"action"`
		} `json:"rules"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	usernameSet := make(map[string]bool)
	for _, client := range input.Clients {
		username := strings.TrimSpace(strings.ToLower(client.Username))
		if username == "" && client.Password != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required for all users with password"})
			return
		}
		if username != "" && client.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required for all users"})
			return
		}
		if username != "" {
			if usernameSet[username] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate usernames are not allowed"})
				return
			}
			usernameSet[username] = true
		}
	}
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(\/\d{1,2})?$`)
	for _, rule := range input.Rules {
		ip := strings.TrimSpace(rule.IP)
		if ip != "" && !ipRegex.MatchString(ip) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid IP format: %s. Use format: 192.168.1.1 or 192.168.1.0/24", rule.IP)})
			return
		}
	}

	list := models.AccessList{
		Name: strings.TrimSpace(input.Name),
	}

	for _, client := range input.Clients {
		if strings.TrimSpace(client.Username) == "" {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(client.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		list.Clients = append(list.Clients, models.AccessListClient{
			Username: strings.TrimSpace(client.Username),
			Password: string(hash),
		})
	}

	for _, rule := range input.Rules {
		if strings.TrimSpace(rule.IP) == "" {
			continue
		}
		list.Rules = append(list.Rules, models.AccessListRule{
			IP:     strings.TrimSpace(rule.IP),
			Action: rule.Action,
		})
	}

	if err := db.DB.Create(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logAudit(c, "create_access_list", fmt.Sprintf("Created access list: %s (%d users, %d rules)", list.Name, len(list.Clients), len(list.Rules)))

	websocket.BroadcastMessage("access_list_created", list)
	c.JSON(http.StatusOK, list)
}

func updateAccessList(c *gin.Context) {
	id := c.Param("id")
	var list models.AccessList
	if err := db.DB.Preload("Clients").Preload("Rules").First(&list, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "access list not found"})
		return
	}

	var input struct {
		Name    string `json:"name"`
		Clients []struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"clients"`
		Rules []struct {
			IP     string `json:"ip"`
			Action string `json:"action"`
		} `json:"rules"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(input.Name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	usernameSet := make(map[string]bool)
	for _, client := range input.Clients {
		username := strings.TrimSpace(strings.ToLower(client.Username))
		if username == "" && client.Password != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required for all users with password"})
			return
		}
		if username != "" {
			if usernameSet[username] {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate usernames are not allowed"})
				return
			}
			usernameSet[username] = true
		}
	}
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(\/\d{1,2})?$`)
	for _, rule := range input.Rules {
		ip := strings.TrimSpace(rule.IP)
		if ip != "" && !ipRegex.MatchString(ip) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid IP format: %s. Use format: 192.168.1.1 or 192.168.1.0/24", rule.IP)})
			return
		}
	}
	oldUsers := make(map[string]bool)
	for _, cl := range list.Clients {
		oldUsers[cl.Username] = true
	}
	oldRules := make(map[string]string)
	for _, r := range list.Rules {
		oldRules[r.IP] = r.Action
	}
	list.Name = strings.TrimSpace(input.Name)
	db.DB.Delete(&models.AccessListClient{}, "access_list_id = ?", list.ID)
	list.Clients = []models.AccessListClient{}

	for _, client := range input.Clients {
		if strings.TrimSpace(client.Username) == "" {
			continue
		}
		if client.Password == "" {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(client.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		list.Clients = append(list.Clients, models.AccessListClient{
			AccessListID: list.ID,
			Username:     strings.TrimSpace(client.Username),
			Password:     string(hash),
		})
	}
	db.DB.Delete(&models.AccessListRule{}, "access_list_id = ?", list.ID)
	list.Rules = []models.AccessListRule{}

	for _, rule := range input.Rules {
		if strings.TrimSpace(rule.IP) == "" {
			continue
		}
		list.Rules = append(list.Rules, models.AccessListRule{
			AccessListID: list.ID,
			IP:           strings.TrimSpace(rule.IP),
			Action:       rule.Action,
		})
	}

	if err := db.DB.Save(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go caddy.ApplyConfig()
	var changes []string
	newUsers := make(map[string]bool)
	for _, c := range input.Clients {
		newUsers[c.Username] = true
	}

	var addedUsers, removedUsers []string
	for u := range newUsers {
		if !oldUsers[u] {
			addedUsers = append(addedUsers, u)
		}
	}
	for u := range oldUsers {
		if !newUsers[u] {
			removedUsers = append(removedUsers, u)
		}
	}
	if len(addedUsers) > 0 {
		changes = append(changes, fmt.Sprintf("Added users: %s", strings.Join(addedUsers, ", ")))
	}
	if len(removedUsers) > 0 {
		changes = append(changes, fmt.Sprintf("Removed users: %s", strings.Join(removedUsers, ", ")))
	}
	newRules := make(map[string]string)
	for _, r := range input.Rules {
		newRules[r.IP] = r.Action
	}

	var addedRules, removedRules, modifiedRules []string
	for ip, action := range newRules {
		if oldAction, exists := oldRules[ip]; !exists {
			addedRules = append(addedRules, fmt.Sprintf("%s (%s)", ip, action))
		} else if oldAction != action {
			modifiedRules = append(modifiedRules, fmt.Sprintf("%s (%s -> %s)", ip, oldAction, action))
		}
	}
	for ip, action := range oldRules {
		if _, exists := newRules[ip]; !exists {
			removedRules = append(removedRules, fmt.Sprintf("%s (%s)", ip, action))
		}
	}

	if len(addedRules) > 0 {
		changes = append(changes, fmt.Sprintf("Added rules: %s", strings.Join(addedRules, ", ")))
	}
	if len(removedRules) > 0 {
		changes = append(changes, fmt.Sprintf("Removed rules: %s", strings.Join(removedRules, ", ")))
	}
	if len(modifiedRules) > 0 {
		changes = append(changes, fmt.Sprintf("Modified rules: %s", strings.Join(modifiedRules, ", ")))
	}

	details := fmt.Sprintf("Updated access list: %s", list.Name)
	if len(changes) > 0 {
		details += ". " + strings.Join(changes, ". ")
	}

	logAudit(c, "update_access_list", details)

	websocket.BroadcastMessage("access_list_updated", list)
	c.JSON(http.StatusOK, list)
}

func deleteAccessList(c *gin.Context) {
	id := c.Param("id")
	var list models.AccessList
	db.DB.First(&list, id)

	if err := db.DB.Delete(&models.AccessList{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go caddy.ApplyConfig()

	logAudit(c, "delete_access_list", "Deleted access list: "+list.Name)

	websocket.BroadcastMessage("access_list_deleted", gin.H{"id": id})
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
