package main

import (
	"api/contexts"
	"api/mvc"
	"api/services"
	"api/utils"
	"flag"

	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Gather environment variables
	// TODO: use only flags instead?
	port := os.Getenv("GOHAN_API_PORT")
	if port == "" {
		fmt.Println("Port unset - using default")
		port = "5000"
	}

	// -- Gather flags (if any)
	var (
		vcfPath string

		elasticsearchUrl      string
		elasticsearchUsername string
		elasticsearchPassword string

		drsUrl      string
		drsUsername string
		drsPassword string
	)

	flag.StringVar(&vcfPath, "vcfPath", "./vcfs", "VCF Path")

	flag.StringVar(&elasticsearchUrl, "elasticsearchUrl", "https://elasticsearch.gohan.local", "Elasticsearch URL")
	flag.StringVar(&elasticsearchUsername, "elasticsearchUsername", "elastic", "Elasticsearch Username")
	flag.StringVar(&elasticsearchPassword, "elasticsearchPassword", "changeme!", "Elasticsearch Password")

	flag.StringVar(&drsUrl, "drsUrl", "https://drs.gohan.local", "DRS URL")
	flag.StringVar(&drsUsername, "drsUsername", "drsadmin", "DRS Basic Auth Gateway Username")
	flag.StringVar(&drsPassword, "drsPassword", "gohandrspassword123", "DRS Basic Auth Gateway Password")

	flag.Parse()

	fmt.Printf("Using : \n\tVCF Directory Path : %s \n\tElasticsearch Url : %s \n\tElasticsearch Username : %s\n\tDRS Url : %s\n\tDRS Username : %s\n",
		vcfPath, elasticsearchUrl, elasticsearchUsername, drsUrl, drsUsername)
	// --

	// Instantiate Server
	e := echo.New()

	// Service Connections:
	// -- Elasticsearch
	es := utils.CreateEsConnection(elasticsearchUrl, elasticsearchUsername, elasticsearchPassword)
	// -- TODO: DRS ?
	// 		(or perhaps just create an http client with credentials when necessary
	//		rather than have one global http client ?)

	// Service Singletons
	// TODO: Get Authz parameters from config or something
	az := services.NewAuthzService(true, "", "", []string{"X-AUTHN-TOKEN"})

	// Configure Server
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	// -- Override handlers with "custom Gohan" context
	//		to be able to provide variables and global singletons
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &contexts.GohanContext{c, es, vcfPath, drsUrl, drsUsername, drsPassword}
			return h(cc)
		}
	})

	// Global Middleware
	e.Use(az.MandateAuthorizationTokensMiddleware)

	// Begin MVC Routes
	// -- Root
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Welcome to the next generation Gohan v2 API using Golang!")
	})

	// -- Variants
	e.GET("variants/get/by/variantId", mvc.VariantsGetByVariantId)
	e.GET("variants/get/by/sampleId", mvc.VariantsGetBySampleId)

	e.GET("variants/count/by/variantId", mvc.VariantsCountByVariantId)
	e.GET("variants/count/by/sampleId", mvc.VariantsCountBySampleId)

	e.GET("/variants/ingest", mvc.VariantsIngestTest)

	// Run
	e.Logger.Fatal(e.Start(":" + port))
}
