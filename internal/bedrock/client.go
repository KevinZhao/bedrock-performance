package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// ServiceTier represents the Bedrock service tier
type ServiceTier string

const (
	ServiceTierDefault  ServiceTier = "default"
	ServiceTierPriority ServiceTier = "priority"
	ServiceTierFlex     ServiceTier = "flex"
)

// ClientConfig holds the configuration needed to create a Bedrock client
type ClientConfig struct {
	Region      string
	AccessKey   string
	SecretKey   string
	ModelID     string
	MaxTokens   int
	Temperature float64
	ServiceTier ServiceTier
}

// Client wraps the AWS Bedrock Runtime client
type Client struct {
	client      *bedrockruntime.Client
	modelID     string
	maxTokens   int
	temperature float64
	serviceTier ServiceTier
}

// NewClient creates a new Bedrock client
func NewClient(region, accessKey, secretKey, modelID string, maxTokens int, temperature float64) *Client {
	var cfg aws.Config
	var err error

	// If credentials are empty, use default credential chain (env, shared credentials, IAM role, etc.)
	if accessKey == "" || secretKey == "" {
		cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load default AWS config: %v\n", err)
			// Fallback to empty config
			cfg = aws.Config{Region: region}
		}
	} else {
		cfg = aws.Config{
			Region: region,
			Credentials: credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		}
	}

	return &Client{
		client:      bedrockruntime.NewFromConfig(cfg),
		modelID:     modelID,
		maxTokens:   maxTokens,
		temperature: temperature,
		serviceTier: ServiceTierDefault,
	}
}

// NewClientFromConfig creates a new Bedrock client from ClientConfig
func NewClientFromConfig(cfg *ClientConfig) *Client {
	client := NewClient(cfg.Region, cfg.AccessKey, cfg.SecretKey, cfg.ModelID, cfg.MaxTokens, cfg.Temperature)
	if cfg.ServiceTier != "" {
		client.serviceTier = cfg.ServiceTier
	}
	return client
}

// InvokeNonStreaming invokes the model without streaming
func (c *Client) InvokeNonStreaming(ctx context.Context, prompt string) *InvokeResult {
	result := &InvokeResult{
		StartTime: time.Now(),
	}

	// Prepare request body based on model type
	var requestBody []byte
	var err error

	if c.isClaudeModel() {
		requestBody, err = c.prepareClaudeRequest(prompt)
	} else if c.isDeepSeekModel() {
		requestBody, err = c.prepareDeepSeekRequest(prompt)
	} else if c.isMistralModel() {
		requestBody, err = c.prepareMistralRequest(prompt)
	} else if c.isQwenModel() {
		requestBody, err = c.prepareQwenRequest(prompt)
	} else if c.isLlamaModel() {
		requestBody, err = c.prepareLlamaRequest(prompt)
	} else {
		result.Error = fmt.Errorf("unsupported model: %s", c.modelID)
		result.ErrorType = "UnsupportedModel"
		result.EndTime = time.Now()
		return result
	}

	if err != nil {
		result.Error = fmt.Errorf("failed to prepare request: %w", err)
		result.ErrorType = "RequestPreparationError"
		result.EndTime = time.Now()
		return result
	}

	// Invoke the model
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestBody,
	}

	output, err := c.client.InvokeModel(ctx, input)
	result.EndTime = time.Now()

	if err != nil {
		result.Error = err
		result.ErrorType = c.categorizeError(err)
		result.HTTPStatusCode = c.extractHTTPStatusFromError(err)
		return result
	}

	// Set HTTP status code (200 for successful invocation)
	result.HTTPStatusCode = 200

	// Parse response based on model type
	if c.isClaudeModel() {
		c.parseClaudeResponse(output.Body, result)
	} else if c.isDeepSeekModel() {
		c.parseDeepSeekResponse(output.Body, result)
	} else if c.isMistralModel() {
		c.parseMistralResponse(output.Body, result)
	} else if c.isQwenModel() {
		c.parseQwenResponse(output.Body, result)
	} else if c.isLlamaModel() {
		c.parseLlamaResponse(output.Body, result)
	}

	result.Success = result.Error == nil
	return result
}

