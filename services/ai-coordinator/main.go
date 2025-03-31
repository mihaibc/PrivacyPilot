package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	// Use the module name defined in this service's go.mod
	"privacypilot-ai-coordinator/internal/clients"
	"privacypilot-ai-coordinator/internal/handlers"
)

func main() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	router := gin.Default()

	// --- Service Clients for AI Adapters ---
	// Initialize Ollama Client
	ollamaAdapterURL := strings.TrimRight(os.Getenv("OLLAMA_ADAPTER_URL"), "/")
	// Log a warning but don't make it fatal, client handles empty URL internally
	if ollamaAdapterURL == "" {
		log.Println("Warning: OLLAMA_ADAPTER_URL environment variable not set. Ollama functionality may be unavailable.")
	}
	ollamaClient := clients.NewOllamaAdapterClient(ollamaAdapterURL)

	// Placeholder for Azure Client initialization (when created)
	// azureAdapterURL := strings.TrimRight(os.Getenv("AZURE_AI_ADAPTER_URL"), "/")
	// if azureAdapterURL == "" { log.Println("Warning: AZURE_AI_ADAPTER_URL not set.") }
	// azureClient := clients.NewAzureAdapterClient(azureAdapterURL)

	// Placeholder for Stable Diffusion Client initialization (when created)
	// sdAdapterURL := strings.TrimRight(os.Getenv("STABLE_DIFFUSION_ADAPTER_URL"), "/")
	// if sdAdapterURL == "" { log.Println("Warning: STABLE_DIFFUSION_ADAPTER_URL not set.") }
	// sdClient := clients.NewStableDiffusionAdapterClient(sdAdapterURL)

	// --- Handlers ---
	// Instantiate the main process handler, passing all instantiated adapter clients
	processHandler := handlers.NewProcessHandler(
		ollamaClient,
		// azureClient, // Pass other clients here when available
		// sdClient,
	)

	// --- Routes ---
	router.GET("/health", healthCheckHandler)
	// Register the main processing route, handled by the ProcessHandler
	router.POST("/process", processHandler.HandleProcessRequest)

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083" // Default port for AI Coordinator service
	}
	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("AI Coordinator Service starting on port %s", port)
	// Log the configured adapter URLs for easier debugging
	log.Printf("--> Configured Ollama Adapter URL: %s", ollamaAdapterURL)
	// log.Printf("--> Configured Azure AI Adapter URL: %s", azureAdapterURL) // Uncomment when added
	// log.Printf("--> Configured Stable Diffusion Adapter URL: %s", sdAdapterURL) // Uncomment when added

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start AI Coordinator Service: %v", err)
	}
}

// healthCheckHandler provides a basic health endpoint
func healthCheckHandler(c *gin.Context) {
	// TODO: Add checks for connectivity to configured adapters if desired
	c.JSON(http.StatusOK, gin.H{"status": "OK", "service": "AI Coordinator Service"})
}
