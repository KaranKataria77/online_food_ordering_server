package model

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CartId           primitive.ObjectID `json:"cartId" validate:"required"" bson:"cartId,omitempty"`
	Payment          primitive.ObjectID `json:"payment" bson:"payment"`
	IsOrderDelivered bool               `json:"isOrderDelivered" bson:"isOrderDelivered"`
	IsOrderCancelled bool               `json:"isOrderCancelled" bson:"isOrderCancelled"`
}

func (order *Order) Validate() error {
	validate := validator.New()
	return validate.Struct(order)
}
