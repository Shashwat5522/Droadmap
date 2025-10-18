package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Document represents a PDF document stored in MongoDB
type Document struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantName    string             `bson:"tenant_name" json:"tenant_name"`
	FileName      string             `bson:"file_name" json:"file_name"`
	FileSize      int64              `bson:"file_size" json:"file_size"`
	StoragePath   string             `bson:"storage_path" json:"storage_path"`
	StorageURL    string             `bson:"storage_url" json:"storage_url"`
	ExtractedText string             `bson:"extracted_text" json:"extracted_text,omitempty"`
	Summary       string             `bson:"summary" json:"summary"`
	UploadedAt    time.Time          `bson:"uploaded_at" json:"uploaded_at"`
	IsDeleted     bool               `bson:"is_deleted" json:"is_deleted"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// UploadResponse represents the API response for upload
type UploadResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

