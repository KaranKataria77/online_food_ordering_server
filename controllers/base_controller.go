package controllers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Router   *mux.Router
	database *mongo.Database
}

var connectionString string

const dbName = "anime"

var collection *mongo.Collection

func (server *Server) Init() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Error while loading env")
	}
	username := os.Getenv("MONGO_USERNAME")
	password := os.Getenv("MONGO_PASSWORD")
	connectionString = "mongodb+srv://" + username + ":" + password + "@cluster0.tydt9cl.mongodb.net/?retryWrites=true&w=majority"
	// client options
	clientOptions := options.Client().ApplyURI(connectionString)

	// connect to DB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB Connection successed")

	server.database = client.Database(dbName)
	server.initUserCollection()
}
