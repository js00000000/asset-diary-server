package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
)

type GeminiChatService struct {
	client *genai.Client
}

type ChatMessage struct {
	Role  string `json:"role"` // "user" or "model"
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const InvalidSymbolError = "invalid symbol"
const model = "gemini-3-flash"

func NewGeminiChatService() *GeminiChatService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		panic("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create genai client: %v", err))
	}

	return &GeminiChatService{
		client: client,
	}
}

func (s *GeminiChatService) GenerateContent(message string) (string, error) {
	if s.client == nil {
		panic("gemini client is not properly initialized")
	}

	ctx := context.Background()
	resp, err := s.client.Models.GenerateContent(ctx, model, genai.Text(message), &genai.GenerateContentConfig{
		Tools: []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		},
	})
	if err != nil {
		panic(err)
	}

	if resp == nil || len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		panic("no valid response from the model")
	}

	log.Println(resp.Text())
	return resp.Text(), nil
}

func (s *GeminiChatService) GenerateContentWithJSON(message string) (string, error) {
	resp, err := s.GenerateContent(message)
	if err != nil {
		return "", err
	}

	text := cleanMarkdownJSON(resp)
	var errResponse ErrorResponse
	if err := json.Unmarshal([]byte(text), &errResponse); err != nil {
		return "", err
	}
	if errResponse.Error != "" {
		return "", errors.New(InvalidSymbolError)
	}

	return text, nil
}

func cleanMarkdownJSON(input string) string {
	input = strings.TrimSpace(input)

	if strings.HasPrefix(input, "```json") {
		input = strings.TrimPrefix(input, "```json")
		input = strings.TrimSuffix(input, "```")
		return strings.TrimSpace(input)
	}

	if strings.HasPrefix(input, "```") {
		input = strings.TrimPrefix(input, "```")
		input = strings.TrimSuffix(input, "```")
		return strings.TrimSpace(input)
	}

	return input
}
