package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// AIService handles AI summarization operations
type AIService struct {
	client *openai.Client
}

// NewAIService creates a new AI service
func NewAIService(apiKey string) *AIService {
	var client *openai.Client
	if apiKey != "" {
		client = openai.NewClient(apiKey)
	}
	return &AIService{client: client}
}

// GenerateSummary generates a summary of the given text using OpenAI
func (s *AIService) GenerateSummary(ctx context.Context, text string) (string, error) {
	// If no API key is configured, return a simple fallback summary
	if s.client == nil {
		return s.generateFallbackSummary(text), nil
	}

	// Truncate text if it's too long (OpenAI has token limits)
	maxChars := 12000 // Roughly 3000 tokens
	if len(text) > maxChars {
		text = text[:maxChars] + "..."
	}

	// Call OpenAI API
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant that summarizes documents concisely in 2-3 sentences.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Please provide a concise summary of the following document:\n\n%s", text),
			},
		},
		MaxTokens:   150,
		Temperature: 0.7,
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// generateFallbackSummary creates a simple summary when AI service is unavailable
func (s *AIService) generateFallbackSummary(text string) string {
	// Take first 500 characters as a simple summary
	maxLen := 500
	if len(text) < maxLen {
		maxLen = len(text)
	}

	summary := text[:maxLen]
	
	// Try to end at a sentence
	if lastPeriod := strings.LastIndex(summary, "."); lastPeriod > 100 {
		summary = summary[:lastPeriod+1]
	}

	return strings.TrimSpace(summary) + "..."
}

