package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"privacypilot-anonymizer-service/internal/clients"

	"github.com/gin-gonic/gin"
)

// Request/Response structs remain the same for this service's external API
type AnonymizeRequest struct {
	Text string `json:"text" binding:"required"`
}

type AnonymizeResponse struct {
	OriginalText   string `json:"original_text"`
	AnonymizedText string `json:"anonymized_text"`
}

// Global variable for the AI Coordinator client (or use dependency injection)
var aiCoordClient *clients.AICoordinatorClient

func main() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	// --- Instantiate AI Coordinator Client ---
	aiCoordinatorURL := strings.TrimRight(os.Getenv("AI_COORDINATOR_URL"), "/")
	if aiCoordinatorURL == "" {
		log.Fatal("AI_COORDINATOR_URL environment variable not set for Anonymizer Service")
	}
	aiCoordClient = clients.NewAICoordinatorClient(aiCoordinatorURL) // Assign to global variable
	//-----------------------------------------

	router := gin.Default()

	// --- Routes ---
	router.GET("/health", healthCheckHandler)
	router.POST("/anonymize", anonymizeHandler) // Handler now uses the client

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("Anonymizer Service starting on port %s", port)
	log.Printf("--> AI Coordinator URL: %s", aiCoordinatorURL)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start Anonymizer Service: %v", err)
	}
}

// healthCheckHandler remains the same
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK", "service": "Anonymizer Service"})
}

// anonymizeHandler processes the anonymization request using the AI Coordinator Client
func anonymizeHandler(c *gin.Context) {
	var req AnonymizeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// --- Call AI Coordinator ---
	log.Printf("Anonymizer Service: Requesting anonymization from AI Coordinator for text.")
	anonymizeResult, err := aiCoordClient.RequestAnonymization(req.Text)
	if err != nil {
		log.Printf("Anonymizer Service: Error calling AI Coordinator: %v", err)
		// Respond with a server error if the coordinator call failed
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process anonymization request via AI Coordinator"})
		return
	}
	// --------------------------

	resp := AnonymizeResponse{
		OriginalText:   req.Text,
		AnonymizedText: anonymizeResult.AnonymizedText, // Use result from coordinator
	}

	log.Printf("Anonymizer Service: Successfully processed anonymization request.")
	c.JSON(http.StatusOK, resp)
}

// Remove the old placeholder function:
// func performSimpleAnonymization(text string) string { ... }
