package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"privacypilot-api-gateway/internal/clients"
	"privacypilot-api-gateway/internal/handlers"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mock Moderation Service Setup ---

// Mock Moderation Service Response (Example: OK)
var mockModerationResponseOK = clients.ModerationResponse{
	IsAcceptable:    true,
	Flags:           []string{},
	Details:         "Content looks fine.",
	ConfidenceScore: 0.98,
}

// Mock Moderation Service Response (Example: Flagged)
var mockModerationResponseFlagged = clients.ModerationResponse{
	IsAcceptable:    false,
	Flags:           []string{"hate_speech"},
	Details:         "Detected potential hate speech.",
	ConfidenceScore: 0.90,
}

// Mock Moderation Service Server
func setupMockModerationServer(t *testing.T, expectedReqBody clients.ModerationRequest, responseToReturn clients.ModerationResponse, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Mock Moderation: Expected POST request")
		assert.Equal(t, "/moderate", r.URL.Path, "Mock Moderation: Expected path /moderate")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Mock Moderation: Expected Content-Type header")

		var reqBody clients.ModerationRequest // Use client's request struct
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err, "Mock Moderation: Failed to decode request body")
		// Compare expected vs actual request body fields
		assert.Equal(t, expectedReqBody.Text, reqBody.Text, "Mock Moderation: Unexpected text in request body")
		assert.Equal(t, expectedReqBody.ImageURL, reqBody.ImageURL, "Mock Moderation: Unexpected imageUrl in request body")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode) // Use provided status code
		err = json.NewEncoder(w).Encode(responseToReturn)
		assert.NoError(t, err, "Mock Moderation: Failed to encode response")
	}))
}

// Updated Router Setup to include Moderation Client/Handler
func setupGatewayRouter(anonymizerURL, moderationURL string) *gin.Engine { // Add moderationURL param
	gin.SetMode(gin.TestMode)
	router := gin.New()

	anonymizerClient := clients.NewAnonymizerClient(anonymizerURL)
	moderationClient := clients.NewModerationClient(moderationURL) // Create moderation client

	anonymizeHandler := handlers.NewAnonymizeHandler(anonymizerClient)
	moderateHandler := handlers.NewModerateHandler(moderationClient) // Create moderation handler

	router.GET("/health", healthCheckHandler)
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/anonymize", anonymizeHandler.HandleAnonymize)
		apiV1.POST("/moderate", moderateHandler.HandleModerate) // Register moderation handler
	}

	return router
}

// --- Moderation Endpoint Tests ---

func TestModerateRoute_Success_OK(t *testing.T) {
	// Input data for the gateway request
	gatewayRequestBody := handlers.ModerateGatewayRequest{Text: "This is safe text"}
	requestBodyBytes, _ := json.Marshal(gatewayRequestBody)

	// Expected body to be received by the mock moderation service
	expectedModerationReq := clients.ModerationRequest{Text: "This is safe text"}

	// 1. Setup Mock Moderation Service (to return OK response)
	mockServer := setupMockModerationServer(t, expectedModerationReq, mockModerationResponseOK, http.StatusOK)
	defer mockServer.Close()

	// 2. Setup Gateway Router (Anonymizer URL not needed for this test, pass empty string or mock)
	router := setupGatewayRouter("", mockServer.URL) // Pass mock moderation server URL

	// 3. Perform Request to Gateway
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/moderate", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Assertions
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK from gateway")
	var gatewayResponseBody clients.ModerationResponse
	err := json.Unmarshal(rr.Body.Bytes(), &gatewayResponseBody)
	assert.NoError(t, err)
	assert.Equal(t, mockModerationResponseOK, gatewayResponseBody, "Gateway response should match mock OK response")
}

func TestModerateRoute_Success_Flagged(t *testing.T) {
	// Input data
	gatewayRequestBody := handlers.ModerateGatewayRequest{Text: "Potentially problematic text"}
	requestBodyBytes, _ := json.Marshal(gatewayRequestBody)
	expectedModerationReq := clients.ModerationRequest{Text: "Potentially problematic text"}

	// 1. Setup Mock Moderation Service (to return Flagged response)
	mockServer := setupMockModerationServer(t, expectedModerationReq, mockModerationResponseFlagged, http.StatusOK) // Still 200 OK, but flagged content
	defer mockServer.Close()

	// 2. Setup Gateway Router
	router := setupGatewayRouter("", mockServer.URL)

	// 3. Perform Request
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/moderate", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Assertions
	assert.Equal(t, http.StatusOK, rr.Code)
	var gatewayResponseBody clients.ModerationResponse
	err := json.Unmarshal(rr.Body.Bytes(), &gatewayResponseBody)
	assert.NoError(t, err)
	assert.Equal(t, mockModerationResponseFlagged, gatewayResponseBody, "Gateway response should match mock Flagged response")
}

func TestModerateRoute_ModerationServiceError(t *testing.T) {
	// Input data
	gatewayRequestBody := handlers.ModerateGatewayRequest{Text: "Some text"}
	requestBodyBytes, _ := json.Marshal(gatewayRequestBody)
	expectedModerationReq := clients.ModerationRequest{Text: "Some text"}

	// 1. Setup Mock Moderation Service that returns an error status
	mockServer := setupMockModerationServer(t, expectedModerationReq, clients.ModerationResponse{}, http.StatusInternalServerError) // Simulate 500
	defer mockServer.Close()

	// 2. Setup Gateway Router
	router := setupGatewayRouter("", mockServer.URL)

	// 3. Perform Request
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/moderate", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Assertions
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code, "Expected Service Unavailable from gateway")
	var errorResponse map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "Failed to process request", "Expected error message")
}

func TestModerateRoute_GatewayBadRequest_MissingFields(t *testing.T) {
	// 1. Setup mock server (should not be called)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Mock moderation server should not be called")
	}))
	defer mockServer.Close()

	// 2. Setup Gateway Router
	router := setupGatewayRouter("", mockServer.URL)

	// 3. Perform Request with empty body
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/moderate", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// 4. Assertions
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected Bad Request from gateway for missing fields")
	var errorResponse map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse["error"], "text or imageUrl must be provided")
}

// Keep existing tests for /health and /api/v1/anonymize
// func TestHealthCheckRoute(t *testing.T) { ... }
// func TestAnonymizeRoute_Success(t *testing.T) { ... }
// func TestAnonymizeRoute_AnonymizerServiceError(t *testing.T) { ... }
// func TestAnonymizeRoute_BadRequest(t *testing.T) { ... }

// Remember helper functions like setupMockAnonymizerServer if not shown fully above.
