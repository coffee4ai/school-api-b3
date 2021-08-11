package handlers

import (
	"fmt"
	"net/http"

	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/middleware"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func registerClassRoutes(base *echo.Group) {
	base.GET("/class/:id", GetClass)
	base.DELETE("/class/:id", DeleteClass)
	base.PATCH("/class/:id", UpdateClass)
	base.POST("/class", CreateClass)
}

func GetClass(c echo.Context) error {
	id := c.Param("id")
	if id == "all" {

		if auth, err := middleware.VerifyTokenRole(c.Request().Header.Get("Authorization"), "admin"); !auth {
			return c.JSON(http.StatusUnauthorized, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		fmt.Println("Getclass 2")
		classes, err := models.GetClasses()
		if err != nil {
			fmt.Println("Getclass 3")
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		fmt.Println("Getclass 4")
		return c.JSON(http.StatusOK, classes)
	} else if primitive.IsValidObjectID(id) {
		class, err := models.GetClass("_id", id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		return c.JSON(http.StatusOK, class)
	}
	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"Invalid Params, please specify a class ID"})
}

func CreateClass(c echo.Context) error {
	fmt.Println("CreateClass")
	class := models.Class{}

	if auth, err := middleware.VerifyTokenRole(c.Request().Header.Get("Authorization"), "admin"); !auth {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	if err := c.Bind(&class); err != nil {
		fmt.Println("Error while binding", err, err.Error())
		return c.JSON(http.StatusInternalServerError, struct {
			M string `json:"message"`
		}{"Internal Server Error"})
	}

	c.Echo().Validator = middleware.Val
	if err := c.Validate(class); err != nil {
		m := middleware.GetErrorString(err)
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{m})
	}

	newClass, err := models.CreateClass(class.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{err.Error()})
	}
	return c.JSON(http.StatusOK, newClass)
}

func DeleteClass(c echo.Context) error {

	// Deleting a class has huge consequences, all the sections under it needs to be deleted first, before which
	// all students/teachers linked to it needs to be unlinked.

	//verify its a admin

	//validate the class id

	//get the sections and unlink all the students and teachers

	//delete all the sections

	//delete the classs

	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"invalid params, please specify a class id"})
}

func UpdateClass(c echo.Context) error {
	// class := models.Class{}

	//verify the role

	//validate the class id

	//bind the class to get the new information to be updated from JSON body & validate it

	//update the class

	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"invalid params, please specify a class id"})
}
