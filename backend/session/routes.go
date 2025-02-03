package session

import (
	"example.com/collaborative-coding-editor/middleware"
	"github.com/gorilla/mux"
)

// RegisterSessionRoutes adds session-related endpoints to the main router.
func RegisterSessionRoutes(router *mux.Router) {
	// Create a subrouter under /session
	sessionRouter := router.PathPrefix("/session").Subrouter()
	sessionRouter.Use(middleware.JWTAuthentication)

	sessionRouter.HandleFunc("/save", SaveSession).Methods("POST")
	sessionRouter.HandleFunc("/{room_id}", GetSession).Methods("GET")
	sessionRouter.HandleFunc("/export/{room_id}", ExportSession).Methods("GET")
	sessionRouter.HandleFunc("/audit", LogAudit).Methods("POST")
	sessionRouter.HandleFunc("/audit/{room_id}", GetAuditLogs).Methods("GET")
}
