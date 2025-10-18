package services

import (
	"context"
	"fmt"

	"github.com/bacancy/droadmap/internal/models"
	"github.com/bacancy/droadmap/internal/repository"
	"github.com/jackc/pgx/v5"
)

// TenantService handles tenant management operations
type TenantService struct {
	postgresRepo *repository.PostgresRepository
	mongoRepo    *repository.MongoRepository
	mongoHost    string
	mongoPort    string
}

// NewTenantService creates a new tenant service
func NewTenantService(postgresRepo *repository.PostgresRepository, mongoRepo *repository.MongoRepository, mongoHost, mongoPort string) *TenantService {
	return &TenantService{
		postgresRepo: postgresRepo,
		mongoRepo:    mongoRepo,
		mongoHost:    mongoHost,
		mongoPort:    mongoPort,
	}
}

// GetOrCreateTenant gets an existing tenant or creates a new one with a dedicated database
func (s *TenantService) GetOrCreateTenant(ctx context.Context, tenantName string) (*models.Tenant, error) {
	// Step 1: Check if tenant exists in master database
	tenant, err := s.postgresRepo.GetTenantByName(ctx, tenantName)
	if err == nil {
		// Tenant exists
		return tenant, nil
	}

	// Check if error is "not found" or something else
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error checking tenant: %w", err)
	}

	// Step 2: Tenant doesn't exist - create new tenant database
	fmt.Printf("Creating new tenant database for: %s\n", tenantName)

	// Create MongoDB database for this tenant
	err = s.mongoRepo.CreateTenantDatabase(ctx, tenantName)
	if err != nil {
		return nil, fmt.Errorf("unable to create tenant database: %w", err)
	}

	// Step 3: Save tenant metadata to master database
	tenant = &models.Tenant{
		TenantName: tenantName,
		DBHost:     s.mongoHost,
		DBPort:     27017,
		DBName:     fmt.Sprintf("tenant_%s", tenantName),
		Status:     "active",
	}

	err = s.postgresRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("unable to save tenant metadata: %w", err)
	}

	fmt.Printf("✓ Tenant database created successfully for: %s\n", tenantName)
	return tenant, nil
}

// ValidateTenantName checks if tenant name is valid
func (s *TenantService) ValidateTenantName(tenantName string) error {
	if tenantName == "" {
		return fmt.Errorf("tenant name is required")
	}

	if len(tenantName) < 3 {
		return fmt.Errorf("tenant name must be at least 3 characters")
	}

	if len(tenantName) > 50 {
		return fmt.Errorf("tenant name must be less than 50 characters")
	}

	// Only allow alphanumeric and underscores
	for _, char := range tenantName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("tenant name can only contain letters, numbers, and underscores")
		}
	}

	return nil
}

// DeleteTenant performs soft delete on a tenant (marks as deleted in both PostgreSQL and MongoDB)
func (s *TenantService) DeleteTenant(ctx context.Context, tenantName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	stats["soft_deleted"] = false
	stats["documents_marked_deleted"] = int64(0)

	// Step 1: Check if tenant exists
	tenant, err := s.postgresRepo.GetTenantByName(ctx, tenantName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return stats, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return stats, fmt.Errorf("error checking tenant: %w", err)
	}

	// Step 2: Soft delete all documents in MongoDB
	modifiedCount, err := s.mongoRepo.SoftDeleteAllDocuments(ctx, tenantName)
	if err != nil {
		return stats, fmt.Errorf("failed to soft delete documents: %w", err)
	}
	stats["documents_marked_deleted"] = modifiedCount

	// Step 3: Soft delete tenant metadata in PostgreSQL
	err = s.postgresRepo.DeleteTenant(ctx, tenantName)
	if err != nil {
		return stats, fmt.Errorf("failed to soft delete tenant: %w", err)
	}
	stats["soft_deleted"] = true

	// Note: Files in MinIO are still preserved
	stats["note"] = "Tenant and all documents marked as deleted. Data can be restored."
	stats["storage_preserved"] = fmt.Sprintf("MinIO: %s/*", tenantName)

	fmt.Printf("Tenant '%s' soft deleted: DB=%s, Documents=%d marked as deleted\n", 
		tenantName, tenant.DBName, stats["documents_marked_deleted"])

	return stats, nil
}

// RestoreTenant restores a soft-deleted tenant and its documents
func (s *TenantService) RestoreTenant(ctx context.Context, tenantName string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Step 1: Restore tenant in PostgreSQL
	err := s.postgresRepo.RestoreTenant(ctx, tenantName)
	if err != nil {
		return stats, fmt.Errorf("failed to restore tenant: %w", err)
	}
	stats["tenant_restored"] = true

	// Step 2: Restore all documents in MongoDB
	modifiedCount, err := s.mongoRepo.RestoreAllDocuments(ctx, tenantName)
	if err != nil {
		return stats, fmt.Errorf("failed to restore documents: %w", err)
	}
	stats["documents_restored"] = modifiedCount

	fmt.Printf("Tenant '%s' restored: Documents=%d restored\n", tenantName, modifiedCount)
	return stats, nil
}

// ListDeletedTenants retrieves all soft-deleted tenants
func (s *TenantService) ListDeletedTenants(ctx context.Context) ([]models.Tenant, error) {
	return s.postgresRepo.ListDeletedTenants(ctx)
}

// ListTenants retrieves all tenants
func (s *TenantService) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	return s.postgresRepo.ListTenants(ctx)
}

