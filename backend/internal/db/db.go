package db

import (
	"caddy-proxy-manager/internal/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("/data/caddy-proxy-manager.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		DB, err = gorm.Open(sqlite.Open("caddy-proxy-manager.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			panic("failed to connect database")
		}
	}
	err = DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.Host{}, &models.Location{}, &models.Upstream{}, &models.Setting{}, &models.Certificate{}, &models.AccessList{}, &models.AccessListClient{}, &models.AccessListRule{}, &models.AuditLog{}, &models.Stream{}, &models.RevokedToken{})
	if err != nil {
		log.Fatal("failed to migrate database")
	}
	var count int64
	DB.Raw("SELECT COUNT(*) FROM pragma_table_info('hosts') WHERE name='ssl_actual_provider'").Scan(&count)
	if count == 0 {
		DB.Exec("ALTER TABLE hosts ADD COLUMN ssl_actual_provider TEXT DEFAULT ''")
	}
	DB.Raw("SELECT COUNT(*) FROM pragma_table_info('streams') WHERE name='ssl_actual_provider'").Scan(&count)
	if count == 0 {
		DB.Exec("ALTER TABLE streams ADD COLUMN ssl_actual_provider TEXT DEFAULT ''")
	}
	permissions := []models.Permission{
		{Name: "View Hosts", Code: "hosts:read", Description: "Can view proxy hosts"},
		{Name: "Create Hosts", Code: "hosts:create", Description: "Can create proxy hosts"},
		{Name: "Update Hosts", Code: "hosts:update", Description: "Can update proxy hosts"},
		{Name: "Delete Hosts", Code: "hosts:delete", Description: "Can delete proxy hosts"},
		{Name: "View Streams", Code: "streams:read", Description: "Can view streams"},
		{Name: "Create Streams", Code: "streams:create", Description: "Can create streams"},
		{Name: "Update Streams", Code: "streams:update", Description: "Can update streams"},
		{Name: "Delete Streams", Code: "streams:delete", Description: "Can delete streams"},
		{Name: "View Access Lists", Code: "access_lists:read", Description: "Can view access lists"},
		{Name: "Create Access Lists", Code: "access_lists:create", Description: "Can create access lists"},
		{Name: "Update Access Lists", Code: "access_lists:update", Description: "Can update access lists"},
		{Name: "Delete Access Lists", Code: "access_lists:delete", Description: "Can delete access lists"},
		{Name: "View Certificates", Code: "certificates:read", Description: "Can view certificates"},
		{Name: "Create Certificates", Code: "certificates:create", Description: "Can create certificates"},
		{Name: "Delete Certificates", Code: "certificates:delete", Description: "Can delete certificates"},
		{Name: "View Users", Code: "users:read", Description: "Can view users"},
		{Name: "Create Users", Code: "users:create", Description: "Can create users"},
		{Name: "Update Users", Code: "users:update", Description: "Can update users"},
		{Name: "Delete Users", Code: "users:delete", Description: "Can delete users"},
		{Name: "View Roles", Code: "roles:read", Description: "Can view roles"},
		{Name: "Create Roles", Code: "roles:create", Description: "Can create roles"},
		{Name: "Update Roles", Code: "roles:update", Description: "Can update roles"},
		{Name: "Delete Roles", Code: "roles:delete", Description: "Can delete roles"},
		{Name: "View Settings", Code: "settings:read", Description: "Can view settings"},
		{Name: "Update Settings", Code: "settings:update", Description: "Can update settings"},
		{Name: "View Logs", Code: "logs:read", Description: "Can view audit logs"},
	}

	for _, p := range permissions {
		DB.FirstOrCreate(&p, models.Permission{Code: p.Code})
	}
	var roleCount int64
	DB.Model(&models.Role{}).Count(&roleCount)
	if roleCount == 0 {
		var allPermissions []models.Permission
		DB.Find(&allPermissions)
		DB.Create(&models.Role{Name: "Administrator", Description: "Full access to all features", Permissions: allPermissions})
		var readPermissions []models.Permission
		DB.Where("code LIKE ?", "%:read").Find(&readPermissions)
		DB.Create(&models.Role{Name: "User", Description: "Limited access", Permissions: readPermissions})
	} else {
		var adminRole models.Role
		if err := DB.Where("name = ?", "Administrator").First(&adminRole).Error; err == nil {
			var allPermissions []models.Permission
			DB.Find(&allPermissions)
			DB.Model(&adminRole).Association("Permissions").Replace(allPermissions)
		}
	}
	DB.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Host{})
}
