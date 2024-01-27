package controllers

import (
	"fmt"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func (server *Server) InitRoutes() {
	server.Router = mux.NewRouter()
	server.Router.Use(handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	))
	server.Router.HandleFunc("/api/user", server.CreateUser).Methods("POST", "OPTIONS")
	server.Router.HandleFunc("/api/user/{id}", server.GetUser).Methods("GET")
	server.Router.HandleFunc("/api/user/{id}", server.UpdateUser).Methods("PUT")
	server.Router.HandleFunc("/api/order", server.CreateOrder).Methods("POST")
	server.Router.HandleFunc("/api/order/{id}", server.UpdateOrder).Methods("PATCH")
	server.Router.HandleFunc("/api/orders", server.GetAllOrders).Methods("GET")
	fmt.Println("Routes initilized")
}
