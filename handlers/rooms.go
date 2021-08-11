package handlers

import (
	"fmt"
	"net/http"

	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"github.com/coffee4ai/school-api-b3/middleware"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoomRoutes(g *echo.Group) {
	g.GET("/room/:id", GetRoom)
	g.POST("/room", CreateRoom)
}

func GetRoom(c echo.Context) error {
	//room/<id> or room/all?class_id=<class-id>
	id := c.Param("id")
	if id == "all" {
		cid := c.QueryParam("class_id")
		if cid == "" {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{"provide a class ID for getting all rooms"})
		}

		class, err := models.GetClass("_id", cid)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		allRooms := make([]models.OnlineRoom, 0)
		for _, v := range class.Sections {
			s, err := models.GetSection("_id", v.SectionID.Hex())
			if err != nil {
				fmt.Printf("Skipping %v section with error %v", v.SectionID.String(), err)
				continue
			}
			room, err := models.GetRoom("_id", s.OnlineRoom.Hex())
			if err != nil {
				fmt.Printf("Skipping %v section with error %v", room.ID.String(), err)
				continue
			}
			allRooms = append(allRooms, room)
		}
		return c.JSON(http.StatusOK, allRooms)
	} else if primitive.IsValidObjectID(id) {
		room, err := models.GetRoom("_id", id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		return c.JSON(http.StatusOK, room)
	}
	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"invalid params, please specify a room id"})
}

func CreateRoom(c echo.Context) error {
	//room?section_id=<section-id>
	var room models.OnlineRoom
	if auth, err := middleware.VerifyTokenRole(c.Request().Header.Get("Authorization"), "admin"); !auth {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	section_id := c.QueryParam("section_id")
	if !mongodb.IsValidMongoID(section_id) {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{"provide a section id for creating a room"})
	}

	c.Echo().Validator = middleware.Val
	if err := c.Bind(&room); err != nil {
		fmt.Println("Error while binding", err)
		return c.JSON(http.StatusInternalServerError, struct {
			M string `json:"message"`
		}{"Internal Server Error"})
	}

	if err := c.Validate(room); err != nil {
		m := middleware.GetErrorString(err)
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{m})
	}

	s, err := models.GetSection("_id", section_id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{err.Error()})
	}
	if _, err := models.GetRoom("_id", s.OnlineRoom.Hex()); err == nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{"room already exist for given section"})
	}

	res, err := models.CreateRoom(room.Name)
	s.OnlineRoom = res.ID
	return c.JSON(http.StatusOK, res)
}