// InvokeStreaming invokes the model with streaming
func (c *Client) InvokeStreaming(ctx context.Context, prompt string) *InvokeResult {
	result := &InvokeResult{
		StartTime: time.Now(),
	}

	// Prepare request body (Claude, DeepSeek, Mistral, and Qwen support streaming)
	if !c.isClaudeModel() && !c.isDeepSeekModel() && !c.isMistralModel() && !c.isQwenModel() {
		result.Error = fmt.Errorf("streaming only supported for Claude, DeepSeek, Mistral, and Qwen models")
		result.ErrorType = "UnsupportedOperation"
		result.EndTime = time.Now()
		return result
	}

	var requestBody []byte
	var err error
	if c.isClaudeModel() {
		requestBody, err = c.prepareClaudeRequest(prompt)
	} else if c.isMistralModel() {
		requestBody, err = c.prepareMistralRequest(prompt)
	} else if c.isQwenModel() {
		requestBody, err = c.prepareQwenRequest(prompt)
	} else {
		requestBody, err = c.prepareDeepSeekRequest(prompt)
	}
	if err != nil {
		result.Error = fmt.Errorf("failed to prepare request: %w", err)
		result.ErrorType = "RequestPreparationError"
		result.EndTime = time.Now()
		return result
	}

	// Invoke the model with streaming
	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(c.modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestBody,
	}

	// Set service tier if specified
	if c.serviceTier != "" && c.serviceTier != ServiceTierDefault {
		input.ServiceTier = types.ServiceTierType(c.serviceTier)
	}

	output, err := c.client.InvokeModelWithResponseStream(ctx, input)
	if err != nil {
		result.Error = err
		result.ErrorType = c.categorizeError(err)
		result.HTTPStatusCode = c.extractHTTPStatusFromError(err)
		result.EndTime = time.Now()
		return result
	}

	// Set HTTP status code (200 for successful streaming start)
	result.HTTPStatusCode = 200

	// Process the stream based on model type
	if c.isClaudeModel() {
		c.processClaudeStream(output.GetStream(), result)
	} else if c.isMistralModel() {
		c.processMistralStream(output.GetStream(), result)
	} else if c.isQwenModel() {
		c.processQwenStream(output.GetStream(), result)
	} else if c.isDeepSeekModel() {
		c.processDeepSeekStream(output.GetStream(), result)
	}

	result.EndTime = time.Now()
	result.Success = result.Error == nil
	return result
}

// prepareClaudeRequest prepares a request for Claude models
func (c *Client) prepareClaudeRequest(prompt string) ([]byte, error) {
	req := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        c.maxTokens,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	}
	return json.Marshal(req)
}

// prepareDeepSeekRequest prepares a request for DeepSeek models (without anthropic_version)
func (c *Client) prepareDeepSeekRequest(prompt string) ([]byte, error) {
	// DeepSeek uses Messages API like Claude but without the anthropic_version field
	req := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
	}
	return json.Marshal(req)
}

// prepareLlamaRequest prepares a request for Llama models
func (c *Client) prepareLlamaRequest(prompt string) ([]byte, error) {
	req := LlamaRequest{
		Prompt:      prompt,
		MaxGenLen:   c.maxTokens,
		Temperature: c.temperature,
	}
	return json.Marshal(req)
}

// parseClaudeResponse parses a Claude response
func (c *Client) parseClaudeResponse(body []byte, result *InvokeResult) {
	var resp ClaudeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		result.Error = fmt.Errorf("failed to parse response: %w", err)
		result.ErrorType = "ResponseParseError"
		return
	}

	result.InputTokens = resp.Usage.InputTokens
	result.OutputTokens = resp.Usage.OutputTokens

	if len(resp.Content) > 0 {
		result.ResponseContent = resp.Content[0].Text
	}
}

// parseDeepSeekResponse parses a DeepSeek response (OpenAI format)
func (c *Client) parseDeepSeekResponse(body []byte, result *InvokeResult) {
	var resp DeepSeekResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		result.Error = fmt.Errorf("failed to parse response: %w", err)
		result.ErrorType = "ResponseParseError"
		return
	}

	result.InputTokens = resp.Usage.PromptTokens
	result.OutputTokens = resp.Usage.CompletionTokens

	if len(resp.Choices) > 0 {
		result.ResponseContent = resp.Choices[0].Message.Content
	}
}

