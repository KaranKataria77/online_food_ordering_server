package model

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Cart struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId       primitive.ObjectID `json:"userId" validate:"required" bson:"userId,omitempty"`
	FoodItems    []string           `json:"foodItems" validate:"required" bson:"foodItems,omitempty"`
	TotalValue   int64              `json:"totalValue" validate:"required" bson:"totalValue,omitempty"`
	IsCartActive bool               `json:"isCartActive" bson:"isCartActive,omitempty"`
}

func (cart *Cart) Validate() error {
	validate := validator.New()
	return validate.Struct(cart)
}
