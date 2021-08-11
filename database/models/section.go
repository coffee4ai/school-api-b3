package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Section struct {
	ID         primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string               `json:"section_name" bson:"section_name" validate:"required,min=1,max=32"`
	Students   []primitive.ObjectID `json:"students,omitempty" bson:"students,omitempty"`
	Teachers   []primitive.ObjectID `json:"teachers,omitempty" bson:"teachers,omitempty"`
	OnlineRoom primitive.ObjectID   `json:"online_room,omitempty" bson:"online_room,omitempty"`
	CreatedAt  time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at" bson:"updated_at"`
}

func GetSection(key, value string) (Section, error) {
	var v interface{}
	var section Section

	sC, err := mongodb.GetCollectionHandle("section")
	if err != nil {
		return Section{}, fmt.Errorf("connection to db failed")
	}

	if key == "_id" {
		v, _ = mongodb.GetMongoIDFromHex(value)
	} else {
		v = value
	}

	err = sC.FindOne(context.Background(),
		bson.D{{Key: key, Value: v}}).Decode(&section)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Section{}, fmt.Errorf("section not found")
		}
	}
	return section, nil
}

func CreateSection(name string) (Section, error) {

	section := Section{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	sC, err := mongodb.GetCollectionHandle("section")
	if err != nil {
		return Section{}, fmt.Errorf("connection to db failed")
	}
	//I need to check if there is a section already in this
	id, err := sC.InsertOne(context.Background(), section)
	if err != nil {
		return Section{}, fmt.Errorf("unable to create section : %v", err)
	}
	return GetSection("_id", id.InsertedID.(primitive.ObjectID).Hex())
}

func AddUserToSection(user User) {

	field := ""

	sC, err := mongodb.GetCollectionHandle("section")
	if err != nil {
		log.Fatal(err)
	}

	switch user.Roles[0] {
	case "student":
		field = "students"
	case "teacher":
		field = "teachers"
	}

	update := bson.M{
		"$addToSet": bson.M{
			field: user.ID,
		},
	}

	result, err := sC.UpdateOne(
		context.Background(),
		bson.M{"_id": user.BelongsTo.SectionID},
		update,
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
}