// parseLlamaResponse parses a Llama response
func (c *Client) parseLlamaResponse(body []byte, result *InvokeResult) {
	var resp LlamaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		result.Error = fmt.Errorf("failed to parse response: %w", err)
		result.ErrorType = "ResponseParseError"
		return
	}

	result.InputTokens = resp.PromptTokenCount
	result.OutputTokens = resp.GenerationTokenCount
	result.ResponseContent = resp.Generation
}

// processClaudeStream processes a Claude streaming response
func (c *Client) processClaudeStream(stream *bedrockruntime.InvokeModelWithResponseStreamEventStream, result *InvokeResult) {
	firstToken := true
	var contentBuilder strings.Builder

	for event := range stream.Events() {
		switch e := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var streamEvent ClaudeStreamEvent
			if err := json.Unmarshal(e.Value.Bytes, &streamEvent); err != nil {
				result.Error = fmt.Errorf("failed to parse stream event: %w", err)
				result.ErrorType = "StreamParseError"
				return
			}

			// Record TTFT on first content
			if firstToken && streamEvent.Type == "content_block_delta" {
				result.TTFT = time.Since(result.StartTime)
				firstToken = false
			}

			// Accumulate content
			if streamEvent.Delta != nil && streamEvent.Delta.Text != "" {
				contentBuilder.WriteString(streamEvent.Delta.Text)
			}

			// Capture usage information
			if streamEvent.Type == "message_delta" && streamEvent.Usage != nil {
				result.OutputTokens = streamEvent.Usage.OutputTokens
			}

			if streamEvent.Type == "message_start" && streamEvent.Message != nil {
				result.InputTokens = streamEvent.Message.Usage.InputTokens
			}

		default:
			// Handle any error types that are not chunk events
			// AWS SDK error types are handled here
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		if err != io.EOF {
			// Log full error for debugging StreamError cases
			fmt.Fprintf(os.Stderr, "[DEBUG] StreamError details: %v\n", err)
			result.Error = fmt.Errorf("stream error: %w", err)
			result.ErrorType = "StreamError"
			return
		}
	}

	result.ResponseContent = contentBuilder.String()
}

// processDeepSeekStream processes a DeepSeek streaming response (OpenAI format with AWS Bedrock metrics)
func (c *Client) processDeepSeekStream(stream *bedrockruntime.InvokeModelWithResponseStreamEventStream, result *InvokeResult) {
	firstToken := true
	var contentBuilder strings.Builder

	for event := range stream.Events() {
		switch e := event.(type) {
		case *types.ResponseStreamMemberChunk:
			// Parse as generic JSON to access all fields
			var genericEvent map[string]interface{}
			if err := json.Unmarshal(e.Value.Bytes, &genericEvent); err == nil {
				// Check for AWS Bedrock invocation metrics (in final chunk)
				if metrics, ok := genericEvent["amazon-bedrock-invocationMetrics"].(map[string]interface{}); ok {
					if inputTokens, ok := metrics["inputTokenCount"].(float64); ok {
						result.InputTokens = int(inputTokens)
					}
					if outputTokens, ok := metrics["outputTokenCount"].(float64); ok {
						result.OutputTokens = int(outputTokens)
					}
				}

				// Check for content in choices
				if choices, ok := genericEvent["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok && content != "" {
								if firstToken {
									result.TTFT = time.Since(result.StartTime)
									firstToken = false
								}
								contentBuilder.WriteString(content)
							}
						}
					}
				}
			}

		default:
			// Handle any error types that are not chunk events
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		if err != io.EOF {
			// Log full error for debugging StreamError cases
			fmt.Fprintf(os.Stderr, "[DEBUG] StreamError (DeepSeek) details: %v\n", err)
			result.Error = fmt.Errorf("stream error: %w", err)
			result.ErrorType = "StreamError"
			return
		}
	}

	result.ResponseContent = contentBuilder.String()
}

// isClaudeModel checks if the model is a Claude model
func (c *Client) isClaudeModel() bool {
	return strings.Contains(strings.ToLower(c.modelID), "claude") ||
		strings.Contains(strings.ToLower(c.modelID), "anthropic")
}

