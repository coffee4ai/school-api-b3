package handlers

import (
	"fmt"
	"net/http"

	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/middleware"

	"github.com/labstack/echo/v4"
)

func registerB3Routes(base *echo.Group) {
	base.GET("/start", StartClass)
	base.GET("/join", JoinClass)
}

func StartClass(c echo.Context) error {

	token := c.Request().Header.Get("Authorization")
	admin, err := middleware.VerifyTokenRole(token, "admin")
	teacher, err := middleware.VerifyTokenRole(token, "teacher")

	if !admin && !teacher {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	class_id := c.QueryParam("class_id")
	section_id := c.QueryParam("section_id")

	if class_id == "" || section_id == "" {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{"provide a class id and section id for starting an online class"})
	}

	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"Invalid Params, please specify a class ID"})
}

func JoinClass(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	admin, err := middleware.VerifyTokenRole(token, "admin")
	student, err := middleware.VerifyTokenRole(token, "student")

	if !admin && !student {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{err.Error()})
	}
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
