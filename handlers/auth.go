package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/coffee4ai/school-api-b3/config"
	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"github.com/coffee4ai/school-api-b3/middleware"
	"github.com/cristalhq/jwt/v3"
	"github.com/labstack/echo/v4"
)

func registerAuthRoutes(base *echo.Group) {
	base.POST("/signin", SignIn)
	// g.POST("/signup", SignUp)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type LoginCreds struct {
	UserID   string `json:"user_id" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=5,max=32"`
}

func SignIn(c echo.Context) error {
	loginInfo := LoginCreds{}
	if err := c.Bind(&loginInfo); err != nil {
		// Bind error are not internal server errors - this can happend when we get invalid json data
		fmt.Println("Error while binding", err, err.Error())
		return c.JSON(http.StatusInternalServerError, struct {
			M string `json:"message"`
		}{"Internal Server Error"})
	}

	fmt.Println("Rx login Info", loginInfo)
	c.Echo().Validator = middleware.Val
	if err := c.Validate(loginInfo); err != nil {
		m := middleware.GetErrorString(err)
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{m})
	}

	user, err := models.GetUser("user_id", loginInfo.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, struct {
			M string `json:"message"`
		}{err.Error()})
	}

	if !middleware.ComparePassword(loginInfo.Password, []byte(user.Password)) {
		return c.JSON(http.StatusUnauthorized, struct {
			M string `json:"message"`
		}{"invalid password"})
	}
	sKey := config.GetApiSecret()
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(sKey))
	checkErr(err)

	claims := &middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  user.Roles,
			ID:        mongodb.GetStringFromMongoID(user.ID),
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(time.Hour * 24 * 7)},
		},
		User_ID: user.UserID,
	}

	// create a Builder
	builder := jwt.NewBuilder(signer)

	// and build a Token
	newToken, err := builder.Build(claims)
	checkErr(err)

	token := newToken.String()
	fmt.Println(token)
	//dont send the password
	user.Password = ""
	// user.Password = "" //do not send the password, this should be empty
	return c.JSON(http.StatusOK, struct {
		Message string      `json:"message"`
		U       models.User `json:"user"`
		Token   string      `json:"token"`
	}{"Success", user, token})
}
