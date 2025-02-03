package auth

import (
	"example.com/collaborative-coding-editor/middleware"
	"github.com/gorilla/mux"
)

func RegisterAuthRoutes(router *mux.Router) {

	//Public endpoints
	router.HandleFunc("/auth/register", Register).Methods("POST")
	router.HandleFunc("/auth/login", Login).Methods("POST")
	router.HandleFunc("/auth/refresh", Refresh).Methods("POST")

	//Protected Routes
	protected := router.PathPrefix("/auth").Subrouter()
	protected.Use(middleware.JWTAuthentication)
	protected.HandleFunc("/profile", Profile).Methods("GET")
	protected.HandleFunc("/invitations", GetActiveInvitations).Methods("GET")
}
