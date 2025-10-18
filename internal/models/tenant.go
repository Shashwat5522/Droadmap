package models

import "time"

// Tenant represents a tenant in the master database
type Tenant struct {
	ID           int        `json:"id"`
	TenantName   string     `json:"tenant_name"`
	DBHost       string     `json:"db_host"`
	DBPort       int        `json:"db_port"`
	DBName       string     `json:"db_name"`
	Status       string     `json:"status"` // active, provisioning, failed
	IsDeleted    bool       `json:"is_deleted"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

