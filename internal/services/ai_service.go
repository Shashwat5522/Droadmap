package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AIService handles AI summarization using Google Gemini API
type AIService struct {
	apiKey string
	client *http.Client
}

// NewAIService creates a new AI service with Google Gemini
func NewAIService(geminiAPIKey string) *AIService {
	service := &AIService{
		apiKey: geminiAPIKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if geminiAPIKey != "" {
		fmt.Println("✓ AI Provider: Google Gemini 2.5 Flash (Free Tier)")
	} else {
		fmt.Println("✓ AI Provider: Fallback (Text Extraction)")
	}

	return service
}

// GenerateSummary generates a summary of the given text using Google Gemini API
func (s *AIService) GenerateSummary(ctx context.Context, text string) (string, error) {
	if s.apiKey == "" {
		return s.generateFallbackSummary(text), nil
	}

	maxChars := 30000
	if len(text) > maxChars {
		text = text[:maxChars] + "..."
	}

	summary, err := s.callGeminiAPI(ctx, text)
	if err != nil {
		fmt.Printf("⚠ Gemini API error: %v, using fallback\n", err)
		return s.generateFallbackSummary(text), nil
	}

	return summary, nil
}

// callGeminiAPI makes the HTTP request to Google Gemini API
func (s *AIService) callGeminiAPI(ctx context.Context, text string) (string, error) {
	// Using gemini-2.5-flash which is available in free tier
	// This is the latest, fastest model available (v1 API is stable)
	endpoint := "https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent"
	url := fmt.Sprintf("%s?key=%s", endpoint, s.apiKey)

	// Build payload - NOTE: we don't use generationConfig as it can cause MAX_TOKENS issues
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": s.buildPrompt(text),
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	// Extract text from response
	if len(geminiResp.Candidates) > 0 {
		candidate := geminiResp.Candidates[0]
		if len(candidate.Content.Parts) > 0 {
			result := candidate.Content.Parts[0].Text
			return strings.TrimSpace(result), nil
		}
	}

	return "", fmt.Errorf("no text in response")
}

// buildPrompt constructs the prompt for Gemini
func (s *AIService) buildPrompt(text string) string {
	sys := "You are a helpful assistant that summarizes documents in 2-3 sentences."
	usr := "Provide a concise summary of this document:"
	return sys + "\n\n" + usr + "\n\n" + text + "\n\nSummary:"
}

// GeminiResponse represents Gemini API response structure
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
		Index        int    `json:"index"`
	} `json:"candidates"`
}

// generateFallbackSummary creates basic summary from text
func (s *AIService) generateFallbackSummary(text string) string {
	maxLen := 500
	if len(text) < maxLen {
		maxLen = len(text)
	}

	summary := text[:maxLen]

	if lastPeriod := strings.LastIndex(summary, "."); lastPeriod > 100 {
		summary = summary[:lastPeriod+1]
	}

	return strings.TrimSpace(summary) + "..."
}