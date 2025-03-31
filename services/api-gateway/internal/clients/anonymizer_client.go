package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// AnonymizerRequest matches the expected input structure of the Anonymizer service
type AnonymizerRequest struct {
	Text string `json:"text"`
}

// AnonymizerResponse matches the expected output structure of the Anonymizer service
type AnonymizerResponse struct {
	OriginalText   string `json:"original_text"`
	AnonymizedText string `json:"anonymized_text"`
	// Add other fields if the anonymizer service returns more details
}

// AnonymizerClient holds configuration for the client
type AnonymizerClient struct {
	BaseURL    string
	HttpClient *http.Client
}

// NewAnonymizerClient creates a new client instance
func NewAnonymizerClient(baseURL string) *AnonymizerClient {
	return &AnonymizerClient{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second, // Sensible default timeout
		},
	}
}

// AnonymizeText sends a request to the anonymizer service
func (c *AnonymizerClient) AnonymizeText(text string) (*AnonymizerResponse, error) {
	requestPayload := AnonymizerRequest{Text: text}
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Error marshalling anonymizer request payload: %v", err)
		return nil, fmt.Errorf("failed to create request payload: %w", err)
	}

	reqUrl := fmt.Sprintf("%s/anonymize", c.BaseURL) // Ensure BaseURL doesn't have a trailing slash
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request to anonymizer service: %v", err)
		return nil, fmt.Errorf("failed to create anonymizer request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add other headers like trace IDs if implementing tracing

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Printf("Error sending request to anonymizer service at %s: %v", reqUrl, err)
		return nil, fmt.Errorf("anonymizer service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Attempt to read error body for more context (optional)
		// var errorBody map[string]interface{}
		// _ = json.NewDecoder(resp.Body).Decode(&errorBody) // Ignore decode error here
		log.Printf("Anonymizer service returned non-OK status: %d", resp.StatusCode)
		return nil, fmt.Errorf("anonymizer service returned status %d", resp.StatusCode)
	}

	var anonymizerResp AnonymizerResponse
	if err := json.NewDecoder(resp.Body).Decode(&anonymizerResp); err != nil {
		log.Printf("Error decoding anonymizer service response: %v", err)
		return nil, fmt.Errorf("failed to decode anonymizer response: %w", err)
	}

	log.Printf("Successfully received anonymized text from service.")
	return &anonymizerResp, nil
}
