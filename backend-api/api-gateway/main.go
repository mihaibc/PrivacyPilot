package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

// OAuth middleware: checks for an Authorization header and validates the token.
func oauthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" || !validateToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Simulated OAuth token validation. In production, integrate with an OAuth provider.
func validateToken(token string) bool {
	// Expect token in format "Bearer valid-token"
	return token == "Bearer valid-token"
}

// proxyRequest forwards incoming requests to the specified targetURL.
func proxyRequest(targetURL string, w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	mux := http.NewServeMux()

	// Endpoint: /anonymize routes to the anonymizer-service (assumed to run on port 8081)
	mux.HandleFunc("/anonymize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		proxyRequest("http://localhost:8081/anonymize", w, r)
	})

	// Endpoint: /moderate routes to the moderation-service (assumed to run on port 8082)
	mux.HandleFunc("/moderate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		proxyRequest("http://localhost:8082/moderate", w, r)
	})

	// Apply OAuth middleware to all routes.
	handler := oauthMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway running on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
