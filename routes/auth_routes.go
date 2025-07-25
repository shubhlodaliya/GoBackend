package routes

import (
    "gobackend/controllers"
    "github.com/gorilla/mux"
)

func RegisterAuthRoutes(r *mux.Router) {
    r.HandleFunc("/api/auth", controllers.AuthHandler).Methods("POST")
}
