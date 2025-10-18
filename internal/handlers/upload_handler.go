package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bacancy/droadmap/internal/models"
	"github.com/bacancy/droadmap/internal/services"
	"github.com/bacancy/droadmap/internal/repository"
	"github.com/gin-gonic/gin"
)

// UploadHandler handles PDF upload requests
type UploadHandler struct {
	tenantService  *services.TenantService
	pdfService     *services.PDFService
	aiService      *services.AIService
	storageService *services.StorageService
	mongoRepo      *repository.MongoRepository
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(
	tenantService *services.TenantService,
	pdfService *services.PDFService,
	aiService *services.AIService,
	storageService *services.StorageService,
	mongoRepo *repository.MongoRepository,
) *UploadHandler {
	return &UploadHandler{
		tenantService:  tenantService,
		pdfService:     pdfService,
		aiService:      aiService,
		storageService: storageService,
		mongoRepo:      mongoRepo,
	}
}

// HandleUpload processes PDF upload requests
func (h *UploadHandler) HandleUpload(c *gin.Context) {
	startTime := time.Now()
	ctx := context.Background()

	// Step 1: Parse form data
	tenantName := c.PostForm("tenantName")
	file, err := c.FormFile("pdf")
	
	if err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   "PDF file is required",
		})
		return
	}

	// Step 2: Validate inputs
	if err := h.tenantService.ValidateTenantName(tenantName); err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid tenant name: %s", err.Error()),
		})
		return
	}

	if err := h.pdfService.ValidatePDF(file); err != nil {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid PDF file: %s", err.Error()),
		})
		return
	}

	fmt.Printf("\n📥 Processing upload for tenant: %s, file: %s\n", tenantName, file.Filename)

	// Step 3: Get or create tenant (creates MongoDB database if new)
	fmt.Println("→ Checking tenant database...")
	tenant, err := h.tenantService.GetOrCreateTenant(ctx, tenantName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get/create tenant: %s", err.Error()),
		})
		return
	}
	fmt.Printf("✓ Tenant database ready: %s\n", tenant.DBName)

	// Step 4: Extract text from PDF
	fmt.Println("→ Extracting text from PDF...")
	extractedText, err := h.pdfService.ExtractText(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to extract PDF content: %s", err.Error()),
		})
		return
	}
	fmt.Printf("✓ Extracted %d characters of text\n", len(extractedText))

	// Step 5: Upload file to storage
	fmt.Println("→ Uploading file to storage...")
	storagePath, storageURL, err := h.storageService.UploadFile(ctx, tenantName, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to store file: %s", err.Error()),
		})
		return
	}
	fmt.Printf("✓ File stored at: %s\n", storagePath)

	// Step 6: Generate AI summary
	fmt.Println("→ Generating AI summary...")
	summary, err := h.aiService.GenerateSummary(ctx, extractedText)
	if err != nil {
		fmt.Printf("⚠ AI summarization failed: %s\n", err.Error())
		summary = "Summary generation failed. Please check AI service configuration."
	} else {
		fmt.Printf("✓ Summary generated (%d characters)\n", len(summary))
	}

	// Step 7: Store document in tenant's MongoDB database
	fmt.Println("→ Storing document in database...")
	document := &models.Document{
		TenantName:    tenantName,
		FileName:      file.Filename,
		FileSize:      file.Size,
		StoragePath:   storagePath,
		StorageURL:    storageURL,
		ExtractedText: extractedText,
		Summary:       summary,
		UploadedAt:    time.Now(),
		IsDeleted:     false,
		DeletedAt:     nil,
	}

	err = h.mongoRepo.InsertDocument(ctx, tenantName, document)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to store document: %s", err.Error()),
		})
		return
	}

	processingTime := time.Since(startTime).Milliseconds()
	fmt.Printf("✓ Document stored successfully (ID: %s)\n", document.ID.Hex())
	fmt.Printf("✅ Total processing time: %dms\n\n", processingTime)

	// Step 8: Return success response
	c.JSON(http.StatusOK, models.UploadResponse{
		Success: true,
		Data: map[string]interface{}{
			"document_id":       document.ID.Hex(),
			"tenant_name":       tenantName,
			"file_name":         file.Filename,
			"file_size":         file.Size,
			"summary":           summary,
			"storage_url":       storageURL,
			"uploaded_at":       document.UploadedAt,
			"processing_time_ms": processingTime,
		},
	})
}

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HandleHealth returns service health status
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "pdf-ingestion-service",
	})
}

