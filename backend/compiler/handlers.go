package compiler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"example.com/collaborative-coding-editor/config"
)

// Compile Request
type CompileRequest struct {
	Script       string `json:"script"`
	Language     string `json:"language"`
	VersionIndex string `json:"versionIndex"`
	Stdin        string `json:"stdin,omitempty"`
}

// Jdoodle Request
type JDoodleRequest struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Script       string `json:"script"`
	Language     string `json:"language"`
	VersionIndex string `json:"versionIndex"`
	Stdin        string `json:"stdin,omitempty"`
}

// Jdoodle Response
type JDoodleResponse struct {
	Output     string `json:"output"`
	StatusCode int    `json:"statusCode"`
	Memory     string `json:"memory"`
	CPUTime    string `json:"cpuTime"`
	Error      string `json:"error,omitempty"`
}

// Compile Code
func CompileCode(w http.ResponseWriter, r *http.Request) {

	// Decode the incoming JSON
	var compileRequest CompileRequest
	if err := json.NewDecoder(r.Body).Decode(&compileRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if compileRequest.Script == "" || compileRequest.Language == "" || compileRequest.VersionIndex == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Build Jdoodle request
	jdoodleRequest := JDoodleRequest{
		ClientId:     config.AppConfig.JDoodleClientID,
		ClientSecret: config.AppConfig.JDoodleClientSecret,
		Script:       compileRequest.Script,
		Language:     compileRequest.Language,
		VersionIndex: compileRequest.VersionIndex,
		Stdin:        compileRequest.Stdin,
	}

	reqBody, err := json.Marshal(jdoodleRequest)
	if err != nil {
		log.Printf("Error marshalling JDoodle request: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Create an HTTP client
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// Send thr POST request
	resp, err := httpClient.Post(config.AppConfig.JDoodleEndpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error calling Jdoodle API: %v", err)
		http.Error(w, "Error calling code execution service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read Jdoodles response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response from code execution service", http.StatusInternalServerError)
		return
	}

	var jdoodleResponse JDoodleResponse
	if err := json.Unmarshal(body, &jdoodleResponse); err != nil {
		log.Printf("Error parsing Jdoodle response: %v", err)
		http.Error(w, "Error parsing Jdoodle response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jdoodleResponse)
}
