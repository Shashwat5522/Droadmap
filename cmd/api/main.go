package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/bacancy/droadmap/internal/config"
	"github.com/bacancy/droadmap/internal/handlers"
	"github.com/bacancy/droadmap/internal/repository"
	"github.com/bacancy/droadmap/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🚀 Starting Multi-Tenant PDF Ingestion Service...")

	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No .env file found, using environment variables or defaults")
	}

	// Load configuration
	cfg := config.Load()
	fmt.Printf("✓ Configuration loaded\n")

	// Initialize PostgreSQL (Master Database)
	fmt.Printf("→ Connecting to PostgreSQL (Master DB)...\n")
	postgresRepo, err := repository.NewPostgresRepository(cfg.PostgresConnString())
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresRepo.Close()
	fmt.Printf("✓ Connected to PostgreSQL\n")

	// Initialize schema
	if err := postgresRepo.InitSchema(context.Background()); err != nil {
		log.Fatalf("❌ Failed to initialize schema: %v", err)
	}
	fmt.Printf("✓ Database schema initialized\n")

	// Initialize MongoDB (Tenant Databases)
	fmt.Printf("→ Connecting to MongoDB...\n")
	mongoRepo, err := repository.NewMongoRepository(cfg.MongoConnString())
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer mongoRepo.Close(context.Background())
	fmt.Printf("✓ Connected to MongoDB\n")

	// Initialize MinIO (Storage)
	fmt.Printf("→ Connecting to MinIO...\n")
	storageService, err := services.NewStorageService(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}

	// Ensure bucket exists
	if err := storageService.EnsureBucketExists(context.Background()); err != nil {
		log.Fatalf("❌ Failed to ensure bucket exists: %v", err)
	}
	fmt.Printf("✓ Storage initialized (bucket: %s)\n", cfg.MinIOBucket)

	// Initialize services
	tenantService := services.NewTenantService(postgresRepo, mongoRepo, cfg.MongoHost, cfg.MongoPort)
	pdfService := services.NewPDFService()
	aiService := services.NewAIService(cfg.GeminiAPIKey)


	// Initialize handlers
	uploadHandler := handlers.NewUploadHandler(tenantService, pdfService, aiService, storageService, mongoRepo)
	tenantHandler := handlers.NewTenantHandler(tenantService)
	healthHandler := handlers.NewHealthHandler()

	// Setup Gin router
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Routes
	router.GET("/health", healthHandler.HandleHealth)
	
	v1 := router.Group("/api/v1")
	{
		// Upload endpoint
		v1.POST("/upload", uploadHandler.HandleUpload)
		
		// Tenant management endpoints
		v1.GET("/tenants", tenantHandler.ListTenants)
		v1.GET("/tenants/deleted", tenantHandler.ListDeletedTenants)
		v1.DELETE("/tenant/:name", tenantHandler.DeleteTenant)
		v1.POST("/tenant/:name/restore", tenantHandler.RestoreTenant)
	}

	// Start server
	fmt.Printf("\n✅ Server ready!\n")
	fmt.Printf("📡 Listening on port %s\n", cfg.Port)
	fmt.Printf("\n📚 API Endpoints:\n")
	fmt.Printf("  POST   http://localhost:%s/api/v1/upload\n", cfg.Port)
	fmt.Printf("  GET    http://localhost:%s/api/v1/tenants\n", cfg.Port)
	fmt.Printf("  GET    http://localhost:%s/api/v1/tenants/deleted\n", cfg.Port)
	fmt.Printf("  DELETE http://localhost:%s/api/v1/tenant/:name (soft delete)\n", cfg.Port)
	fmt.Printf("  POST   http://localhost:%s/api/v1/tenant/:name/restore\n", cfg.Port)
	fmt.Printf("  GET    http://localhost:%s/health\n\n", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

