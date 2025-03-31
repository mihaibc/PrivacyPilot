package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Define task types known by the coordinator
const (
	TaskTypeAnonymizeText = "anonymize_text"
	// Add other task types later: "moderate_text", "moderate_image", etc.
)

// AICoordinatorRequest defines the structure for requests TO the coordinator
type AICoordinatorRequest struct {
	TaskType string            `json:"task_type"`        // e.g., TaskTypeAnonymizeText
	Payload  interface{}       `json:"payload"`          // The data for the task (e.g., map[string]string{"text": "..."})
	Config   map[string]string `json:"config,omitempty"` // Optional: specific model hints, etc.
}

// AICoordinatorResponse defines the generic structure for responses FROM the coordinator
// The actual 'Result' will vary based on the task type.
type AICoordinatorResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"` // Use map[string]interface{} or specific structs
	Error   string      `json:"error,omitempty"`
}

// AnonymizeTextResult defines the expected structure within the 'Result' field for anonymization tasks
type AnonymizeTextResult struct {
	AnonymizedText string `json:"anonymized_text"`
	// Add other fields returned by the specific AI adapter via the coordinator if needed
}

// AICoordinatorClient holds configuration
type AICoordinatorClient struct {
	BaseURL    string
	HttpClient *http.Client
}

// NewAICoordinatorClient creates a new client instance
func NewAICoordinatorClient(baseURL string) *AICoordinatorClient {
	return &AICoordinatorClient{
		BaseURL: baseURL,
		HttpClient: &http.Client{
			Timeout: 20 * time.Second, // AI tasks might take longer
		},
	}
}

// RequestAnonymization sends an anonymization task request to the AI Coordinator
func (c *AICoordinatorClient) RequestAnonymization(text string) (*AnonymizeTextResult, error) {
	// The payload specific to the anonymize_text task
	payload := map[string]string{"text": text}

	coordReq := AICoordinatorRequest{
		TaskType: TaskTypeAnonymizeText,
		Payload:  payload,
		// Config: map[string]string{"model_preference": "ollama-fast"}, // Example optional config
	}

	payloadBytes, err := json.Marshal(coordReq)
	if err != nil {
		log.Printf("Error marshalling AI coordinator request payload: %v", err)
		return nil, fmt.Errorf("failed to create AI coordinator request payload: %w", err)
	}

	// Assuming the coordinator has a single endpoint like /process
	reqUrl := fmt.Sprintf("%s/process", c.BaseURL)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request to AI coordinator: %v", err)
		return nil, fmt.Errorf("failed to create AI coordinator request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Printf("Error sending request to AI coordinator at %s: %v", reqUrl, err)
		return nil, fmt.Errorf("AI coordinator request failed: %w", err)
	}
	defer resp.Body.Close()

	// Decode the generic response first
	var coordResp AICoordinatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&coordResp); err != nil {
		log.Printf("Error decoding AI coordinator generic response (status %d): %v", resp.StatusCode, err)
		// Return error even if status is 200 but body is unparsable
		return nil, fmt.Errorf("failed to decode AI coordinator response: %w", err)
	}

	// Check for non-OK status codes OR explicit failure in the response body
	if resp.StatusCode != http.StatusOK || !coordResp.Success {
		errorMsg := coordResp.Error
		if errorMsg == "" {
			errorMsg = fmt.Sprintf("AI Coordinator returned status %d with success=false", resp.StatusCode)
		}
		log.Printf("AI Coordinator request failed: %s", errorMsg)
		return nil, fmt.Errorf("AI coordinator task failed: %s", errorMsg)
	}

	// If successful, parse the specific result type (AnonymizeTextResult)
	// The Result field is likely a map[string]interface{} after generic decoding,
	// so we marshal it back to bytes and unmarshal into the specific struct.
	resultBytes, err := json.Marshal(coordResp.Result)
	if err != nil {
		log.Printf("Error marshalling coordinator result field: %v", err)
		return nil, fmt.Errorf("failed to process coordinator result structure: %w", err)
	}

	var anonymizeResult AnonymizeTextResult
	if err := json.Unmarshal(resultBytes, &anonymizeResult); err != nil {
		log.Printf("Error unmarshalling specific AnonymizeTextResult from coordinator response: %v", err)
		return nil, fmt.Errorf("failed to decode specific anonymization result: %w", err)
	}

	log.Printf("Successfully received anonymization result via AI Coordinator.")
	return &anonymizeResult, nil
}
