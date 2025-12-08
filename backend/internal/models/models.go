package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Code        string `gorm:"unique" json:"code"`
	Description string `json:"description"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Password string `json:"-"`
	RoleID   *uint  `json:"role_id"`
	Role     Role   `json:"role"`
}

type Host struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Domain              string         `gorm:"unique" json:"domain"`
	Target              string         `json:"target"`
	Type                string         `json:"type"`
	SSL                 bool           `json:"ssl"`
	SSLProvider         string         `json:"ssl_provider"`
	SSLActualProvider   string         `json:"ssl_actual_provider"`
	HSTSEnabled         bool           `json:"hsts_enabled"`
	BlockExploits       bool           `json:"block_exploits"`
	CacheAssets         bool           `json:"cache_assets"`
	Websockets          bool           `json:"websockets"`
	ForwardingCode      int            `json:"forwarding_code"`
	UseDNSChallenge     bool           `json:"use_dns_challenge"`
	DNSProvider         string         `json:"dns_provider"`
	DNSToken            string         `json:"dns_token"`
	AccessListID        *uint          `json:"access_list_id"`
	CertificateID       *uint          `json:"certificate_id"`
	SSLStatus           string         `json:"ssl_status" gorm:"default:'ready'"`
	SSLError            string         `json:"ssl_error"`
	Locations           []Location     `json:"locations" gorm:"foreignKey:HostID;constraint:OnDelete:CASCADE"`
	Upstreams           []Upstream     `json:"upstreams" gorm:"foreignKey:HostID;constraint:OnDelete:CASCADE"`
	LoadBalancing       string         `json:"load_balancing" gorm:"default:''"`
	LBTryDuration       string         `json:"lb_try_duration"`
	LBTryInterval       string         `json:"lb_try_interval"`
	HealthCheck         bool           `json:"health_check" gorm:"default:false"`
	HealthCheckPath     string         `json:"health_check_path"`
	HealthCheckInterval string         `json:"health_check_interval"`
	IsActive            bool           `json:"is_active" gorm:"default:true"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

type Upstream struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	HostID      *uint  `json:"host_id"`
	LocationID  *uint  `json:"location_id"`
	Target      string `json:"target"`
	Weight      int    `json:"weight" gorm:"default:1"`
	MaxFails    int    `json:"max_fails" gorm:"default:0"`
	FailTimeout string `json:"fail_timeout"`
}

type Location struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	HostID              uint       `json:"host_id"`
	Path                string     `json:"path"`
	Target              string     `json:"target"`
	Upstreams           []Upstream `json:"upstreams" gorm:"foreignKey:LocationID;constraint:OnDelete:CASCADE"`
	LoadBalancing       string     `json:"load_balancing" gorm:"default:''"`
	LBTryDuration       string     `json:"lb_try_duration"`
	LBTryInterval       string     `json:"lb_try_interval"`
	HealthCheck         bool       `json:"health_check" gorm:"default:false"`
	HealthCheckPath     string     `json:"health_check_path"`
	HealthCheckInterval string     `json:"health_check_interval"`
}

type AccessList struct {
	ID        uint               `json:"id" gorm:"primaryKey"`
	Name      string             `json:"name"`
	Clients   []AccessListClient `json:"clients" gorm:"foreignKey:AccessListID;constraint:OnDelete:CASCADE"`
	Rules     []AccessListRule   `json:"rules" gorm:"foreignKey:AccessListID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time          `json:"created_at"`
}

type AccessListClient struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	AccessListID uint   `json:"access_list_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type AccessListRule struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	AccessListID uint   `json:"access_list_id"`
	IP           string `json:"ip"`
	Action       string `json:"action"`
}

type Certificate struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Domain    string         `json:"domain"`
	CertFile  string         `json:"cert_file"`
	KeyFile   string         `json:"key_file"`
	Provider  string         `json:"provider"`
	Status    string         `json:"status" gorm:"default:'ready'"`
	Error     string         `json:"error"`
	AutoRenew bool           `json:"auto_renew" gorm:"default:true"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

type Stream struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Name              string         `json:"name"`
	ListenPort        int            `json:"listen_port"`
	TCPEnabled        bool           `json:"tcp_enabled" gorm:"default:true"`
	UDPEnabled        bool           `json:"udp_enabled" gorm:"default:false"`
	Target            string         `json:"target"`
	SSL               bool           `json:"ssl" gorm:"default:false"`
	SSLProvider       string         `json:"ssl_provider"`
	SSLActualProvider string         `json:"ssl_actual_provider"`
	SSLStatus         string         `json:"ssl_status" gorm:"default:'ready'"`
	SSLError          string         `json:"ssl_error"`
	CertificateID     *uint          `json:"certificate_id"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type RevokedToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TokenID   string    `gorm:"uniqueIndex" json:"token_id"`
	UserID    uint      `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt time.Time `json:"revoked_at"`
}
