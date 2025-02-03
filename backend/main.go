package main

import (
	"fmt"
	"log"
	"net/http"

	"example.com/collaborative-coding-editor/auth"
	"example.com/collaborative-coding-editor/collaboration"
	"example.com/collaborative-coding-editor/compiler"
	"example.com/collaborative-coding-editor/config"
	"example.com/collaborative-coding-editor/rooms"
	"example.com/collaborative-coding-editor/session"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	// create a new router
	router := mux.NewRouter()

	// Health-check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Backend is running")
	}).Methods("GET")

	// Load config
	config.LoadConfig()

	// Connect DB
	auth.Connect()

	// Setup router
	auth.RegisterAuthRoutes(router)
	rooms.RegisterRoomRoutes(router)
	compiler.RegisterCompilerRoutes(router)
	session.RegisterSessionRoutes(router)

	// Websocket router
	router.HandleFunc("/collaboration/{room_id}", collaboration.WebSocketHandler)

	// CORS Settings
	allowedOrigins, allowedMethods, allowedHeaders := defineCorsSettings(router)
	cors := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router)

	// start the server
	startMuxServer(cors)
}

func defineCorsSettings(router *mux.Router) (handlers.CORSOption, handlers.CORSOption, handlers.CORSOption) {
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	return allowedOrigins, allowedMethods, allowedHeaders
}

func startMuxServer(handler http.Handler) {
	log.Println("Backend server starting at 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
