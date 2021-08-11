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

type Class struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"class_name" bson:"class_name" validate:"required,min=3,max=32"`
	Sections []SectionWithName  `json:"sections,omitempty" bson:"sections,omitempty"`
	// Sections  []SectionWithName `json:"sections" bson:"sections"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type SectionWithName struct {
	SectionID   primitive.ObjectID `json:"section_id,omitempty" bson:"section_id,omitempty"`
	SectionName string             `json:"section_name,omitempty" bson:"section_name,omitempty"`
	// SectionID   primitive.ObjectID `json:"section_id" bson:"section_id"`
	// SectionName string             `json:"section_name" bson:"section_name"`
}

func GetClass(key, value string) (Class, error) {
	fmt.Println("GetClass", key, value)
	var class Class
	var v interface{}

	cC, err := mongodb.GetCollectionHandle("class")
	if err != nil {
		return Class{}, fmt.Errorf("connection to db failed")
	}

	if key == "_id" {
		v, _ = mongodb.GetMongoIDFromHex(value)
	} else {
		v = value
	}

	fmt.Println("GetClass 2", key, v)

	err = cC.FindOne(context.Background(),
		bson.D{
			{Key: key, Value: v},
		},
	).Decode(&class)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Class{}, fmt.Errorf("class not found")
		}
	}
	return class, nil
}

func GetClasses() ([]Class, error) {
	var classes []Class

	cC, err := mongodb.GetCollectionHandle("class")
	if err != nil {
		return []Class{}, fmt.Errorf("connection to db failed")
	}

	cursor, err := cC.Find(context.Background(), bson.D{{}})
	if err != nil {
		fmt.Println(err)
		//do we ever get this error for Find?, this was from FindOne
		if err == mongo.ErrNoDocuments {
			return []Class{}, fmt.Errorf("no class found")
		}
	}

	if cursor.RemainingBatchLength() == 0 {
		return []Class{}, fmt.Errorf("no class found")
	}

	if err = cursor.All(context.Background(), &classes); err != nil {
		return []Class{}, fmt.Errorf("internal server error : %v", err)
	}

	fmt.Println("Classes", classes)
	return classes, nil
}

func CreateClass(name string) (Class, error) {

	class := Class{}

	cC, err := mongodb.GetCollectionHandle("class")
	if err != nil {
		return Class{}, fmt.Errorf("connection to db failed")
	}

	err = cC.FindOne(context.Background(),
		bson.D{
			{Key: "class_name", Value: name},
		},
	).Decode(&class)

	if err == nil {
		return Class{}, fmt.Errorf("class already exist")
	}

	//It does not create a empty array, makes it null - none of the MONGODB 'array' calls works on it.
	// t1 := make([]SectionWithName, 0)
	// t2 := SectionWithName{SectionID: primitive.ObjectID{}}
	// t1 = append(t1, t2)
	// class.Sections = t1

	class = Class{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// Sections:  t1,		//omitempty with the bson tag works fine.
	}

	id, err := cC.InsertOne(context.Background(), class)
	if err != nil {
		return Class{}, fmt.Errorf("unable to create class : %v", err)
	}

	// id.InsertedID.(primitive.ObjectID).String()) ---> ObjectID("60f2da6f9e5991acff7d8fa4")

	c, err := GetClass("_id", id.InsertedID.(primitive.ObjectID).Hex())
	if err != nil {
		return Class{}, fmt.Errorf("class create, unabale to retrieve at this time %v", err)
	}

	return c, nil
}

func AddSectionToClass(cid primitive.ObjectID, section SectionWithName) {

	fmt.Println("AddSectionToClass", cid, section.SectionID, section.SectionName)
	//how will you check if the section exists - this should have been done while adding section

	cC, err := mongodb.GetCollectionHandle("class")
	if err != nil {
		log.Fatal("connection to db failed", err)
	}

	update := bson.M{
		"$addToSet": bson.M{
			"sections": section,
		},
	}

	// update := bson.M{
	// 	"$push": bson.M{
	// 		"sections": bson.M{
	// 			"$each": section,
	// 			"$position":0,
	// 		}
	// 	},
	// }

	// ({_id:xx}, {$push:{letters : {$each:['first one'], $position:0}}})

	result, err := cC.UpdateOne(
		context.Background(),
		bson.M{"_id": cid},
		update,
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)

}
