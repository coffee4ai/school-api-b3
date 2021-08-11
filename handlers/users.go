package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"github.com/coffee4ai/school-api-b3/middleware"
	"github.com/labstack/echo/v4"
)

func registerUserRoutes(base *echo.Group) {
	base.GET("/user/:id", GetUser)
	base.POST("/user", CreateUser)
}

func GetUser(c echo.Context) error {
	// - /user/:id?conditions - applicable for /user/<mongodb-id> or /user/<all>
	// conditions - role=teacher

	id := c.Param("id")
	if id == "all" {
		if auth, err := middleware.VerifyTokenRole(c.Request().Header.Get("Authorization"), "admin"); !auth {
			return c.JSON(http.StatusUnauthorized, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		//Parse all the filters and populate the UserFilter struct
		var filters models.UserFilters
		role := c.QueryParam("role")
		roles := make([]string, 0)
		roles = append(roles, role)
		if role != "" && middleware.IsValidRole(roles) {
			//validate role to be either student or teacher
			filters.Role = role
		}
		users, err := models.GetUsers(filters)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		return c.JSON(http.StatusOK, users)
	} else if mongodb.IsValidMongoID(id) {
		user, err := models.GetUser("_id", id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		user.Password = ""
		return c.JSON(http.StatusOK, user)
	}
	return c.JSON(http.StatusBadRequest, struct {
		M string `json:"message"`
	}{"Invalid Params, please specify a user ID"})

}

func CreateUser(c echo.Context) error {

	func() {
		fmt.Println("Checking the DB for class_name Class Two")
		fmt.Println(models.GetClass("class_name", "Class Two"))
	}()
	//there is no param, this is for a single user & data is in JSON body
	if auth, err := middleware.VerifyTokenRole(c.Request().Header.Get("Authorization"), "admin"); !auth {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{err.Error()})
	}
	if len(c.QueryParams()) == 0 {
		u := models.User{}
		// section := models.Section{}

		c.Echo().Validator = middleware.Val
		if err := c.Bind(&u); err != nil {
			fmt.Println("Error while binding", err)
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{"invalid json data"})
		}
		fmt.Println("This is what I for single user", u)
		if err := c.Validate(u); err != nil {
			m := middleware.GetErrorString(err)
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{m})
		}

		class, err := models.GetClass("_id", u.BelongsTo.ClassID.Hex())
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		fmt.Println("Got class name is ", class.Name)

		section, err := models.GetSection("_id", u.BelongsTo.SectionID.Hex())
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}
		fmt.Println("Got section name is ", section.Name)

		if u.Password == "" {
			u.Password = u.UserID + "@123"
		}
		p, _ := middleware.GenerateHash(u.Password)
		u.Password = string(p)
		if len(u.Roles) == 0 {
			u.Roles = make([]string, 0)
			u.Roles = append(u.Roles, "student")
		}
		u.CreatedAt, u.UpdatedAt = time.Now(), time.Now()
		u.BelongsTo.ClassID = class.ID
		u.BelongsTo.ClassName = class.Name
		u.BelongsTo.SectionID = section.ID
		u.BelongsTo.SectionName = section.Name
		user, err := models.CreateUser(u)
		if err != nil {
			return c.JSON(http.StatusBadRequest, struct {
				M string `json:"message"`
			}{err.Error()})
		}

		models.AddUserToSection(user)

		//will the user have more than one role?

		user.Password = ""
		return c.JSON(http.StatusOK, user)
	}
	// we got a excel filename as which has the data
	// api/v1/user?file_name=hello.xlsx&role=student
	fileName := c.QueryParam("file_name")
	fmt.Println("Rx file : ", fileName)
	fileType := strings.Split(fileName, ".")
	fmt.Println(fileType)
	if len(fileType) < 2 || fileType[1] != "xlsx" {
		fmt.Println("invalid file foramt, upload an excel file with .xlsx")
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{"invalid file foramt, upload an excel file with .xlsx"})
	}
	roleInfo := c.QueryParam("role")
	r := make([]string, 0)
	r = append(r, roleInfo)
	if !middleware.IsValidRole(r) || roleInfo == "admin" {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{"invalid role in url"})
	}
	listStudents, err := loadUserData(fileName, roleInfo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	for _, v := range listStudents {
		//at this point all errors are ruled out
		if _, err := models.CreateUser(v); err != nil {
			fmt.Println("This is never supposed to happen", err)
		}
		models.AddUserToSection(v)
	}
	mess := fmt.Sprintf("successfully added %d %s", len(listStudents), roleInfo)

	return c.JSON(http.StatusOK, struct {
		M string `json:"message"`
	}{mess})

}

//move this to a different package which handles all excel related data & return a slice of users or a json object
//tried moving it to middleware, which results into a import cycle
func loadUserData(fileName, role string) ([]models.User, error) {
	// Name UserID ClassName SectionName
	fmt.Println("This is the file name", fileName)

	listStudents := make([]models.User, 0)
	f, err := excelize.OpenFile("/tmp/" + fileName)
	if err != nil {
		fmt.Println(err)
		return []models.User{}, fmt.Errorf("given file not uploaded")
	}
	sheetOne := f.GetSheetName(1)
	fmt.Println("This is the sheet name - ", sheetOne)
	rows, err := f.Rows(sheetOne)
	if err != nil {
		fmt.Println(err)
		return []models.User{}, fmt.Errorf("invalid excel sheet data")
	}

	for i := 0; rows.Next(); i++ {
		if i == 0 {
			//not a good way to skip the header row
			continue
		}
		colCell := rows.Columns()
		if err != nil {
			fmt.Println(err)
		}
		var s models.User

		//We need to make sure each row has 4 columns before we start reading them -
		//this way it will make sure there is no panic for out of range

		s.Name = colCell[0]
		s.UserID = colCell[1]

		cls, err := models.GetClass("class_name", colCell[2])
		if err != nil {
			fmt.Println(err)
			return []models.User{}, fmt.Errorf("error parsing excel - class name error at row %d", i)
		}
		s.BelongsTo.ClassID = cls.ID
		s.BelongsTo.ClassName = cls.Name
		section, err := models.GetSection("section_name", colCell[3])
		if err != nil {
			fmt.Println(err)
			return []models.User{}, fmt.Errorf("error parsing excel - section name error at row %d", i)
		}
		s.BelongsTo.SectionID = section.ID
		s.BelongsTo.SectionName = section.Name
		s.Roles = []string{role}
		if err = middleware.Val.Validate(s); err != nil {
			// middleware.Val.val.Struct(s); err != nil {
			return []models.User{}, fmt.Errorf(middleware.GetErrorString(err))
		}
		p, _ := middleware.GenerateHash(s.UserID + "@123")
		s.Password = string(p)
		listStudents = append(listStudents, s)
	}
	return listStudents, nil
}
