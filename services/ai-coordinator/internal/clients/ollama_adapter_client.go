package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io" // Import io
	"log"
	"net/http"
	"time"
)

// Request structure to send TO the Ollama Adapter (now includes optional model)
type OllamaAdapterAnonymizeRequest struct {
	Text  string `json:"text"`
	Model string `json:"model,omitempty"` // Optional model override
}

// Response structure received FROM the Ollama Adapter (now includes model used)
type OllamaAdapterAnonymizeResponse struct {
	AnonymizedText string `json:"anonymized_text"`
	ModelUsed      string `json:"model_used"`
}

// OllamaAdapterClient remains the same
type OllamaAdapterClient struct {
	BaseURL    string
	HttpClient *http.Client
}

// NewOllamaAdapterClient remains the same
func NewOllamaAdapterClient(baseURL string) *OllamaAdapterClient {
	if baseURL == "" {
		log.Println("Warning: Ollama Adapter URL is empty. Client created but will likely fail.")
	}
	return &OllamaAdapterClient{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 65 * time.Second,
		},
	}
}

// AnonymizeText now accepts an optional model hint
func (c *OllamaAdapterClient) AnonymizeText(payload map[string]interface{}, modelHint string) (*OllamaAdapterAnonymizeResponse, error) {
	if c.BaseURL == "" {
		return nil, fmt.Errorf("ollama adapter client not configured (URL is empty)")
	}

	text, ok := payload["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("invalid or missing 'text' field in payload for Ollama anonymization")
	}

	// Use the modelHint if provided
	adapterReq := OllamaAdapterAnonymizeRequest{
		Text:  text,
		Model: modelHint, // Pass the hint (can be empty string)
	}

	payloadBytes, err := json.Marshal(adapterReq)
	if err != nil {
		log.Printf("Error marshalling Ollama adapter request payload: %v", err)
		return nil, fmt.Errorf("failed to create adapter request payload: %w", err)
	}

	reqUrl := fmt.Sprintf("%s/anonymize", c.BaseURL)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		// ... error handling ...
		log.Printf("Error creating request to Ollama adapter: %v", err)
		return nil, fmt.Errorf("failed to create Ollama adapter request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		// ... error handling ...
		log.Printf("Error sending request to Ollama adapter at %s: %v", reqUrl, err)
		return nil, fmt.Errorf("ollama adapter request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// ... error handling ...
		bodyBytes, _ := io.ReadAll(resp.Body) // Use io.ReadAll
		log.Printf("Ollama adapter returned non-OK status: %d. Body: %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("ollama adapter returned status %d", resp.StatusCode)
	}

	var adapterResp OllamaAdapterAnonymizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&adapterResp); err != nil {
		// ... error handling ...
		log.Printf("Error decoding Ollama adapter response: %v", err)
		return nil, fmt.Errorf("failed to decode Ollama adapter response: %w", err)
	}

	log.Printf("Successfully received response from Ollama Adapter (Model Used: %s).", adapterResp.ModelUsed)
	return &adapterResp, nil
}
