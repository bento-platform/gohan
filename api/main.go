package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch"
	es7 "github.com/elastic/go-elasticsearch"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func createEsConnection(e *echo.Echo) *es7.Client {

	var (
		clusterURLs = []string{"http://localhost:9200"}
		username    = "elastic"
		password    = "changeme!"
	)
	cfg := elasticsearch.Config{
		Addresses: clusterURLs,
		Username:  username,
		Password:  password,
	}
	es7Client, _ := es7.NewClient(cfg)

	e.Logger.Debug(es7.Version)

	return es7Client
}

func main() {
	port := os.Getenv("GOHAN_API_PORT")
	if port == "" {
		fmt.Println("Port unset ; using default")
		port = "5000"
	}

	e := echo.New()
	e.Use(middleware.Recover())

	// Service Connections:
	// -- Elasticsearch
	es := createEsConnection(e)

	// TODO: Set up DRS connection

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, "Welcome to the next generation Gohan v2 API using Golang!")
	})

	e.GET("/searchtest", func(c echo.Context) error {
		// Testing ES
		es.Search()
		// {{proto}}://{{host}}:{{port}}/variants/_search

		// Build the request body.
		var queryString = `
			"bool": {
			  "filter": [
				{
				  "bool": {
					"must": [
					  {
						"query_string": {
						  "query": "chrom:*"
						}
					  }
					]
				  }
				}
			  ]
			}
		`
		query := constructQuery(queryString, 2)
		var buf bytes.Buffer
		if buffErr := json.NewEncoder(&buf).Encode(query); buffErr != nil {
			e.Logger.Fatalf("Error encoding query: %s", buffErr)
		}

		// Perform the search request.
		res, searchErr := es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex("variants"),
			es.Search.WithBody(&buf),
			es.Search.WithTrackTotalHits(true),
			es.Search.WithPretty(),
		)
		if searchErr != nil {
			fmt.Println("Error getting response: %s", searchErr)
		}

		respBuf := new(strings.Builder)
		_, respErr := io.Copy(respBuf, res.Body)
		if respErr != nil {
			fmt.Println("Error forming response: %s", respErr)
		}
		// check errors
		fmt.Println(respBuf.String())

		defer res.Body.Close()

		fmt.Println(res)
		return c.JSON(http.StatusOK, respBuf)
	})

	e.Logger.Fatal(e.Start(":" + port))
}

func constructQuery(q string, size int) *strings.Reader {

	// Build a query string from string passed to function
	var query = `{"query": {`

	// Concatenate query string with string passed to method call
	query = query + q

	// Use the strconv.Itoa() method to convert int to string
	query = query + `}, "size": ` + strconv.Itoa(size) + `}`
	fmt.Println("\nquery:", query)

	// Check for JSON errors
	isValid := json.Valid([]byte(query)) // returns bool

	// Default query is "{}" if JSON is invalid
	if isValid == false {
		fmt.Println("constructQuery() ERROR: query string not valid:", query)
		fmt.Println("Using default match_all query")
		query = "{}"
	} else {
		fmt.Println("constructQuery() valid JSON:", isValid)
	}

	// Build a new string from JSON query
	var b strings.Builder
	b.WriteString(query)

	// Instantiate a *strings.Reader object from string
	read := strings.NewReader(b.String())

	// Return a *strings.Reader object
	return read
}
