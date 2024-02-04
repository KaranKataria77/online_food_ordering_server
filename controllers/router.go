package controllers

import (
	"fmt"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func (server *Server) InitRoutes() {
	server.Router = mux.NewRouter()
	server.Router.Use(handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowCredentials(),
	))
	// user
	server.Router.HandleFunc("/api/user", server.CreateUser).Methods("POST", "OPTIONS")
	server.Router.HandleFunc("/api/user", server.GetUser).Methods("GET", "OPTIONS")
	server.Router.HandleFunc("/api/user/{id}", server.UpdateUser).Methods("PUT")
	server.Router.HandleFunc("/api/user-login/", server.UserLogin).Methods("POST", "OPTIONS")
	// cart
	server.Router.HandleFunc("/api/cart/", server.GetCart).Methods("GET")
	server.Router.HandleFunc("/api/cart/", server.CreateCart).Methods("POST", "OPTIONS")
	server.Router.HandleFunc("/api/cart/", server.UpdateCart).Methods("PATCH", "OPTIONS")
	server.Router.HandleFunc("/api/cart-deactivate/", server.DeactivateCart).Methods("GET", "OPTIONS")
	// orders
	server.Router.HandleFunc("/api/order", server.CreateOrder).Methods("POST", "OPTIONS")
	// server.Router.HandleFunc("/api/order/{id}", server.UpdateOrder).Methods("PATCH")
	server.Router.HandleFunc("/api/orders", server.GetAllOrders).Methods("GET")
	fmt.Println("Routes initilized")
}
