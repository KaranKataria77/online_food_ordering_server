package main

import (
	"fmt"
	"log"
	"net/http"
	"online_food_ordering/controllers"
)

func main() {
	// newUser := model.User{
	// 	Name: "karan",
	// }

	server := controllers.Server{}
	server.Init()
	server.InitRoutes()
	log.Fatal(http.ListenAndServe("127.0.0.1:4000", server.Router))
	fmt.Println("PORT running at 4000")
}
