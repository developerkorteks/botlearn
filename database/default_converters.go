// Package database - Default XRay converters untuk testing
package database

import (
	"time"
)

// DefaultXRayConverters berisi converter default untuk testing
var DefaultXRayConverters = []XRayConverter{
	{
		CommandName:     "convertbizz",
		DisplayName:     "XL-Line-WC",
		BugHost:         "ava.game.naver.com",
		ModifyType:      "wildcard",
		ServerTemplate:  "", // Empty = use legacy wildcard logic
		HostTemplate:    "", // Empty = use legacy wildcard logic
		SNITemplate:     "", // Empty = use legacy wildcard logic
		PathTemplate:    "/rsv",
		GrpcServiceName: "",
		PortOverride:    nil,
		IsActive:        true,
		UsageCount:      0,
		CreatedBy:       "system",
	},
	{
		CommandName:     "convertinsta",
		DisplayName:     "XL-Instagram-SNI",
		BugHost:         "chat.instagram.com",
		ModifyType:      "sni",
		PathTemplate:    "",
		GrpcServiceName: "",
		PortOverride:    nil,
		IsActive:        true,
		UsageCount:      0,
		CreatedBy:       "system",
	},
	{
		CommandName:     "convertnetflix",
		DisplayName:     "XL-Netflix-WS",
		BugHost:         "cache.netflix.com",
		ModifyType:      "ws",
		PathTemplate:    "/upvmess",
		GrpcServiceName: "",
		PortOverride:    nil,
		IsActive:        true,
		UsageCount:      0,
		CreatedBy:       "system",
	},
	{
		CommandName:     "convertgopay",
		DisplayName:     "XL-Gopay-Midtrans-WC",
		BugHost:         "api.midtrans.com",
		ModifyType:      "wildcard",
		PathTemplate:    "",
		GrpcServiceName: "vmess-grpc",
		PortOverride:    nil,
		IsActive:        true,
		UsageCount:      0,
		CreatedBy:       "system",
	},
	{
		CommandName:     "convertgrpc",
		DisplayName:     "Generic-gRPC",
		BugHost:         "cloudflare.com",
		ModifyType:      "grpc",
		PathTemplate:    "",
		GrpcServiceName: "grpc-service",
		PortOverride:    nil,
		IsActive:        true,
		UsageCount:      0,
		CreatedBy:       "system",
	},
	{
		CommandName:     "convertcustom",
		DisplayName:     "Custom-Template-Demo",
		BugHost:         "cloudflare.com",
		ModifyType:      "custom",
		ServerTemplate:  "{bug_host}",
		HostTemplate:    "{bug_host}.{original_server}",
		SNITemplate:     "{bug_host}.{original_server}",
		PathTemplate:    "",
		GrpcServiceName: "",
		PortOverride:    nil,
		IsActive:        true,
		UsageCount:      0,
		CreatedBy:       "system",
	},
}

// InsertDefaultConverters menambahkan default converters ke database
func InsertDefaultConverters(repo Repository) error {
	for _, converter := range DefaultXRayConverters {
		// Cek apakah converter sudah ada
		existing, err := repo.GetXRayConverter(converter.CommandName)
		if err != nil {
			return err
		}
		
		// Jika belum ada, tambahkan
		if existing == nil {
			converter.CreatedAt = time.Now()
			converter.UpdatedAt = time.Now()
			
			err = repo.CreateXRayConverter(&converter)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}