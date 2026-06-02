package llm

import (
	"context"
	"fmt"
)

// Provider 大模型统一调用接口
// 所有模型厂商（千问、豆包、智谱等）都实现此接口
type Provider interface {
	// AnalyzeImage 传入图片 URL 和提示词，返回模型原始文本响应
	AnalyzeImage(ctx context.Context, imageURL, prompt string) (string, error)
	// GetEmbedding 传入文本，返回其高维向量表示 (如 1536 维)
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
}

// Config LLM 配置
type Config struct {
	Provider string `mapstructure:"provider"` // "qwen" | "doubao" | "glm"
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
}

// New 根据 provider 类型创建对应的实现
func New(cfg Config) (Provider, error) {
	switch cfg.Provider {
	case "qwen":
		return NewQwenProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}
}
