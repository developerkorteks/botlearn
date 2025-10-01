// Package database - XRay Converter models
package database

import (
	"time"
)

// XRayConverter menyimpan konfigurasi converter command
type XRayConverter struct {
	ID              int       `json:"id" db:"id"`
	CommandName     string    `json:"command_name" db:"command_name"`         // "convertbizz"
	DisplayName     string    `json:"display_name" db:"display_name"`         // "XL-Line-WC"
	
	// Basic settings
	BugHost         string    `json:"bug_host" db:"bug_host"`                 // "ava.game.naver.com"
	ModifyType      string    `json:"modify_type" db:"modify_type"`           // "wildcard", "sni", "ws", "grpc", "custom"
	
	// Advanced flexible settings (empty = use original/auto)
	ServerTemplate  string    `json:"server_template" db:"server_template"`   // "{bug_host}" or "{bug_ip}" or custom
	HostTemplate    string    `json:"host_template" db:"host_template"`       // "{bug_host}.{original_server}" or custom
	SNITemplate     string    `json:"sni_template" db:"sni_template"`         // "{bug_host}.{original_server}" or custom
	
	// Network specific settings
	PathTemplate    string    `json:"path_template" db:"path_template"`       // "/rsv", "/vmess"
	GrpcServiceName string    `json:"grpc_service_name" db:"grpc_service_name"` // for grpc
	PortOverride    *int      `json:"port_override" db:"port_override"`       // optional port override
	
	// Status and tracking
	IsActive        bool      `json:"is_active" db:"is_active"`               // status aktif/tidak
	UsageCount      int       `json:"usage_count" db:"usage_count"`           // jumlah penggunaan
	CreatedBy       string    `json:"created_by" db:"created_by"`             // admin yang membuat
	CreatedAt       time.Time `json:"created_at" db:"created_at"`             // waktu dibuat
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`             // waktu diupdate
}

// XRayConversionLog menyimpan log penggunaan converter
type XRayConversionLog struct {
	ID               int       `json:"id" db:"id"`
	ConverterName    string    `json:"converter_name" db:"converter_name"`       // nama converter yang digunakan
	UserJID          string    `json:"user_jid" db:"user_jid"`                   // JID user
	GroupJID         string    `json:"group_jid" db:"group_jid"`                 // JID grup
	OriginalProtocol string    `json:"original_protocol" db:"original_protocol"` // vmess/vless/trojan/ss
	OriginalNetwork  string    `json:"original_network" db:"original_network"`   // ws/grpc/tcp/h2/httpupgrade
	OriginalServer   string    `json:"original_server" db:"original_server"`     // server original
	ModifiedServer   string    `json:"modified_server" db:"modified_server"`     // server setelah modifikasi
	Success          bool      `json:"success" db:"success"`                     // status berhasil/gagal
	ErrorMessage     *string   `json:"error_message" db:"error_message"`         // pesan error jika gagal
	UsedAt           time.Time `json:"used_at" db:"used_at"`                     // waktu penggunaan
}

// DetectedXRayConfig struktur untuk hasil detection XRay link
type DetectedXRayConfig struct {
	Protocol      string            `json:"protocol"`       // vmess/vless/trojan/ss
	Server        string            `json:"server"`         // server address
	Port          int               `json:"port"`           // port number
	UUID          string            `json:"uuid"`           // user id / password
	Network       string            `json:"network"`        // ws/grpc/tcp/h2/httpupgrade/kcp
	TLS           bool              `json:"tls"`            // tls enabled
	SNI           string            `json:"sni"`            // server name indication
	Host          string            `json:"host"`           // host header
	Path          string            `json:"path"`           // path for ws/httpupgrade/h2
	ServiceName   string            `json:"service_name"`   // grpc service name
	HeaderType    string            `json:"header_type"`    // kcp header type
	AlterID       int               `json:"alter_id"`       // vmess alter id
	Cipher        string            `json:"cipher"`         // encryption method
	Remarks       string            `json:"remarks"`        // connection name/remarks
	RawConfig     map[string]interface{} `json:"raw_config"` // original config for reconstruction
}

// ModifiedXRayConfig struktur untuk hasil modifikasi
type ModifiedXRayConfig struct {
	DetectedConfig *DetectedXRayConfig `json:"detected_config"`
	ModifiedLink   string              `json:"modified_link"`
	YAMLConfig     string              `json:"yaml_config"`
	ModifyType     string              `json:"modify_type"`
	BugHost        string              `json:"bug_host"`
	ModifiedServer string              `json:"modified_server"`
	ModifiedHost   string              `json:"modified_host"`
	ModifiedSNI    string              `json:"modified_sni"`
}