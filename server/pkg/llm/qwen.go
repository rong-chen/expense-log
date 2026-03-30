package llm

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// qwenProvider 阿里千问（通义千问 VL）实现
// 基于 OpenAI 兼容模式，BaseURL 指向百炼 DashScope
type qwenProvider struct {
	client *openai.Client
	model  string
}

// NewQwenProvider 创建千问 Provider
func NewQwenProvider(cfg Config) Provider {
	config := openai.DefaultConfig(cfg.APIKey)
	config.BaseURL = cfg.BaseURL
	return &qwenProvider{
		client: openai.NewClientWithConfig(config),
		model:  cfg.Model,
	}
}

// AnalyzeImage 发送图片 URL + prompt，返回模型原始文本
func (q *qwenProvider) AnalyzeImage(ctx context.Context, imageURL, prompt string) (string, error) {
	resp, err := q.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: q.model,
		// 强制 JSON 输出，避免模型返回多余文字
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: prompt,
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("qwen vision api error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("qwen returned empty response")
	}

	return resp.Choices[0].Message.Content, nil
}
