package models

import (
	"context"
	"fmt"
	"time"

	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// this is an instance of an online class, one document per section of a class.
type OnlineRoom struct {
	ID             primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name           string               `json:"room_name" bson:"room_name" validate:"required,min=1,max=32"`
	ClassLive      bool                 `json:"class_live" bson:"class_live"`
	VirtualClass   primitive.ObjectID   `json:"virtualclass,omitempty" bson:"virtualclass,omitempty"`
	CurrentTeacher []primitive.ObjectID `json:"teachers,omitempty" bson:"teachers,omitempty"`
	StudentsJoined []primitive.ObjectID `json:"students,omitempty" bson:"students,omitempty"`
	CreatedAt      time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at" bson:"updated_at"`
}

func GetRoom(key, value string) (OnlineRoom, error) {
	var v interface{}
	var room OnlineRoom

	rC, err := mongodb.GetCollectionHandle("room")
	if err != nil {
		return OnlineRoom{}, fmt.Errorf("connection to db failed")
	}

	if key == "_id" {
		v, _ = mongodb.GetMongoIDFromHex(value)
	} else {
		v = value
	}

	err = rC.FindOne(context.Background(),
		bson.D{{Key: key, Value: v}}).Decode(&room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return OnlineRoom{}, fmt.Errorf("online room not found")
		}
	}
	return room, nil
}

func CreateRoom(name string) (OnlineRoom, error) {

	room := OnlineRoom{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rC, err := mongodb.GetCollectionHandle("room")
	if err != nil {
		return OnlineRoom{}, fmt.Errorf("connection to db failed")
	}
	id, err := rC.InsertOne(context.Background(), room)
	if err != nil {
		return OnlineRoom{}, fmt.Errorf("unable to create room : %v", err)
	}
	return GetRoom("_id", id.InsertedID.(primitive.ObjectID).Hex())
}
