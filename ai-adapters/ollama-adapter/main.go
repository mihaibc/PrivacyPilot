package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api"
)

// Configuration
var (
	ollamaHost         string // e.g., "http://ollama:11434"
	defaultOllamaModel string
	ollamaClient       *api.Client // Use the official client
)

// Request structure for this adapter's endpoint
type AdapterAnonymizeRequest struct {
	Text  string `json:"text" binding:"required"`
	Model string `json:"model,omitempty"` // Optional: Model override from coordinator
}

// Response structure for this adapter's endpoint
type AdapterAnonymizeResponse struct {
	AnonymizedText string `json:"anonymized_text"`
	ModelUsed      string `json:"model_used"` // Return the actual model used
}

func main() {
	// --- Configuration ---
	ollamaHost = os.Getenv("OLLAMA_API_URL") // Expecting URL like http://ollama:11434
	if ollamaHost == "" {
		ollamaHost = "http://host.docker.internal:11434" // Default for compose environment
		log.Printf("Warning: OLLAMA_API_URL not set, defaulting to %s", ollamaHost)
	}

	defaultOllamaModel = os.Getenv("OLLAMA_ANONYMIZE_MODEL")
	if defaultOllamaModel == "" {
		defaultOllamaModel = "mistral:7b"
		log.Printf("Warning: OLLAMA_ANONYMIZE_MODEL not set, defaulting to %s", defaultOllamaModel)
	}

	// --- Initialize Ollama Client ---
	var err error
	ollamaClient, err = newOllamaClient(ollamaHost)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}
	// Test connection during startup (optional but good)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ollamaClient.Heartbeat(ctx); err != nil {
		log.Printf("Warning: Could not connect to Ollama at %s: %v", ollamaHost, err)
		// Decide if this should be fatal or just a warning
	} else {
		log.Printf("Successfully connected to Ollama at %s", ollamaHost)
	}
	// ------------------------------

	// --- Gin Setup ---
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)
	router := gin.Default()

	// --- Routes ---
	router.GET("/health", healthCheckHandler)
	router.POST("/anonymize", anonymizeTextHandler)

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("Ollama Adapter Service starting on port %s", port)
	log.Printf("--> Targeting Ollama API at: %s", ollamaHost)
	log.Printf("--> Default Ollama Model: %s", defaultOllamaModel)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start Ollama Adapter Service: %v", err)
	}
}

// Helper to create Ollama client from host URL
func newOllamaClient(host string) (*api.Client, error) {
	// Parse the URL to extract scheme, host, and port
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama API URL '%s': %w", host, err)
	}

	// Create the client using the parsed URL parts
	client := api.NewClient(parsedURL, http.DefaultClient) // Use default HTTP client or configure one
	return client, nil
}

func healthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	err := ollamaClient.Heartbeat(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":       "Unavailable",
			"service":      "Ollama Adapter Service",
			"ollama_error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":             "OK",
		"service":            "Ollama Adapter Service",
		"default_model":      defaultOllamaModel,
		"ollama_host_status": "Reachable",
	})
}

func anonymizeTextHandler(c *gin.Context) {
	var req AdapterAnonymizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ollama Adapter: Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Determine model to use: request override or default
	modelToUse := req.Model
	if modelToUse == "" {
		modelToUse = defaultOllamaModel
	}

	// --- Call Ollama using Go Client ---
	log.Printf("Ollama Adapter: Requesting anonymization from model '%s'", modelToUse)
	anonymizedText, err := callOllamaAnonymize(c.Request.Context(), req.Text, modelToUse)
	if err != nil {
		log.Printf("Ollama Adapter: Error calling Ollama model '%s': %v", modelToUse, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to anonymize via Ollama model %s: %v", modelToUse, err)})
		return
	}
	// -----------------------------------

	resp := AdapterAnonymizeResponse{
		AnonymizedText: anonymizedText,
		ModelUsed:      modelToUse, // Report which model was actually used
	}
	c.JSON(http.StatusOK, resp)
}

// callOllamaAnonymize uses the official Ollama client library
func callOllamaAnonymize(ctx context.Context, textToAnonymize string, modelName string) (string, error) {
	systemPrompt := "You are an expert text anonymizer. Your task is to identify and replace Personal Identifiable Information (PII) in the provided text with placeholders like [NAME], [EMAIL], [PHONE], [ADDRESS], [CREDIT_CARD], [SSN], etc. Only output the anonymized text, without any introductory phrases, explanations, or markdown formatting. Preserve the original structure and non-sensitive parts of the text."
	prompt := fmt.Sprintf("Anonymize the following text:\n\n\"%s\"", textToAnonymize)

	ollamaReq := api.GenerateRequest{
		Model:  modelName,
		Prompt: prompt,
		System: systemPrompt,
		Stream: new(bool), // Pointer to false for non-streaming
		Options: map[string]interface{}{
			"temperature": 0.2,
		},
	}
	*ollamaReq.Stream = false // Explicitly set stream to false

	var responseBuilder strings.Builder
	var lastResponse *api.GenerateResponse // Keep track of the final response object

	// Use the ResponseFunc to process the response. Even with stream=false,
	// the library might use this pattern.
	responseFunc := func(resp api.GenerateResponse) error {
		// Accumulate response content if needed (though for stream=false, it should be in one go)
		responseBuilder.WriteString(resp.Response)
		lastResponse = &resp // Store the latest response object
		if resp.Done {
			log.Println("Ollama Adapter: Received 'done' signal from Ollama.")
		}
		return nil // Return nil to continue processing
	}

	// Add a timeout to the context specifically for the Generate call
	generateCtx, cancel := context.WithTimeout(ctx, 55*time.Second) // Slightly less than HTTP client timeout
	defer cancel()

	err := ollamaClient.Generate(generateCtx, &ollamaReq, responseFunc)
	if err != nil {
		return "", fmt.Errorf("ollama client generate call failed: %w", err)
	}

	if lastResponse == nil {
		return "", fmt.Errorf("no response received from ollama model '%s'", modelName)
	}

	if lastResponse.Error != "" { // Check for error field within the response struct
		return "", fmt.Errorf("ollama model '%s' returned error: %s", modelName, lastResponse.Error)
	}

	// The accumulated response (or the last response's content if non-streaming)
	anonymizedResult := strings.TrimSpace(responseBuilder.String())
	if anonymizedResult == "" && lastResponse.Response != "" {
		// Fallback if builder didn't capture but last response has content
		anonymizedResult = strings.TrimSpace(lastResponse.Response)
	}

	if anonymizedResult == "" {
		log.Printf("Warning: Ollama model '%s' returned an empty response string.", modelName)
		// Return empty string, but log warning. Or return an error? Depends on requirements.
	}

	log.Printf("Ollama Adapter: Successfully received response from model '%s'.", modelName)
	return anonymizedResult, nil
}
