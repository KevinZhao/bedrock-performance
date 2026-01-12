package bedrock

import "time"

// InvokeResult contains the result of a Bedrock API invocation
type InvokeResult struct {
	Success         bool
	StartTime       time.Time
	EndTime         time.Time
	TTFT            time.Duration // Time to first token (only for streaming)
	InputTokens     int
	OutputTokens    int
	Error           error
	ErrorType       string
	HTTPStatusCode  int    // HTTP response status code
	ResponseContent string
}

// Duration returns the total duration of the request
func (r *InvokeResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// ClaudeRequest represents a request to Claude models
type ClaudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	Messages         []ClaudeMessage `json:"messages"`
	Temperature      float64         `json:"temperature,omitempty"`
}

// ClaudeMessage represents a message in Claude request
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents a response from Claude models
type ClaudeResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []ClaudeContent     `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason"`
	Usage        ClaudeUsage         `json:"usage"`
}

// ClaudeContent represents content in Claude response
type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ClaudeUsage represents token usage in Claude response
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeStreamEvent represents a streaming event from Claude
type ClaudeStreamEvent struct {
	Type         string              `json:"type"`
	Index        int                 `json:"index,omitempty"`
	Delta        *ClaudeStreamDelta  `json:"delta,omitempty"`
	ContentBlock *ClaudeContent      `json:"content_block,omitempty"`
	Message      *ClaudeResponse     `json:"message,omitempty"`
	Usage        *ClaudeUsage        `json:"usage,omitempty"`
}

// ClaudeStreamDelta represents a delta in streaming response
type ClaudeStreamDelta struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

// LlamaRequest represents a request to Llama models
type LlamaRequest struct {
	Prompt      string  `json:"prompt"`
	MaxGenLen   int     `json:"max_gen_len"`
	Temperature float64 `json:"temperature,omitempty"`
}

// LlamaResponse represents a response from Llama models
type LlamaResponse struct {
	Generation           string `json:"generation"`
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
	StopReason           string `json:"stop_reason"`
}

// DeepSeekResponse represents a response from DeepSeek models (OpenAI format)
type DeepSeekResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []DeepSeekChoice       `json:"choices"`
	Usage   DeepSeekUsage          `json:"usage"`
}

// DeepSeekChoice represents a choice in DeepSeek response
type DeepSeekChoice struct {
	Index        int                   `json:"index"`
	Message      DeepSeekMessage       `json:"message"`
	FinishReason string                `json:"finish_reason"`
	Delta        *DeepSeekStreamDelta  `json:"delta,omitempty"`
}

// DeepSeekMessage represents a message in DeepSeek response
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekUsage represents token usage in DeepSeek response
type DeepSeekUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// DeepSeekStreamDelta represents a streaming delta from DeepSeek
type DeepSeekStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// DeepSeekStreamEvent represents a streaming event from DeepSeek (OpenAI format)
type DeepSeekStreamEvent struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []DeepSeekChoice `json:"choices"`
	Usage   *DeepSeekUsage   `json:"usage,omitempty"`
}
