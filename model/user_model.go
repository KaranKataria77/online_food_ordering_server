package model

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string               `json:"name" validate:"required"`
	Email     string               `json:"email" validate:"required,email"`
	Mobile_No string               `json:"mobile_no"`
	Address   string               `json:"address"`
	Location  [2]float64           `json:"location"`
	Orders    []primitive.ObjectID `json:"orders" bson:"orders"`
}

func (user *User) Validate() error {
	validate := validator.New()
	return validate.Struct(user)
}
