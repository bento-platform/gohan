package main

import (
	"api/contexts"
	"api/mvc"
	"api/utils"

	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Gather environment variables
	port := os.Getenv("GOHAN_API_PORT")
	if port == "" {
		fmt.Println("Port unset - using default")
		port = "5000"
	}

	// Instantiate Server
	e := echo.New()

	// Service Connections:
	// -- Elasticsearch
	es := utils.CreateEsConnection()

	// TODO:
	// -- Set up DRS connection

	// Configure Server
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	// -- Override handlers with custom "ES" context
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &contexts.EsContext{c, es}
			return h(cc)
		}
	})

	// Begin MVC Routes
	// -- Root
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Welcome to the next generation Gohan v2 API using Golang!")
	})

	// -- Variants
	e.GET("/variants/searchtest", mvc.VariantsSearchTest)
	// -- TESTING
	e.GET("/variants/ingest", mvc.VariantsIngestTest)

	// Run
	e.Logger.Fatal(e.Start(":" + port))
}
