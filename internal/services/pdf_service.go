package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFService handles PDF processing operations
type PDFService struct{}

// NewPDFService creates a new PDF service
func NewPDFService() *PDFService {
	return &PDFService{}
}

// ExtractText extracts text content from a PDF file
func (s *PDFService) ExtractText(file *multipart.FileHeader) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}
	defer src.Close()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		return "", fmt.Errorf("unable to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy uploaded file to temp file
	if _, err := io.Copy(tmpFile, src); err != nil {
		return "", fmt.Errorf("unable to copy file: %w", err)
	}

	// Close temp file before reading (required for pdf.Open)
	tmpFile.Close()

	// Read PDF content
	f, reader, err := pdf.Open(tmpFile.Name())
	if err != nil {
		// If PDF cannot be opened, return a placeholder text
		// This handles encrypted, corrupted, or unsupported PDFs
		return fmt.Sprintf("PDF file: %s (Text extraction not available - PDF may be scanned, encrypted, or in unsupported format)", file.Filename), nil
	}
	defer f.Close()

	// Extract text from all pages
	var textBuilder strings.Builder
	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Skip pages that fail to extract
			continue
		}

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n")
	}

	extractedText := strings.TrimSpace(textBuilder.String())
	
	// If no text could be extracted, return a placeholder
	if extractedText == "" {
		return fmt.Sprintf("PDF file: %s (No text content found - PDF may be image-based or scanned)", file.Filename), nil
	}

	return extractedText, nil
}

// ValidatePDF checks if the file is a valid PDF
func (s *PDFService) ValidatePDF(file *multipart.FileHeader) error {
	if file == nil {
		return fmt.Errorf("file is required")
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		return fmt.Errorf("file must be a PDF")
	}

	// Check file size (max 50MB)
	if file.Size > 50*1024*1024 {
		return fmt.Errorf("file size must be less than 50MB")
	}

	if file.Size == 0 {
		return fmt.Errorf("file is empty")
	}

	return nil
}