// isDeepSeekModel checks if the model is a DeepSeek model
func (c *Client) isDeepSeekModel() bool {
	return strings.Contains(strings.ToLower(c.modelID), "deepseek")
}

// isLlamaModel checks if the model is a Llama model
func (c *Client) isLlamaModel() bool {
	return strings.Contains(strings.ToLower(c.modelID), "llama") ||
		strings.Contains(strings.ToLower(c.modelID), "meta")
}

// isMistralModel checks if the model is a Mistral model
func (c *Client) isMistralModel() bool {
	return strings.Contains(strings.ToLower(c.modelID), "mistral") ||
		strings.Contains(strings.ToLower(c.modelID), "mixtral")
}

// isQwenModel checks if the model is a Qwen model
func (c *Client) isQwenModel() bool {
	return strings.Contains(strings.ToLower(c.modelID), "qwen")
}

// prepareQwenRequest prepares a request for Qwen models (OpenAI-compatible format)
func (c *Client) prepareQwenRequest(prompt string) ([]byte, error) {
	req := map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
	}
	return json.Marshal(req)
}

// parseQwenResponse parses a Qwen response (OpenAI format)
func (c *Client) parseQwenResponse(body []byte, result *InvokeResult) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		result.Error = fmt.Errorf("failed to parse response: %w", err)
		result.ErrorType = "ResponseParseError"
		return
	}

	// Check for choices array (OpenAI format)
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					result.ResponseContent = content
				}
			}
		}
	}

	// Token usage from usage field or Bedrock metrics
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			result.InputTokens = int(promptTokens)
		}
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			result.OutputTokens = int(completionTokens)
		}
	}

	// Also check for Bedrock metrics
	if metrics, ok := resp["amazon-bedrock-invocationMetrics"].(map[string]interface{}); ok {
		if inputTokens, ok := metrics["inputTokenCount"].(float64); ok {
			result.InputTokens = int(inputTokens)
		}
		if outputTokens, ok := metrics["outputTokenCount"].(float64); ok {
			result.OutputTokens = int(outputTokens)
		}
	}
}

// processQwenStream processes a Qwen streaming response (OpenAI format)
func (c *Client) processQwenStream(stream *bedrockruntime.InvokeModelWithResponseStreamEventStream, result *InvokeResult) {
	firstToken := true
	var contentBuilder strings.Builder

	for event := range stream.Events() {
		switch e := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var genericEvent map[string]interface{}
			if err := json.Unmarshal(e.Value.Bytes, &genericEvent); err == nil {
				// Check for AWS Bedrock invocation metrics (in final chunk)
				if metrics, ok := genericEvent["amazon-bedrock-invocationMetrics"].(map[string]interface{}); ok {
					if inputTokens, ok := metrics["inputTokenCount"].(float64); ok {
						result.InputTokens = int(inputTokens)
					}
					if outputTokens, ok := metrics["outputTokenCount"].(float64); ok {
						result.OutputTokens = int(outputTokens)
					}
				}

				// Check for content in choices (OpenAI streaming format)
				if choices, ok := genericEvent["choices"].([]interface{}); ok && len(choices) > 0 {
					if choice, ok := choices[0].(map[string]interface{}); ok {
						if delta, ok := choice["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok && content != "" {
								if firstToken {
									result.TTFT = time.Since(result.StartTime)
									firstToken = false
								}
								contentBuilder.WriteString(content)
							}
						}
					}
				}
			}

		default:
			// Handle any error types that are not chunk events
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		if err != io.EOF {
			fmt.Fprintf(os.Stderr, "[DEBUG] StreamError (Qwen) details: %v\n", err)
			result.Error = fmt.Errorf("stream error: %w", err)
			result.ErrorType = "StreamError"
			return
		}
	}

	result.ResponseContent = contentBuilder.String()
}

// prepareMistralRequest prepares a request for Mistral models
func (c *Client) prepareMistralRequest(prompt string) ([]byte, error) {
	req := map[string]interface{}{
		"prompt":      "<s>[INST] " + prompt + " [/INST]",
		"max_tokens":  c.maxTokens,
		"temperature": c.temperature,
	}
	return json.Marshal(req)
}

