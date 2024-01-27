package model

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId           primitive.ObjectID `json:"userId" validate:"required"" bson:"users,omitempty"`
	FoodItems        []string           `json:"foodItems" validate:"required"`
	TotalValue       string             `json:"totalValue" validate:"required"`
	Payment          primitive.ObjectID `json:"payment" bson:"payment"`
	IsOrderDelivered bool               `json:"isOrderDelivered"`
	IsOrderCancelled bool               `json:"isOrderCancelled"`
}

func (order *Order) Validate() error {
	validate := validator.New()
	return validate.Struct(order)
}
