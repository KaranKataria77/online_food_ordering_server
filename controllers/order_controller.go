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

var ErrOrderNotFound = errors.New("Order not found")

type OrderUpdate struct {
	FoodItems  []string `json:"foodItems"`
	TotalValue string   `json:"totalValue"`
}

func (server *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// server.enableCORS(&w)
	fmt.Println("Create order route called")
	collection = server.database.Collection("orders")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Allow-Control-Allow-Methods", "POST")
	w.Header().Set("Content-Type", "application/x-www-form-urlencode")

	var order model.Order
	_ = json.NewDecoder(r.Body).Decode(&order)
	err := order.Validate()
	if err != nil {
		errorResponse := map[string]string{"error": err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
	} else {
		insertOrder(order)
		json.NewEncoder(w).Encode(order)
	}
}

func (server *Server) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// server.enableCORS(&w)
	w.Header().Set("Allow-Control-Allow-Methods", "PATCH")
	collection = server.database.Collection("orders")
	vars := mux.Vars(r)
	id := vars["id"]
	var order OrderUpdate
	_ = json.NewDecoder(r.Body).Decode(&order)
	err := updateOrderByID(id, &order)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "Order not found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Something went wrong"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(order)
}

func (server *Server) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	// server.enableCORS(&w)
	w.Header().Set("Allow-Control-Allow-Methods", "GET")
	collection = server.database.Collection("orders")
	var orders []model.Order
	err := fetchAllOrders(&orders)
	if err != nil {
		if err == ErrUserNotFound {
			errorResponse := map[string]string{"error": "No Orders found"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		} else {
			errorResponse := map[string]string{"error": "Internal Server error"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
		}
		return
	}
	json.NewEncoder(w).Encode(orders)
}

func fetchAllOrders(orders *[]model.Order) error {
	filter := bson.D{{}}

	results, err := collection.Find(context.Background(), filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrOrderNotFound
		}
		return err
	}
	fmt.Println("Order are ", results)
	for results.Next(context.Background()) {
		var order model.Order
		err := results.Decode(&order)
		if err != nil {
			return err
		}
		*orders = append(*orders, order)
	}
	return nil
}

func updateOrderByID(orderId string, order *OrderUpdate) error {
	id, _ := primitive.ObjectIDFromHex(orderId)
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

func insertOrder(order model.Order) {
	fmt.Println("Order collection created")
	inserted, err := collection.InsertOne(context.Background(), order)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("One movie inserted ID ", inserted.InsertedID)
}
