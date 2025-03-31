package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"privacypilot-anonymizer-service/internal/clients"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mock AI Coordinator Setup ---

// Mock AI Coordinator Response (Successful Anonymization)
var mockCoordinatorResponseAnonymizeOK = clients.AICoordinatorResponse{
	Success: true,
	Result: map[string]interface{}{ // Simulating map[string]interface{} before specific unmarshal
		"anonymized_text": "Contact me at [REDACTED_EMAIL] or call [REDACTED_NUMBER].",
	},
}

// Mock AI Coordinator Server
func setupMockAICoordinatorServer(t *testing.T, expectedTaskType string, expectedPayload map[string]string, responseToReturn clients.AICoordinatorResponse, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/process", r.URL.Path) // Assuming /process endpoint
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody clients.AICoordinatorRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		// Validate Task Type
		assert.Equal(t, expectedTaskType, reqBody.TaskType)

		// Validate Payload (assuming map[string]string for text anonymization payload)
		payloadMap, ok := reqBody.Payload.(map[string]interface{})
		assert.True(t, ok, "Payload should be a map")
		for key, val := range expectedPayload {
			assert.Equal(t, val, payloadMap[key], "Payload field mismatch")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		err = json.NewEncoder(w).Encode(responseToReturn)
		assert.NoError(t, err)
	}))
}

// Setup Router with Mocked AI Coordinator Client
func setupAnonymizerRouterWithMocks(coordURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create client pointing to mock server
	aiCoordClient = clients.NewAICoordinatorClient(coordURL) // Assign to global for handler to use

	// Setup router (as before)
	router := gin.New()
	router.GET("/health", healthCheckHandler)
	router.POST("/anonymize", anonymizeHandler)
	return router
}

// --- Tests ---

// TestAnonymizerHealthCheckRoute remains the same as before...
func TestAnonymizerHealthCheckRoute(t *testing.T) {
	// No external dependencies needed, can use simplified setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", healthCheckHandler)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var responseBody map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "OK", responseBody["status"])
	assert.Equal(t, "Anonymizer Service", responseBody["service"])
}

func TestAnonymizeHandler_Success(t *testing.T) {
	// Input to the anonymizer service endpoint
	inputText := "Contact me at test@example.com or call 123456789."
	requestBody := AnonymizeRequest{Text: inputText}
	requestBodyBytes, _ := json.Marshal(requestBody)

	// Expected payload for the AI Coordinator request
	expectedCoordPayload := map[string]string{"text": inputText}
	expectedCoordTaskType := clients.TaskTypeAnonymizeText

	// 1. Setup Mock AI Coordinator Server (to return OK response)
	mockServer := setupMockAICoordinatorServer(t, expectedCoordTaskType, expectedCoordPayload, mockCoordinatorResponseAnonymizeOK, http.StatusOK)
	defer mockServer.Close()

	// 2. Setup Anonymizer Router with Mock Client
	router := setupAnonymizerRouterWithMocks(mockServer.URL)

	// 3. Perform Request to Anonymizer Service
	req, _ := http.NewRequest(http.MethodPost, "/anonymize", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Assertions
	assert.Equal(t, http.StatusOK, rr.Code)

	var responseBody AnonymizeResponse
	err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, inputText, responseBody.OriginalText)

	// Extract expected anonymized text from the mock coordinator response result
	expectedAnonymizedText := mockCoordinatorResponseAnonymizeOK.Result.(map[string]interface{})["anonymized_text"]
	assert.Equal(t, expectedAnonymizedText, responseBody.AnonymizedText)
}

func TestAnonymizeHandler_CoordinatorError(t *testing.T) {
	// Input
	inputText := "Some text"
	requestBody := AnonymizeRequest{Text: inputText}
	requestBodyBytes, _ := json.Marshal(requestBody)
	expectedCoordPayload := map[string]string{"text": inputText}
	expectedCoordTaskType := clients.TaskTypeAnonymizeText

	// 1. Setup Mock AI Coordinator to return an error (e.g., internal server error)
	mockErrorResponse := clients.AICoordinatorResponse{Success: false, Error: "AI Model Failed"}
	mockServer := setupMockAICoordinatorServer(t, expectedCoordTaskType, expectedCoordPayload, mockErrorResponse, http.StatusInternalServerError) // Or OK status with Success=false
	defer mockServer.Close()

	// 2. Setup Router with Mock
	router := setupAnonymizerRouterWithMocks(mockServer.URL)

	// 3. Perform Request
	req, _ := http.NewRequest(http.MethodPost, "/anonymize", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Assertions
	// Anonymizer should return an internal server error if coordinator fails
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var errorResponse map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Failed to process anonymization request via AI Coordinator")
}

// TestAnonymizeHandler_BadRequest remains the same as before, testing input validation
func TestAnonymizeHandler_BadRequest(t *testing.T) {
	// 1. Setup mock server (should not be called)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("AI Coordinator should not be called on bad anonymizer request")
	}))
	defer mockServer.Close()

	// 2. Setup Router
	router := setupAnonymizerRouterWithMocks(mockServer.URL) // Pass mock URL even if not called

	// 3. Perform Request with malformed JSON
	malformedJSON := `{"text": "some text",}`
	req, _ := http.NewRequest(http.MethodPost, "/anonymize", bytes.NewBufferString(malformedJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// 4. Perform Request with missing field
	missingFieldJSON := `{}`
	reqMissing, _ := http.NewRequest(http.MethodPost, "/anonymize", bytes.NewBufferString(missingFieldJSON))
	reqMissing.Header.Set("Content-Type", "application/json")
	rrMissing := httptest.NewRecorder()
	router.ServeHTTP(rrMissing, reqMissing)
	assert.Equal(t, http.StatusBadRequest, rrMissing.Code)
}
