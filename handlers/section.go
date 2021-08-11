package handlers

import (
	"fmt"
	"net/http"

	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/middleware"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func registerSectionRoutes(base *echo.Group) {
	base.GET("/section/:id", GetSection)
	base.POST("/section", CreateSection)

}

func GetSection(c echo.Context) error {
	//section/<id> or section/all?class_id=<class-id> one
	id := c.Param("id")
	if id == "all" {
		cid := c.QueryParam("class_id")
		if cid == "" {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{"provide a class ID for getting all sections"})
		}

		class, err := models.GetClass("_id", cid)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		allSections := make([]models.Section, 0)
		for _, v := range class.Sections {
			s, err := models.GetSection("_id", v.SectionID.Hex())
			if err != nil {
				fmt.Printf("Skipping %v section with error %v", v.SectionID.String(), err)
				continue
			}
			allSections = append(allSections, s)
		}
		return c.JSON(http.StatusOK, allSections)
	} else if primitive.IsValidObjectID(id) {
		sec, err := models.GetSection("_id", id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		return c.JSON(http.StatusOK, sec)
	}
	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"Invalid Params, please specify a class ID"})
}

func CreateSection(c echo.Context) error {
	fmt.Println("CreateSection")

	//section?class_id=<class-id>
	var section models.Section
	if auth, err := middleware.VerifyTokenRole(c.Request().Header.Get("Authorization"), "admin"); !auth {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	class_id := c.QueryParam("class_id")
	if class_id == "" {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{"provide a class ID for creating a section"})
	}

	c.Echo().Validator = middleware.Val
	if err := c.Bind(&section); err != nil {
		fmt.Println("Error while binding", err)
		return c.JSON(http.StatusInternalServerError, struct {
			M string `json:"message"`
		}{"Internal Server Error"})
	}
	fmt.Println("This is what I for single section", section)
	if err := c.Validate(section); err != nil {
		m := middleware.GetErrorString(err)
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{m})
	}

	class, err := models.GetClass("_id", class_id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	for _, v := range class.Sections {
		if v.SectionName == section.Name {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{"section already present in given class"})
		}
	}

	newSection, err := models.CreateSection(section.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	secName := models.SectionWithName{
		SectionID:   newSection.ID,
		SectionName: newSection.Name,
	}
	models.AddSectionToClass(class.ID, secName)

	return c.JSON(http.StatusOK, newSection)
}
