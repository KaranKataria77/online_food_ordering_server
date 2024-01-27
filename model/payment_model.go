package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Payment struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId     primitive.ObjectID `json:"userId,omitempty" bson:"users,omitempty"`
	TotalValue string             `json:"totalValue"`
}
