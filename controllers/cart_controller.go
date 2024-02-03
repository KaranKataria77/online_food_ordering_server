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

type CartUpdate struct {
	FoodItems  []string `json:"foodItems"`
	TotalValue string   `json:"totalValue"`
}

var ErrCartNotFound = errors.New("Cart not found")

func (server *Server) CreateCart(w http.ResponseWriter, r *http.Request) {
	// server.enableCORS(&w)
	fmt.Println("Create cart route called")
	collection = server.database.Collection("carts")
	w.Header().Set("Content-Type", "application/x-www-form-urlencode")
	token, cokkieErr := readCookie(r)
	fmt.Println("Reading token from cookies")
	if cokkieErr != nil {
		fmt.Println("Error reading in cookies ", cokkieErr)
		errorResponse := map[string]string{"error": "Unauthorized user"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	id, tokenErr := verifyToken(token)
	if tokenErr != nil {
		fmt.Println("Invalid Token ", tokenErr)
		return
	}

	var cart model.Cart
	cart.UserId, _ = primitive.ObjectIDFromHex(id)
	_ = json.NewDecoder(r.Body).Decode(&cart)
	err := cart.Validate()
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
	} else {
		insertCart(cart)
		json.NewEncoder(w).Encode(cart)
	}
}
func (server *Server) DeactivateCart(w http.ResponseWriter, r *http.Request) {
	// server.enableCORS(&w)
	// vars := mux.Vars(r)
	collection = server.database.Collection("carts")
	id, er := getUserIdFromToken(r)
	if er != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": er,
		})
	}
	err := deactivateCartByID(id)
	if err != nil {
		log.Panic("Error while deactivating cart", err)
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Cart is deactivated",
		})
	}
}
func (server *Server) UpdateCart(w http.ResponseWriter, r *http.Request) {
	// server.enableCORS(&w)
	w.Header().Set("Allow-Control-Allow-Methods", "PATCH")
	collection = server.database.Collection("carts")
	vars := mux.Vars(r)
	id := vars["id"]
	var cart CartUpdate
	_ = json.NewDecoder(r.Body).Decode(&cart)
	err := updateCartByID(id, &cart)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "Cart not found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Something went wrong"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(cart)
}
func (server *Server) GetCart(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get cart route called")
	collection = server.database.Collection("carts")
	id, er := getUserIdFromToken(r)
	if er != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": er,
		})
	}
	fmt.Println("get cart route - Id is ", id)
	var cart model.Cart
	err := getCartByID(id, &cart)
	if err != nil {
		if err == ErrCartNotFound {
			errorResponse := map[string]string{"message": "Cart not found"}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Internal Server error"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(cart)
}

func getCartByID(userId string, cart *model.Cart) error {
	id, _ := primitive.ObjectIDFromHex(userId)
	fmt.Println("Object id from string ", id)
	// filter := bson.D{{Key: "isCartActive", Value: true}}
	filter := bson.D{{Key: "userId", Value: id}, {Key: "isCartActive", Value: true}}

	err := collection.FindOne(context.Background(), filter).Decode(cart)
	if err != nil {
		fmt.Println("Error while finding Cart ", err, id)
		if err == mongo.ErrNoDocuments {
			return ErrCartNotFound
		}
		return err
	}
	return nil
}

func updateCartByID(cartId string, order *CartUpdate) error {
	id, _ := primitive.ObjectIDFromHex(cartId)
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{"foodItems", order.FoodItems}, {"totalValue", order.TotalValue}}}}

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

func deactivateCartByID(userId string) error {

	id, er := primitive.ObjectIDFromHex(userId)
	if er != nil {
		log.Panic("Error while converting  user Id to object ID ", er)
	}
	filter := bson.D{{"userId", id}, {"isCartActive", true}}
	update := bson.D{
		{"$set", bson.D{
			{"isCartActive", false},
		}},
	}

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

func insertCart(cart model.Cart) {
	fmt.Println("Cart collection created")
	inserted, err := collection.InsertOne(context.Background(), cart)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("One movie inserted ID ", inserted.InsertedID)
}
