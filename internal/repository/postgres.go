package repository

import (
	"context"
	"fmt"

	"github.com/bacancy/droadmap/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository handles PostgreSQL operations for the master database
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(connString string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

// InitSchema creates the necessary tables if they don't exist
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS tenants (
		id SERIAL PRIMARY KEY,
		tenant_name VARCHAR(255) UNIQUE NOT NULL,
		db_host VARCHAR(500) NOT NULL,
		db_port INTEGER NOT NULL DEFAULT 27017,
		db_name VARCHAR(255) NOT NULL,
		status VARCHAR(50) DEFAULT 'active',
		is_deleted BOOLEAN DEFAULT FALSE,
		deleted_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tenant_name ON tenants(tenant_name);
	CREATE INDEX IF NOT EXISTS idx_is_deleted ON tenants(is_deleted);
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

// GetTenantByName retrieves a tenant by name (excluding soft-deleted)
func (r *PostgresRepository) GetTenantByName(ctx context.Context, tenantName string) (*models.Tenant, error) {
	query := `
		SELECT id, tenant_name, db_host, db_port, db_name, status, is_deleted, deleted_at, created_at, updated_at
		FROM tenants
		WHERE tenant_name = $1 AND is_deleted = FALSE
	`

	var tenant models.Tenant
	err := r.pool.QueryRow(ctx, query, tenantName).Scan(
		&tenant.ID,
		&tenant.TenantName,
		&tenant.DBHost,
		&tenant.DBPort,
		&tenant.DBName,
		&tenant.Status,
		&tenant.IsDeleted,
		&tenant.DeletedAt,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// CreateTenant creates a new tenant record
func (r *PostgresRepository) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (tenant_name, db_host, db_port, db_name, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		tenant.TenantName,
		tenant.DBHost,
		tenant.DBPort,
		tenant.DBName,
		tenant.Status,
	).Scan(&tenant.ID, &tenant.CreatedAt, &tenant.UpdatedAt)
}

// DeleteTenant performs soft delete on a tenant record
func (r *PostgresRepository) DeleteTenant(ctx context.Context, tenantName string) error {
	query := `
		UPDATE tenants 
		SET is_deleted = TRUE, deleted_at = NOW(), status = 'deleted'
		WHERE tenant_name = $1 AND is_deleted = FALSE
	`
	
	result, err := r.pool.Exec(ctx, query, tenantName)
	if err != nil {
		return fmt.Errorf("unable to delete tenant: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant '%s' not found or already deleted", tenantName)
	}
	
	return nil
}

// RestoreTenant restores a soft-deleted tenant
func (r *PostgresRepository) RestoreTenant(ctx context.Context, tenantName string) error {
	query := `
		UPDATE tenants 
		SET is_deleted = FALSE, deleted_at = NULL, status = 'active'
		WHERE tenant_name = $1 AND is_deleted = TRUE
	`
	
	result, err := r.pool.Exec(ctx, query, tenantName)
	if err != nil {
		return fmt.Errorf("unable to restore tenant: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant '%s' not found or not deleted", tenantName)
	}
	
	return nil
}

// ListTenants retrieves all active (non-deleted) tenants
func (r *PostgresRepository) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	query := `
		SELECT id, tenant_name, db_host, db_port, db_name, status, is_deleted, deleted_at, created_at, updated_at
		FROM tenants
		WHERE is_deleted = FALSE
		ORDER BY created_at DESC
	`
	
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to list tenants: %w", err)
	}
	defer rows.Close()
	
	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.TenantName,
			&tenant.DBHost,
			&tenant.DBPort,
			&tenant.DBName,
			&tenant.Status,
			&tenant.IsDeleted,
			&tenant.DeletedAt,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

// ListDeletedTenants retrieves all soft-deleted tenants
func (r *PostgresRepository) ListDeletedTenants(ctx context.Context) ([]models.Tenant, error) {
	query := `
		SELECT id, tenant_name, db_host, db_port, db_name, status, is_deleted, deleted_at, created_at, updated_at
		FROM tenants
		WHERE is_deleted = TRUE
		ORDER BY deleted_at DESC
	`
	
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("unable to list deleted tenants: %w", err)
	}
	defer rows.Close()
	
	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.TenantName,
			&tenant.DBHost,
			&tenant.DBPort,
			&tenant.DBName,
			&tenant.Status,
			&tenant.IsDeleted,
			&tenant.DeletedAt,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	
	return tenants, nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	r.pool.Close()
}

