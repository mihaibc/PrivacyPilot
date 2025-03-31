package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"privacypilot-ai-coordinator/internal/clients"

	"github.com/gin-gonic/gin"
)

// Structs and Constants remain the same...
type AICoordinatorRequest struct {
	TaskType string                 `json:"task_type" binding:"required"`
	Payload  map[string]interface{} `json:"payload" binding:"required"`
	Config   map[string]string      `json:"config,omitempty"` // Config map for hints like model
}
type AICoordinatorResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

const (
	TaskTypeAnonymizeText = "anonymize_text"
	TaskTypeModerateText  = "moderate_text"
	TaskTypeModerateImage = "moderate_image"
)

// ProcessHandler remains the same structure
type ProcessHandler struct {
	OllamaClient *clients.OllamaAdapterClient
	// AzureClient  *clients.AzureAdapterClient
}

// NewProcessHandler remains the same
func NewProcessHandler(ollamaClient *clients.OllamaAdapterClient /* other clients */) *ProcessHandler {
	if ollamaClient == nil {
		log.Println("Warning: OllamaAdapterClient is nil during ProcessHandler creation.")
	}
	return &ProcessHandler{
		OllamaClient: ollamaClient,
	}
}

func (h *ProcessHandler) HandleProcessRequest(c *gin.Context) {
	var req AICoordinatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("AI Coordinator: Error binding JSON for /process: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body: " + err.Error()})
		return
	}

	log.Printf("AI Coordinator: Received task '%s'", req.TaskType)

	var result interface{}
	var err error

	// --- Routing Logic ---
	switch strings.ToLower(req.TaskType) {
	case TaskTypeAnonymizeText:
		log.Printf("AI Coordinator: Routing '%s' task to Ollama Adapter.", req.TaskType)
		if h.OllamaClient == nil {
			err = fmt.Errorf("ollama adapter client is not configured")
		} else {
			// Extract model hint from config, if present
			modelHint := ""
			if req.Config != nil {
				modelHint = req.Config["model"] // Look for a "model" key in the config map
			}
			if modelHint != "" {
				log.Printf("AI Coordinator: Using model hint from request config: '%s'", modelHint)
			}

			// Call the Ollama adapter client, passing the hint
			var adapterResp *clients.OllamaAdapterAnonymizeResponse
			adapterResp, err = h.OllamaClient.AnonymizeText(req.Payload, modelHint) // Pass modelHint

			if adapterResp != nil {
				// Store the structured result including the model used
				result = map[string]string{
					"anonymized_text": adapterResp.AnonymizedText,
					"model_used":      adapterResp.ModelUsed,
				}
			}
		}

	// ... cases for TaskTypeModerateText, TaskTypeModerateImage remain the same ...
	case TaskTypeModerateText:
		log.Printf("AI Coordinator: Routing '%s' task (Not Implemented Yet)", req.TaskType)
		err = fmt.Errorf("adapter for task '%s' not implemented yet", req.TaskType)
	case TaskTypeModerateImage:
		log.Printf("AI Coordinator: Routing '%s' task (Not Implemented Yet)", req.TaskType)
		err = fmt.Errorf("adapter for task '%s' not implemented yet", req.TaskType)

	default:
		// ... default case remains the same ...
		log.Printf("AI Coordinator: Unsupported task type '%s'", req.TaskType)
		c.JSON(http.StatusBadRequest, AICoordinatorResponse{Success: false, Error: fmt.Sprintf("Unsupported task type: %s", req.TaskType)})
		return
	}
	// --- End Routing ---

	if err != nil {
		// ... error handling remains the same ...
		log.Printf("AI Coordinator: Error processing task '%s': %v", req.TaskType, err)
		c.JSON(http.StatusInternalServerError, AICoordinatorResponse{Success: false, Error: fmt.Sprintf("Failed to process task '%s': %v", req.TaskType, err)})
		return
	}

	log.Printf("AI Coordinator: Successfully processed task '%s'", req.TaskType)
	c.JSON(http.StatusOK, AICoordinatorResponse{
		Success: true,
		Result:  result,
	})
}
