package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ModerationRequest matches the expected input structure of the Moderation service
type ModerationRequest struct {
	Text     string `json:"text,omitempty"`     // Use omitempty if fields are optional
	ImageURL string `json:"imageUrl,omitempty"` // JSON field name matches Node.js convention
}

// ModerationResponse matches the expected output structure of the Moderation service
type ModerationResponse struct {
	IsAcceptable    bool     `json:"is_acceptable"`
	Flags           []string `json:"flags"`
	Details         string   `json:"details"`
	ConfidenceScore float64  `json:"confidence_score"`
}

// ModerationClient holds configuration for the client
type ModerationClient struct {
	BaseURL    string
	HttpClient *http.Client
}

// NewModerationClient creates a new client instance
func NewModerationClient(baseURL string) *ModerationClient {
	return &ModerationClient{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 15 * time.Second, // Moderation might take longer
		},
	}
}

// ModerateContent sends a request to the moderation service
func (c *ModerationClient) ModerateContent(text, imageURL string) (*ModerationResponse, error) {
	requestPayload := ModerationRequest{
		Text:     text,
		ImageURL: imageURL,
	}
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Error marshalling moderation request payload: %v", err)
		return nil, fmt.Errorf("failed to create request payload: %w", err)
	}

	reqUrl := fmt.Sprintf("%s/moderate", c.BaseURL) // Ensure BaseURL doesn't have a trailing slash
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request to moderation service: %v", err)
		return nil, fmt.Errorf("failed to create moderation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add trace IDs etc.

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Printf("Error sending request to moderation service at %s: %v", reqUrl, err)
		return nil, fmt.Errorf("moderation service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Moderation service returned non-OK status: %d", resp.StatusCode)
		// Optionally read error body
		return nil, fmt.Errorf("moderation service returned status %d", resp.StatusCode)
	}

	var moderationResp ModerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&moderationResp); err != nil {
		log.Printf("Error decoding moderation service response: %v", err)
		return nil, fmt.Errorf("failed to decode moderation response: %w", err)
	}

	log.Printf("Successfully received moderation result from service.")
	return &moderationResp, nil
}
