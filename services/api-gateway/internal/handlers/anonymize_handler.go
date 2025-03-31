package handlers

import (
	"log"
	"net/http"

	"privacypilot-api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

// AnonymizeRequest represents the expected input to the API Gateway's endpoint
type AnonymizeGatewayRequest struct {
	Text string `json:"text" binding:"required"`
}

// AnonymizeHandler holds dependencies for the handler, like the client
type AnonymizeHandler struct {
	Anonymizer *clients.AnonymizerClient
}

// NewAnonymizeHandler creates a new handler instance
func NewAnonymizeHandler(anonymizerClient *clients.AnonymizerClient) *AnonymizeHandler {
	return &AnonymizeHandler{
		Anonymizer: anonymizerClient,
	}
}

// HandleAnonymize is the Gin handler function
func (h *AnonymizeHandler) HandleAnonymize(c *gin.Context) {
	var req AnonymizeGatewayRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("API Gateway: Error binding JSON for /anonymize: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Call the anonymizer service via the client
	anonymizeResp, err := h.Anonymizer.AnonymizeText(req.Text)
	if err != nil {
		log.Printf("API Gateway: Error calling anonymizer service: %v", err)
		// Determine appropriate status code based on error type if possible
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to process request with anonymizer service"})
		return
	}

	// Return the response from the anonymizer service directly
	c.JSON(http.StatusOK, anonymizeResp)
}
