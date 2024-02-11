package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"online_food_ordering/consts"
	"online_food_ordering/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrOrderNotFound = errors.New("Order not found")

func (server *Server) CreateOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create order route called")

	session, _ := server.client.StartSession()
	defer session.EndSession(context.Background())
	collection = server.database.Collection("carts")

	err := mongo.WithSession(context.Background(), session, func(sessionContext mongo.SessionContext) error {
		err := session.StartTransaction()
		if err != nil {
			consts.SendErrorResponse(&w, http.StatusInternalServerError, consts.InternalServerError, err)
			session.AbortTransaction(sessionContext)
			return err
		}
		var order model.Order
		_ = json.NewDecoder(r.Body).Decode(&order)
		filter := bson.D{{"_id", order.CartId}}
		update := bson.D{{"$set", bson.D{{"isCartActive", false}}}}
		result, cartUpdateErr := collection.UpdateOne(sessionContext, filter, update)
		if cartUpdateErr != nil {
			consts.SendErrorResponse(&w, http.StatusBadRequest, consts.ErrorUpdatingCart, err)
			session.AbortTransaction(sessionContext)
			return err
		}
		fmt.Println("Updated cart ", result)
		validationErr := order.Validate()
		if validationErr != nil {
			consts.SendErrorResponse(&w, http.StatusBadRequest, consts.ErrorRequiredFieldMissing, err)
			session.AbortTransaction(sessionContext)
			return err
		} else {
			collection = server.database.Collection("orders")
			inserted, err := collection.InsertOne(sessionContext, order)
			if err != nil {
				log.Fatal(err)
			}
			commitErr := session.CommitTransaction(sessionContext)
			if commitErr != nil {
				log.Panic("Error while commiting transaction")
			}
			fmt.Println("One order inserted ID ", inserted.InsertedID)
			json.NewEncoder(w).Encode(order)
		}
		return nil
	})

	if err != nil {
		log.Panic("Error while creating order")
	}
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
