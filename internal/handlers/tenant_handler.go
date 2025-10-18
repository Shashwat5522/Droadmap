package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bacancy/droadmap/internal/models"
	"github.com/bacancy/droadmap/internal/services"
	"github.com/gin-gonic/gin"
)

// TenantHandler handles tenant-related requests
type TenantHandler struct {
	tenantService *services.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// DeleteTenant handles tenant soft deletion requests
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	ctx := context.Background()
	tenantName := c.Param("name")

	fmt.Printf("\n🗑️  Soft delete request for tenant: %s\n", tenantName)

	// Validate tenant name
	if err := h.tenantService.ValidateTenantName(tenantName); err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid tenant name: %s", err.Error()),
		})
		return
	}

	// Soft delete the tenant
	fmt.Println("→ Marking tenant as deleted (data will be preserved)...")
	stats, err := h.tenantService.DeleteTenant(ctx, tenantName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete tenant: %s", err.Error()),
		})
		return
	}

	fmt.Printf("✓ Tenant '%s' soft deleted successfully\n", tenantName)
	fmt.Printf("  - Marked as deleted: %v\n", stats["soft_deleted"])
	fmt.Printf("  - Documents marked deleted: %v\n", stats["documents_marked_deleted"])
	fmt.Printf("  - Note: %v\n\n", stats["note"])

	// Return success response
	c.JSON(http.StatusOK, models.UploadResponse{
		Success: true,
		Data: map[string]interface{}{
			"tenant_name":              tenantName,
			"soft_deleted":             stats["soft_deleted"],
			"documents_marked_deleted": stats["documents_marked_deleted"],
			"can_restore":              true,
			"message":                  fmt.Sprintf("Tenant '%s' and %v documents marked as deleted. Can be restored.", tenantName, stats["documents_marked_deleted"]),
			"restore_command":          fmt.Sprintf("POST /api/v1/tenant/%s/restore", tenantName),
		},
	})
}

// RestoreTenant handles tenant restoration requests
func (h *TenantHandler) RestoreTenant(c *gin.Context) {
	ctx := context.Background()
	tenantName := c.Param("name")

	fmt.Printf("\n♻️  Restore request for tenant: %s\n", tenantName)

	// Validate tenant name
	if err := h.tenantService.ValidateTenantName(tenantName); err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid tenant name: %s", err.Error()),
		})
		return
	}

	// Restore the tenant
	fmt.Println("→ Restoring tenant and documents...")
	stats, err := h.tenantService.RestoreTenant(ctx, tenantName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to restore tenant: %s", err.Error()),
		})
		return
	}

	fmt.Printf("✓ Tenant '%s' restored successfully\n", tenantName)
	fmt.Printf("  - Tenant restored: %v\n", stats["tenant_restored"])
	fmt.Printf("  - Documents restored: %v\n\n", stats["documents_restored"])

	// Return success response
	c.JSON(http.StatusOK, models.UploadResponse{
		Success: true,
		Data: map[string]interface{}{
			"tenant_name":        tenantName,
			"restored":           true,
			"documents_restored": stats["documents_restored"],
			"status":             "active",
			"message":            fmt.Sprintf("Tenant '%s' and %v documents have been restored and are now active", tenantName, stats["documents_restored"]),
		},
	})
}

// ListTenants handles listing all active tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	ctx := context.Background()

	fmt.Println("\n📋 Listing all active tenants...")
	tenants, err := h.tenantService.ListTenants(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to list tenants: %s", err.Error()),
		})
		return
	}

	fmt.Printf("✓ Found %d active tenant(s)\n\n", len(tenants))

	c.JSON(http.StatusOK, models.UploadResponse{
		Success: true,
		Data: map[string]interface{}{
			"tenants": tenants,
			"count":   len(tenants),
			"status":  "active",
		},
	})
}

// ListDeletedTenants handles listing all soft-deleted tenants
func (h *TenantHandler) ListDeletedTenants(c *gin.Context) {
	ctx := context.Background()

	fmt.Println("\n📋 Listing all deleted tenants...")
	tenants, err := h.tenantService.ListDeletedTenants(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to list deleted tenants: %s", err.Error()),
		})
		return
	}

	fmt.Printf("✓ Found %d deleted tenant(s)\n\n", len(tenants))

	c.JSON(http.StatusOK, models.UploadResponse{
		Success: true,
		Data: map[string]interface{}{
			"tenants":   tenants,
			"count":     len(tenants),
			"status":    "deleted",
			"note":      "These tenants can be restored using POST /api/v1/tenant/:name/restore",
		},
	})
}

