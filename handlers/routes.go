package handlers

import "github.com/labstack/echo/v4"

func RegisterRoutes(base *echo.Group) {
	registerUserRoutes(base)
	registerClassRoutes(base)
	registerSectionRoutes(base)
	registerAuthRoutes(base)
	registerPublicRoutes(base)
	registerB3Routes(base)
}
