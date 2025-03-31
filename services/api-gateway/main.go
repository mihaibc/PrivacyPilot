package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"privacypilot-api-gateway/internal/clients"
	"privacypilot-api-gateway/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	router := gin.Default()

	// --- Service Clients ---
	anonymizerURL := strings.TrimRight(os.Getenv("ANONYMIZER_SERVICE_URL"), "/")
	if anonymizerURL == "" {
		log.Fatal("ANONYMIZER_SERVICE_URL environment variable not set")
	}
	anonymizerClient := clients.NewAnonymizerClient(anonymizerURL)

	moderationURL := strings.TrimRight(os.Getenv("MODERATION_SERVICE_URL"), "/")
	if moderationURL == "" {
		log.Fatal("MODERATION_SERVICE_URL environment variable not set")
	}
	moderationClient := clients.NewModerationClient(moderationURL) // Instantiate moderation client

	// aiCoordinatorURL := strings.TrimRight(os.Getenv("AI_COORDINATOR_URL"), "/")
	// if aiCoordinatorURL == "" {
	//  log.Fatal("AI_COORDINATOR_URL environment variable not set")
	// }
	// aiCoordinatorClient := clients.NewAICoordinatorClient(aiCoordinatorURL) // Create later

	// --- Handlers ---
	anonymizerClient := clients.NewAnonymizerClient(anonymizerURL)
	moderationClient := clients.NewModerationClient(moderationURL)
	anonymizeHandler := handlers.NewAnonymizeHandler(anonymizerClient)
	moderateHandler := handlers.NewModerateHandler(moderationClient)
	// aiHandler := handlers.NewAIHandler(aiCoordinatorClient) // Create later

	// --- Routes ---
	router.GET("/health", healthCheckHandler)

	// API v1 Routes
	apiV1 := router.Group("/api/v1")
	{
		// Add authentication middleware here later
		// apiV1.Use(authMiddleware())

		apiV1.POST("/anonymize", anonymizeHandler.HandleAnonymize)
		apiV1.POST("/moderate", moderateHandler.HandleModerate) // Register moderate route
	}

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("API Gateway starting on port %s\n", port)
	log.Printf("--> Anonymizer Service URL: %s", anonymizerURL)
	log.Printf("--> Moderation Service URL: %s", moderationURL) // Log moderation URL
	// log.Printf("--> AI Coordinator URL: %s", aiCoordinatorURL)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// healthCheckHandler remains the same
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK", "service": "API Gateway"})
}

// Placeholder for auth middleware
// ... (keep as before) ...
