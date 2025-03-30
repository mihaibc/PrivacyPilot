package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type AnonymizeRequest struct {
	Text string `json:"text"`
}

type AnonymizeResponse struct {
	AnonymizedText string `json:"anonymized_text"`
}

func anonymizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var reqData AnonymizeRequest
	if err := json.Unmarshal(body, &reqData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	anonymized := anonymizeText(reqData.Text)
	resData := AnonymizeResponse{AnonymizedText: anonymized}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resData)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/anonymize", anonymizeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Anonymizer Service running on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
