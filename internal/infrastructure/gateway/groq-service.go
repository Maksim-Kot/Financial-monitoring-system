package gateway

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/gateway"
	"fms-project/internal/infrastructure/adapter"
	"fms-project/internal/infrastructure/logger"
)

const (
	chatCompletionsPath = "/openai/v1/chat/completions"
	textModel           = "llama-3.3-70b-versatile"
	visionModel         = "meta-llama/llama-4-scout-17b-16e-instruct"
)

//go:embed prompts/text-parser-system-prompt.txt
var promptTextParser string

//go:embed prompts/photo-parser-system-prompt.txt
var promptPhotoParser string

//go:embed prompts/category-classifier-system-prompt.txt
var promptCategoryClassifier string

type GroqServiceConfig struct {
	Logger  logger.Logger
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

type GroqService struct {
	logger     logger.Logger
	httpClient *adapter.HttpClient
}

func NewGroqService(cfg GroqServiceConfig) *GroqService {
	client := adapter.NewHttpClient().
		WithBaseURL(cfg.BaseURL).
		WithTimeout(cfg.Timeout).
		WithRetries(2).
		WithHeader("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))

	return &GroqService{
		logger:     cfg.Logger,
		httpClient: client,
	}
}

func (g *GroqService) ParseText(ctx context.Context, in gateway.TextParserGatewayIn) (gateway.TextParserGatewayOut, error) {
	prompt := strings.TrimSpace(in.Text)
	if prompt == "" {
		return gateway.TextParserGatewayOut{}, errors.New("text payload is empty")
	}

	req := chatCompletionRequest{
		Model: textModel,
		Messages: []chatRequestMessage{
			{
				Role:    "system",
				Content: strings.TrimSpace(promptTextParser),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	items, err := g.parseCompletion(ctx, req)
	if err != nil {
		return gateway.TextParserGatewayOut{}, err
	}

	return gateway.TextParserGatewayOut{Expenses: items}, nil
}

func (g *GroqService) ParsePhoto(ctx context.Context, in gateway.PhotoParserGatewayIn) (gateway.PhotoParserGatewayOut, error) {
	if len(in.Photo) == 0 {
		return gateway.PhotoParserGatewayOut{}, errors.New("photo payload is empty")
	}

	dataURL := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(in.Photo)
	req := chatCompletionRequest{
		Model: visionModel,
		Messages: []chatRequestMessage{
			{
				Role:    "system",
				Content: strings.TrimSpace(promptPhotoParser),
			},
			{
				Role: "user",
				Content: []map[string]any{
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": dataURL,
						},
					},
				},
			},
		},
	}

	items, err := g.parseCompletion(ctx, req)
	if err != nil {
		return gateway.PhotoParserGatewayOut{}, err
	}

	return gateway.PhotoParserGatewayOut{Expenses: items}, nil
}

func (g *GroqService) ClassifyCategories(ctx context.Context, in gateway.CategoryClassifierGatewayIn) (gateway.CategoryClassifierGatewayOut, error) {
	if len(in.Items) == 0 {
		return gateway.CategoryClassifierGatewayOut{Items: []entity.DraftItem{}}, nil
	}

	if len(in.Categories) == 0 {
		return gateway.CategoryClassifierGatewayOut{Items: in.Items}, nil
	}

	// build category map
	categories := make(map[string]entity.Category, len(in.Categories))
	for _, c := range in.Categories {
		categories[c.Name] = c
	}

	prompt := buildClassifyPrompt(in.Items, in.Categories)

	req := chatCompletionRequest{
		Model: textModel,
		Messages: []chatRequestMessage{
			{
				Role:    "system",
				Content: strings.TrimSpace(promptCategoryClassifier),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	classified, err := g.classifyCompletion(ctx, req)
	if err != nil {
		return gateway.CategoryClassifierGatewayOut{}, err
	}

	for i := range in.Items {
		item := &in.Items[i]

		for _, c := range classified {
			if strings.EqualFold(c.Name, item.Name) || c.ID == i+1 {
				if category, ok := categories[c.Category]; ok {
					item.Category = category
				}
				break
			}
		}
	}

	return gateway.CategoryClassifierGatewayOut{Items: in.Items}, nil
}

func (g *GroqService) parseCompletion(ctx context.Context, payload chatCompletionRequest) ([]entity.DraftItem, error) {
	var resp chatResponse
	if err := g.httpClient.PostJSON(ctx, chatCompletionsPath, payload, &resp); err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("groq response contains no choices")
	}

	g.logger.Debug("groq response", "response", resp.Choices[0].Message.Content)
	items, err := extractItemsFromContent(resp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func extractItemsFromContent(content string) ([]entity.DraftItem, error) {
	matches := arrayPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return nil, errors.New("groq content does not contain JSON array")
	}

	for _, rawArray := range matches {
		var parsed []extractedItem
		if err := json.Unmarshal([]byte(rawArray), &parsed); err != nil {
			continue
		}

		items := make([]entity.DraftItem, 0, len(parsed))
		for _, item := range parsed {
			name := strings.TrimSpace(item.Name)
			if name == "" || isZeroFloat64(item.Quantity) || isZeroFloat64(item.UnitPrice) {
				continue
			}

			items = append(items, entity.DraftItem{
				Name:      name,
				Quantity:  item.Quantity,
				UnitPrice: item.UnitPrice,
			})
		}
		return items, nil
	}

	return nil, errors.New("failed to decode item array from groq content")
}

func buildClassifyPrompt(items []entity.DraftItem, categories []entity.Category) string {
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = fmt.Sprintf("%d. %s", i+1, item.Name)
	}

	categoryNames := make([]string, len(categories))
	for i, c := range categories {
		categoryNames[i] = c.Name
	}

	return fmt.Sprintf("Products: %s\nCategories: %s", strings.Join(lines, "; "), strings.Join(categoryNames, ", "))

}

func (g *GroqService) classifyCompletion(ctx context.Context, payload chatCompletionRequest) ([]classifiedItem, error) {
	var resp chatResponse
	if err := g.httpClient.PostJSON(ctx, chatCompletionsPath, payload, &resp); err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("groq response contains no choices")
	}

	g.logger.Debug("groq classify response", "response", resp.Choices[0].Message.Content)
	items, err := extractClassificationFromContent(resp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func extractClassificationFromContent(content string) ([]classifiedItem, error) {
	matches := arrayPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return nil, errors.New("groq content does not contain JSON array")
	}

	for _, rawArray := range matches {
		var parsed []classifiedItem
		if err := json.Unmarshal([]byte(rawArray), &parsed); err != nil {
			continue
		}

		return parsed, nil
	}

	return nil, errors.New("failed to decode classification from groq content")
}

var arrayPattern = regexp.MustCompile(`(?s)\[[\s\S]*?\]`)

type chatCompletionRequest struct {
	Model    string               `json:"model"`
	Messages []chatRequestMessage `json:"messages"`
}

type chatRequestMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatMessage struct {
	Content string `json:"content"`
}

type extractedItem struct {
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type classifiedItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

const epsilon float64 = 1e-9

func isZeroFloat64(v float64) bool {
	return math.Abs(v) < epsilon
}