// parseMistralResponse parses a Mistral response
func (c *Client) parseMistralResponse(body []byte, result *InvokeResult) {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		result.Error = fmt.Errorf("failed to parse response: %w", err)
		result.ErrorType = "ResponseParseError"
		return
	}

	// Mistral response format: {"outputs": [{"text": "...", "stop_reason": "..."}]}
	if outputs, ok := resp["outputs"].([]interface{}); ok && len(outputs) > 0 {
		if output, ok := outputs[0].(map[string]interface{}); ok {
			if text, ok := output["text"].(string); ok {
				result.ResponseContent = text
			}
		}
	}

	// Token usage from Bedrock metrics
	if metrics, ok := resp["amazon-bedrock-invocationMetrics"].(map[string]interface{}); ok {
		if inputTokens, ok := metrics["inputTokenCount"].(float64); ok {
			result.InputTokens = int(inputTokens)
		}
		if outputTokens, ok := metrics["outputTokenCount"].(float64); ok {
			result.OutputTokens = int(outputTokens)
		}
	}
}

// processMistralStream processes a Mistral streaming response
func (c *Client) processMistralStream(stream *bedrockruntime.InvokeModelWithResponseStreamEventStream, result *InvokeResult) {
	firstToken := true
	var contentBuilder strings.Builder

	for event := range stream.Events() {
		switch e := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var genericEvent map[string]interface{}
			if err := json.Unmarshal(e.Value.Bytes, &genericEvent); err == nil {
				// Check for AWS Bedrock invocation metrics (in final chunk)
				if metrics, ok := genericEvent["amazon-bedrock-invocationMetrics"].(map[string]interface{}); ok {
					if inputTokens, ok := metrics["inputTokenCount"].(float64); ok {
						result.InputTokens = int(inputTokens)
					}
					if outputTokens, ok := metrics["outputTokenCount"].(float64); ok {
						result.OutputTokens = int(outputTokens)
					}
				}

				// Check for outputs array (Mistral format)
				if outputs, ok := genericEvent["outputs"].([]interface{}); ok && len(outputs) > 0 {
					if output, ok := outputs[0].(map[string]interface{}); ok {
						if text, ok := output["text"].(string); ok && text != "" {
							if firstToken {
								result.TTFT = time.Since(result.StartTime)
								firstToken = false
							}
							contentBuilder.WriteString(text)
						}
					}
				}
			}

		default:
			// Handle any error types that are not chunk events
		}
	}

	// Check for stream errors
	if err := stream.Err(); err != nil {
		if err != io.EOF {
			fmt.Fprintf(os.Stderr, "[DEBUG] StreamError (Mistral) details: %v\n", err)
			result.Error = fmt.Errorf("stream error: %w", err)
			result.ErrorType = "StreamError"
			return
		}
	}

	result.ResponseContent = contentBuilder.String()
}

// categorizeError categorizes AWS errors
func (c *Client) categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "throttling") || strings.Contains(errStr, "too many"):
		return "ThrottlingError"
	case strings.Contains(errStr, "validation"):
		return "ValidationError"
	case strings.Contains(errStr, "access denied"):
		return "AccessDeniedError"
	case strings.Contains(errStr, "not found"):
		return "ModelNotFoundError"
	case strings.Contains(errStr, "service quota"):
		return "QuotaExceededError"
	case strings.Contains(errStr, "timeout"):
		return "TimeoutError"
	default:
		// Log full error for debugging UnknownError cases
		fmt.Fprintf(os.Stderr, "[DEBUG] UnknownError details: %v\n", err)
		return "UnknownError"
	}
}

// extractHTTPStatusFromError attempts to extract HTTP status code from AWS error
func (c *Client) extractHTTPStatusFromError(err error) int {
	if err == nil {
		return 200
	}

	errStr := err.Error()
	// Common AWS error status codes based on error type
	switch {
	case strings.Contains(errStr, "throttling"):
		return 429
	case strings.Contains(errStr, "validation"):
		return 400
	case strings.Contains(errStr, "access denied"):
		return 403
	case strings.Contains(errStr, "not found"):
		return 404
	case strings.Contains(errStr, "service quota"):
		return 429
	case strings.Contains(errStr, "timeout"):
		return 504
	case strings.Contains(errStr, "internal"):
		return 500
	default:
		return 0 // Unknown error, no HTTP status available
	}
}
