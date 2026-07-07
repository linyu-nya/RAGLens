package embedding

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

var ErrEmptyInput = errors.New("empty embedding input")

type Vector []float32

type Input struct {
	Texts []string
}

type Client interface {
	Embed(ctx context.Context, input Input) ([]Vector, error)
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

func (c *OpenAIClient) Embed(ctx context.Context, input Input) ([]Vector, error) {
	if len(input.Texts) == 0 {
		return nil, ErrEmptyInput
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(openAIEmbeddingRequest{
		Model: c.model,
		Input: input.Texts,
	}); err != nil {
		return nil, fmt.Errorf("encode embedding request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", &body)
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send embedding request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("embedding request failed: status %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	var decoded openAIEmbeddingResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	vectors := make([]Vector, len(decoded.Data))
	for _, item := range decoded.Data {
		if item.Index < 0 || item.Index >= len(decoded.Data) {
			return nil, fmt.Errorf("embedding response index out of range: %d", item.Index)
		}
		vectors[item.Index] = item.Embedding
	}

	return vectors, nil
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Index     int    `json:"index"`
		Embedding Vector `json:"embedding"`
	} `json:"data"`
}

