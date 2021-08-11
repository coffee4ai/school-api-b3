package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/coffee4ai/school-api-b3/config"
	custom "github.com/coffee4ai/school-api-b3/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	*echo.Echo
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, struct {
		M string `json:"status"`
	}{"OK"})
}

func New() Server {
	server := echo.New()
	server.HideBanner = true
	server.Use(middleware.CORS())
	server.Use(custom.Logger)
	server.GET("/", healthCheck)
	return Server{server}
}

func Start(s Server) {

	go func() {
		if err := s.Start(fmt.Sprintf(":%s", config.GetApiPort())); err != nil && err != http.ErrServerClosed {
			log.Fatal("shutting down the server", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
