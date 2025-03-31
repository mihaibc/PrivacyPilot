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
	"github.com/ollama/ollama/api" // Import the official Ollama API library
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

// main function: Entry point of the service
func main() {
	// --- Configuration ---
	ollamaHost = os.Getenv("OLLAMA_API_URL") // Expecting URL like http://ollama:11434
	if ollamaHost == "" {
		ollamaHost = "http://host.docker.internal:11434" // Default for compose environment
		log.Printf("Warning: OLLAMA_API_URL not set, defaulting to %s", ollamaHost)
	}

	defaultOllamaModel = os.Getenv("OLLAMA_ANONYMIZE_MODEL")
	if defaultOllamaModel == "" {
		defaultOllamaModel = "mistral:7b" // Default model if not specified
		log.Printf("Warning: OLLAMA_ANONYMIZE_MODEL not set, defaulting to %s", defaultOllamaModel)
	}

	// --- Initialize Ollama Client ---
	var err error
	ollamaClient, err = newOllamaClient(ollamaHost)
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}
	// Test connection during startup (optional but recommended)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ollamaClient.Heartbeat(ctx); err != nil {
		log.Printf("Warning: Could not connect to Ollama at %s during startup: %v", ollamaHost, err)
		// Depending on requirements, you might choose to exit here if connection is critical
	} else {
		log.Printf("Successfully connected to Ollama at %s during startup", ollamaHost)
	}
	// ------------------------------

	// --- Gin Setup ---
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)
	router := gin.Default() // Includes logger and recovery middleware

	// --- Routes ---
	router.GET("/health", healthCheckHandler)
	router.POST("/anonymize", anonymizeTextHandler) // Endpoint for AI Coordinator to call

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084" // Default port for Ollama adapter
	}
	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("Ollama Adapter Service starting on port %s", port)
	log.Printf("--> Targeting Ollama API at: %s", ollamaHost)
	log.Printf("--> Default Ollama Model: %s", defaultOllamaModel)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start Ollama Adapter Service: %v", err)
	}
}

// newOllamaClient creates an Ollama API client from a host URL string.
func newOllamaClient(host string) (*api.Client, error) {
	// Parse the URL to extract scheme, host, and port
	parsedURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama API URL '%s': %w", host, err)
	}

	// Create the client using the parsed URL parts
	// You can customize the underlying http.Client if needed (timeouts, transport, etc.)
	client := api.NewClient(parsedURL, http.DefaultClient)
	return client, nil
}

// healthCheckHandler checks the connectivity to the configured Ollama instance.
func healthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second) // Use request context with timeout
	defer cancel()

	err := ollamaClient.Heartbeat(ctx)
	if err != nil {
		log.Printf("Health check failed: Could not connect to Ollama at %s: %v", ollamaHost, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":             "Unavailable",
			"service":            "Ollama Adapter Service",
			"ollama_host_status": "Unreachable",
			"error":              err.Error(),
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

// anonymizeTextHandler handles requests to anonymize text via Ollama.
func anonymizeTextHandler(c *gin.Context) {
	var req AdapterAnonymizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Ollama Adapter: Error binding JSON for /anonymize: %v", err)
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
	// Pass the request context down to the Ollama call
	anonymizedText, err := callOllamaAnonymize(c.Request.Context(), req.Text, modelToUse)
	if err != nil {
		log.Printf("Ollama Adapter: Error calling Ollama model '%s': %v", modelToUse, err)
		// Return a server error if the call failed
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to anonymize via Ollama model %s: %v", modelToUse, err)})
		return
	}
	// -----------------------------------

	// Prepare and send the successful response
	resp := AdapterAnonymizeResponse{
		AnonymizedText: anonymizedText,
		ModelUsed:      modelToUse, // Report which model was actually used
	}
	c.JSON(http.StatusOK, resp)
}

// callOllamaAnonymize constructs the prompt and calls the Ollama generate endpoint
// using the official Ollama client library. (Corrected)
func callOllamaAnonymize(ctx context.Context, textToAnonymize string, modelName string) (string, error) {
	// Define the system prompt instructing the model on its task
	systemPrompt := "You are an expert text anonymizer. Your task is to identify and replace Personal Identifiable Information (PII) in the provided text with placeholders like [NAME], [EMAIL], [PHONE], [ADDRESS], [CREDIT_CARD], [SSN], etc. Only output the anonymized text, without any introductory phrases, explanations, or markdown formatting. Preserve the original structure and non-sensitive parts of the text."

	// Define the user prompt containing the text to be processed
	prompt := fmt.Sprintf("Anonymize the following text:\n\n\"%s\"", textToAnonymize)

	// Prepare the request for the Ollama API client
	ollamaReq := api.GenerateRequest{
		Model:  modelName,
		Prompt: prompt,
		System: systemPrompt,
		Stream: new(bool), // Pointer to false for non-streaming response
		Options: map[string]interface{}{
			"temperature": 0.2, // Adjust model parameters as needed (lower temp for less creativity)
		},
	}
	*ollamaReq.Stream = false // Explicitly set stream to false

	var responseBuilder strings.Builder
	var lastResponse *api.GenerateResponse // Keep track of the final response object

	// Define the function to handle the response(s) from the Ollama client
	responseFunc := func(resp api.GenerateResponse) error {
		// Append the response content to the builder
		responseBuilder.WriteString(resp.Response)
		// Store the latest (likely only) response object for metadata access
		lastResponse = &resp
		if resp.Done {
			log.Println("Ollama Adapter: Received 'done' signal from Ollama.")
		}
		return nil // Return nil to indicate successful processing of this response part
	}

	// Add a timeout specifically for the Generate call, slightly less than the adapter's HTTP timeout
	generateCtx, cancel := context.WithTimeout(ctx, 55*time.Second)
	defer cancel()

	// Execute the generate request
	err := ollamaClient.Generate(generateCtx, &ollamaReq, responseFunc)
	if err != nil {
		// This catches errors like connection issues, model not found on Ollama server, timeouts, etc.
		return "", fmt.Errorf("ollama client generate call failed for model '%s': %w", modelName, err)
	}

	// Sanity check: Ensure the response function was actually invoked
	if lastResponse == nil {
		return "", fmt.Errorf("no response object received from ollama model '%s' despite nil error from Generate call", modelName)
	}

	// Errors *within* the model generation (if any) are usually reported via the main `err` above.
	// The GenerateResponse struct itself doesn't have a dedicated Error field in the library v0.x.

	// Extract the final text result, trimming whitespace
	anonymizedResult := strings.TrimSpace(responseBuilder.String())
	// Fallback just in case builder didn't capture but last response exists
	if anonymizedResult == "" && lastResponse.Response != "" {
		anonymizedResult = strings.TrimSpace(lastResponse.Response)
	}

	if anonymizedResult == "" {
		// Log a warning if the model returned nothing, might indicate prompt issues or model limitations
		log.Printf("Warning: Ollama model '%s' returned an empty response string.", modelName)
	}

	log.Printf("Ollama Adapter: Successfully received response from model '%s'.", modelName)
	return anonymizedResult, nil
}
