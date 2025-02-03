package compiler

import (
	"example.com/collaborative-coding-editor/middleware"
	"github.com/gorilla/mux"
)

func RegisterCompilerRoutes(router *mux.Router) {
	// Create a subrouter under /compile and apply middleware
	compilerRouter := router.PathPrefix("/compile").Subrouter()
	compilerRouter.Use(middleware.JWTAuthentication)
	compilerRouter.HandleFunc("", CompileCode).Methods("POST")
}
