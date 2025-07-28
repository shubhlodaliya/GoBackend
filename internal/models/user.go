package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Mobile string             `bson:"mobile" json:"mobile"`
	Token  string             `bson:"token" json:"token"`
}
