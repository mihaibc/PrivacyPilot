package handlers

import (
	"log"
	"net/http"

	"privacypilot-api-gateway/internal/clients"

	"github.com/gin-gonic/gin"
)

// ModerateGatewayRequest represents the expected input to the API Gateway's endpoint
type ModerateGatewayRequest struct {
	Text     string `json:"text"`     // Making Text optional at gateway level for flexibility
	ImageURL string `json:"imageUrl"` // Keep camelCase consistent with Node service input if needed
}

// ModerateHandler holds dependencies
type ModerateHandler struct {
	Moderator *clients.ModerationClient
}

// NewModerateHandler creates a new handler instance
func NewModerateHandler(moderationClient *clients.ModerationClient) *ModerateHandler {
	return &ModerateHandler{
		Moderator: moderationClient,
	}
}

// HandleModerate is the Gin handler function
func (h *ModerateHandler) HandleModerate(c *gin.Context) {
	var req ModerateGatewayRequest

	// Use Bind instead of ShouldBindJSON if you want to handle empty body gracefully
	// Or check if text and imageUrl are both empty after binding
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("API Gateway: Error binding JSON for /moderate: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Basic validation: Ensure at least one field is present
	if req.Text == "" && req.ImageURL == "" {
		log.Printf("API Gateway: Moderation request missing text and imageUrl")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: text or imageUrl must be provided"})
		return
	}

	// Call the moderation service via the client
	moderationResp, err := h.Moderator.ModerateContent(req.Text, req.ImageURL)
	if err != nil {
		log.Printf("API Gateway: Error calling moderation service: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to process request with moderation service"})
		return
	}

	// Return the response from the moderation service
	c.JSON(http.StatusOK, moderationResp)
}
