package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

var ErrEmptyMessages = errors.New("empty chat messages")

type Message struct {
	Role    string
	Content string
}

type Input struct {
	Messages []Message
}

type Result struct {
	Content string
	Model   string
}

type Client interface {
	Complete(ctx context.Context, input Input) (Result, error)
}

type OpenAIOptions struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIClient(options OpenAIOptions) *OpenAIClient {
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &OpenAIClient{
		baseURL:    strings.TrimRight(options.BaseURL, "/"),
		apiKey:     options.APIKey,
		model:      options.Model,
		httpClient: httpClient,
	}
}

func (c *OpenAIClient) Complete(ctx context.Context, input Input) (Result, error) {
	if len(input.Messages) == 0 {
		return Result{}, ErrEmptyMessages
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(openAIChatRequest{
		Model:    c.model,
		Messages: input.Messages,
	}); err != nil {
		return Result{}, fmt.Errorf("encode chat request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", &body)
	if err != nil {
		return Result{}, fmt.Errorf("create chat request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("send chat request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return Result{}, fmt.Errorf("chat request failed: status %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	var decoded openAIChatResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return Result{}, fmt.Errorf("decode chat response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return Result{}, errors.New("chat response contains no choices")
	}

	model := decoded.Model
	if model == "" {
		model = c.model
	}

	return Result{
		Content: decoded.Choices[0].Message.Content,
		Model:   model,
	}, nil
}

type openAIChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openAIChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
