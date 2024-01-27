package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"online_food_ordering/model"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrUserNotFound = errors.New("user not found")

func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-www-form-urlencode")
	fmt.Println("Create user route called")
	collection = server.database.Collection("users")

	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	err := user.Validate()
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
	} else {
		insertOneUser(user)
		json.NewEncoder(w).Encode(user)
	}
}

func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	collection = server.database.Collection("users")
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Println("Id is ", id)
	var user model.User
	err := getUserByID(id, &user)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "User not found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Internal Server error"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (server *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	collection = server.database.Collection("users")
	vars := mux.Vars(r)
	id := vars["id"]
	var user model.User
	_ = json.NewDecoder(r.Body).Decode(&user)
	err := updateUserByID(id, &user)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "User not found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Something went wrong"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func updateUserByID(userId string, user *model.User) error {
	id, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{"name", user.Name}, {"email", user.Email}, {"mobile_no", user.Mobile_No}}}}

	result, err := collection.UpdateOne(context.Background(), filter, update)
	fmt.Println("Result after modified ", result.ModifiedCount)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}
func getUserByID(userId string, user *model.User) error {
	id, _ := primitive.ObjectIDFromHex(userId)
	filter := bson.D{{"_id", id}}

	err := collection.FindOne(context.Background(), filter).Decode(user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
func insertOneUser(user model.User) {
	fmt.Println("User collection created")
	inserted, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("One movie inserted ID ", inserted.InsertedID)
}
