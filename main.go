package main

import (
	"log"

	"github.com/coffee4ai/school-api-b3/config"
	"github.com/coffee4ai/school-api-b3/database/models"
	"github.com/coffee4ai/school-api-b3/database/mongodb"
	"github.com/coffee4ai/school-api-b3/handlers"
	"github.com/coffee4ai/school-api-b3/server"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	if _, err := mongodb.ConnectDB(); err != nil {
		log.Fatal("Error conneting mongodb", err)
	}

	models.CheckAdminAccount()
	s := server.New()
	base := s.Group("/api/v1")
	handlers.RegisterRoutes(base)
	server.Start(s)
}
