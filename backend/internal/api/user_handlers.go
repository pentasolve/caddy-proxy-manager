package api

import (
	"caddy-proxy-manager/internal/auth"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"caddy-proxy-manager/internal/websocket"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)
func CheckSetup(c *gin.Context) {
	var count int64
	db.DB.Model(&models.User{}).Count(&count)
	c.JSON(http.StatusOK, gin.H{"setup_required": count == 0})
}
func SetupFirstUser(c *gin.Context) {
	var count int64
	db.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Setup already completed"})
		return
	}

	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	var adminRole models.Role
	db.DB.Where("name = ?", "Administrator").First(&adminRole)

	user := models.User{
		Username: input.Username,
		Password: hashedPassword,
		RoleID:   &adminRole.ID,
	}
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	logAudit(c, "setup", "Initial setup completed")
	c.JSON(http.StatusOK, gin.H{"message": "Setup completed successfully"})
}
func GetUsers(c *gin.Context) {
	var users []models.User
	db.DB.Preload("Role").Find(&users)
	c.JSON(http.StatusOK, users)
}
func CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		RoleID   uint   `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: input.Username,
		Password: hashedPassword,
	}
	if input.RoleID != 0 {
		user.RoleID = &input.RoleID
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	logAudit(c, "create_user", "Created user: "+user.Username)
	db.DB.Preload("Role").First(&user, user.ID)
	websocket.BroadcastMessage("user_created", user)
	c.JSON(http.StatusOK, user)
}
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		RoleID   uint   `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Password != "" {
		if len(input.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
			return
		}
		hashedPassword, err := auth.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}
	if input.RoleID != 0 {
		user.RoleID = &input.RoleID
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	logAudit(c, "update_user", "Updated user: "+user.Username)
	db.DB.Preload("Role").First(&user, user.ID)
	websocket.BroadcastMessage("user_updated", user)
	c.JSON(http.StatusOK, user)
}
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	if fmt.Sprintf("%v", userID) == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete yourself"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := db.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	logAudit(c, "delete_user", "Deleted user: "+user.Username)
	websocket.BroadcastMessage("user_deleted", gin.H{"id": id})
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user models.User
	if err := db.DB.Preload("Role").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		CurrentPassword string `json:"current_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Username != "" {
		user.Username = input.Username
	}
	if input.Password != "" {
		if input.CurrentPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required to change password"})
			return
		}
		if user.Password != input.CurrentPassword {
			c.JSON(http.StatusForbidden, gin.H{"error": "Incorrect current password"})
			return
		}

		user.Password = input.Password
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	logAudit(c, "update_profile", "Updated profile")
	c.JSON(http.StatusOK, user)
}
func GetRoles(c *gin.Context) {
	var roles []models.Role
	db.DB.Preload("Permissions").Find(&roles)
	c.JSON(http.StatusOK, roles)
}
func CreateRole(c *gin.Context) {
	var input struct {
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		Name:        input.Name,
		Description: input.Description,
	}

	if len(input.PermissionIDs) > 0 {
		var permissions []models.Permission
		db.DB.Find(&permissions, input.PermissionIDs)
		role.Permissions = permissions
	}

	if err := db.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	logAudit(c, "create_role", "Created role: "+role.Name)
	db.DB.Preload("Permissions").First(&role, role.ID)
	websocket.BroadcastMessage("role_created", role)
	c.JSON(http.StatusOK, role)
}
func UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := db.DB.Preload("Permissions").First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	var input struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		role.Name = input.Name
	}
	role.Description = input.Description

	if input.PermissionIDs != nil {
		var permissions []models.Permission
		db.DB.Find(&permissions, input.PermissionIDs)
		if err := db.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permissions"})
			return
		}
	}

	if err := db.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	logAudit(c, "update_role", "Updated role: "+role.Name)
	db.DB.Preload("Permissions").First(&role, role.ID)
	websocket.BroadcastMessage("role_updated", role)
	c.JSON(http.StatusOK, role)
}
func DeleteRole(c *gin.Context) {
	id := c.Param("id")
	var role models.Role
	if err := db.DB.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}
	var count int64
	db.DB.Model(&models.User{}).Where("role_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Cannot delete role: in use by %d users", count)})
		return
	}

	if err := db.DB.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	logAudit(c, "delete_role", "Deleted role: "+role.Name)
	websocket.BroadcastMessage("role_deleted", gin.H{"id": id})
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
}
func GetPermissions(c *gin.Context) {
	var permissions []models.Permission
	db.DB.Find(&permissions)
	c.JSON(http.StatusOK, permissions)
}
