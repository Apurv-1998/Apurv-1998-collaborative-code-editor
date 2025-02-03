package rooms

import (
	"example.com/collaborative-coding-editor/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoomRoutes(router *mux.Router) {
	// Create a subrouter under the /rooms path
	roomRouter := router.PathPrefix("/rooms").Subrouter()
	// Apply JWT middleware to all room routes
	roomRouter.Use(middleware.JWTAuthentication)

	roomRouter.HandleFunc("", CreateRoom).Methods("POST")
	roomRouter.HandleFunc("/{room_id}/invite", GenerateInvite).Methods("POST")
	roomRouter.HandleFunc("/join", JoinRoom).Methods("POST")
	roomRouter.HandleFunc("/history", GetRoomHistory).Methods("GET")
	roomRouter.HandleFunc("/{room_id}", GetRoomDetails).Methods("GET")
	roomRouter.HandleFunc("/{room_id}/close", CloseRoom).Methods("POST")
}
