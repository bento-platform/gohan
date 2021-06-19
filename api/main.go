package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	port := os.Getenv("GOHAN_API_PORT")
	if port == "" {
		fmt.Println("Port unset ; using default")
		port = "5000"
	}

	// TODO: Set up Elasticsearch connection
	// es, _ := elasticsearch.NewClient(
	// 	elasticsearch.SetSniff(false),
	// 	elasticsearch.SetURL("http://elasticsearch.gohan.local"),
	// 	elasticsearch.SetBasicAuth("username", "password"),
	// )
	// log.Println(elasticsearch.Version)
	// log.Println(es.Info())

	// TODO: Set up DRS connection

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, "Welcome to the next generation Gohan v2 API using Golang!")
	})

	e.Logger.Fatal(e.Start(":" + port))
}
