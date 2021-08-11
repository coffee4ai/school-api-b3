package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"github.com/coffee4ai/school-api-b3/middleware"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name" validate:"required,min=3,max=32"`
	UserID     string             `json:"user_id" bson:"user_id" validate:"required,min=3,max=32"`
	Password   string             `json:"password,omitempty" bson:"password,omitempty"`
	Roles      []string           `json:"roles" bson:"roles" validate:"validRoles"`
	Experience string             `json:"experience,omitempty" bson:"experience,omitempty"`
	Expertise  string             `json:"expertise,omitempty" bson:"expertise,omitempty"`
	BelongsTo  ClassSection       `json:"belongs_to" bson:"belongs_to" validate:"required"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

type ClassSection struct {
	ClassName       string             `json:"class_name" bson:"class_name"`
	SectionName       string             `json:"section_name" bson:"section_name"`
	ClassID   primitive.ObjectID `json:"class_id" bson:"class_id" validate:"required"`
	SectionID primitive.ObjectID `json:"section_id" bson:"section_id" validate:"required"`
}

type UserFilters struct {
	Role string
}

func CheckAdminAccount() {
	uC, err := mongodb.GetCollectionHandle("users")
	if err != nil {
		log.Println("connection to db failed - possible user collection does not exists", err)
	}
	// checkHandleErr(err)
	cursor, err := uC.Find(context.Background(), bson.D{{Key: "roles", Value: "admin"}})
	// checkDBErr(err)
	if err != nil {
		log.Println("Something went wrong while reading DB", err)
		return
	}
	if cursor.RemainingBatchLength() == 0 {
		log.Println("No admin account found, creating a admin account")
		hPass, err := middleware.GenerateHash("Admin@123")
		if err != nil {
			log.Println("Error generating password hash", err)
			return
		}
		admin := User{
			Name:      "Admin",
			UserID:    "admin",
			Password:  string(hPass),
			Roles:     []string{"admin"},
			BelongsTo: ClassSection{
				ClassName: "",
				SectionName: "",
				ClassID: primitive.NilObjectID, 
				SectionID: primitive.NilObjectID},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err = uC.InsertOne(context.Background(), admin)
		// checkDBErr(err)
		if err != nil {
			log.Fatal("Unable to create admin account - ABORT", err)
		}
	}
}

// GetUser should except any of the three params - MongoID, User_ID or Name and return the user

func GetUser(key, value string) (User, error) {
	var v interface{}
	var user User
	uC, err := mongodb.GetCollectionHandle("users")
	if err != nil {
		return User{}, fmt.Errorf("connection to db failed")
	}
	// checkHandleErr(err)
	// projection := bson.D{
	// 	{Key: "password", Value: 0},
	// }
	// opts := options.FindOne().SetProjection(projection)
	projection := bson.D{}
	//this is bad - it should be in some #define / iota values
	if key == "_id" {
		v, _ = mongodb.GetMongoIDFromHex(value)
	} else {
		v = value
	}

	err = uC.FindOne(context.Background(),
		bson.D{
			{Key: key, Value: v},
		},
		options.FindOne().SetProjection(projection),
	).Decode(&user)

	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return User{}, fmt.Errorf("user not found")
		}
	}
	fmt.Printf("found document %v", user)
	return user, nil
}

func GetUsers(filters UserFilters) ([]User, error) {
	var user []User
	uC, err := mongodb.GetCollectionHandle("users")
	if err != nil {
		return []User{}, fmt.Errorf("connection to db failed")
	}
	// checkHandleErr(err)

	f := bson.D{{}}
	if filters.Role != "" {
		f = bson.D{{Key: "roles", Value: filters.Role}}
		// f = bson.D{{{"roles", filters.role}}, bson.D{{"roles", bson.D{{"$nin", []string{"admin"}}}}}}
		// bson.D{{"status", bson.D{{"$in", bson.A{"A", "D"}}}}})
		// bson.D{{"roles", bson.D{{"$nin", []string{"admin"}}}}}
	}

	projection := bson.D{
		{Key: "password", Value: 0},
	}
	// opts := options.FindOne().SetProjection(projection)
	// dont send the admin user
	// err := c.Find(bson.M{"friends": bson.M{"$in": arr}}).All(&users)
	// cursor, err := userCollections.Find(context.Background(), f, options.Find().SetProjection(projection))

	//$eq	Matches values that are equal to a specified value.
	//$gt	Matches values that are greater than a specified value.
	//$gte	Matches values that are greater than or equal to a specified value.
	//$in	Matches any of the values specified in an array.
	//$lt	Matches values that are less than a specified value.
	//$lte	Matches values that are less than or equal to a specified value.
	//$ne	Matches all values that are not equal to a specified value.
	//$nin	Matches none of the values specified in an array.

	// #all documents whose country field value is not neither Portugal or Spain
	// query = db.collection.find({
	//     "country" : { '$nin': [
	//         'Portugal',
	//         'Spain']
	//     }
	// })
	// cursor, err := userCollections.Find(context.Background(), bson.D{{"roles", bson.D{{"$nin", []string{"admin"}}}}}, options.Find().SetProjection(projection))

	//Until I fix the admin removal from the Find, no support for roles based filter for now
	cursor, err := uC.Find(context.Background(), f, options.Find().SetProjection(projection))
	// fmt.Printf("%v", cursor)
	// fmt.Println("Cursor result", cursor.RemainingBatchLength(), err)
	if err != nil {
		fmt.Println(err)
		// ErrNoDocuments means that the filter did not match any documents in the collection
		//Check what errors can come from Find, the below one is sprcific for FindOne
		if err == mongo.ErrNoDocuments {
			return []User{}, fmt.Errorf("user not found")
			// return []User{},
			// 	echo.NewHTTPError(http.StatusBadRequest, errorMessage{Message: "User not found"})
		}
	}
	if cursor.RemainingBatchLength() == 0 {
		return []User{}, fmt.Errorf("user not found")
	}
	if err = cursor.All(context.Background(), &user); err != nil {
		message := fmt.Sprintf("Internal server error : %v", err)
		fmt.Println(message)
		return []User{}, fmt.Errorf("internal server error")
	}
	return user, nil
}

func CreateUser(user User) (User, error) {
	uC, err := mongodb.GetCollectionHandle("users")
	if err != nil {
		return User{}, fmt.Errorf("connection to db failed")
	}

	userExists := User{}
	err = uC.FindOne(context.Background(),
		bson.D{
			{Key: "user_id", Value: user.UserID},
		},
	).Decode(&userExists)

	if err == nil {
		return User{}, fmt.Errorf("user already exists")
	}

	res, err := uC.InsertOne(context.Background(), user)
	if err != nil {
		return User{}, fmt.Errorf("unable to create user : %v", err)
	}
	return GetUser("_id", res.InsertedID.(primitive.ObjectID).Hex())
}
