package main

import (
	"api/contexts"
	"api/mvc"
	"api/services"
	"api/utils"
	"flag"
	"log"
	"strconv"
	"strings"

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

		isAuthorizationEnabled  string
		oidcPublicJwksUrl       string
		opaUrl                  string
		requiredHeadersCommaSep string
	)

	flag.StringVar(&vcfPath, "vcfPath", "./vcfs", "VCF Path")

	flag.StringVar(&elasticsearchUrl, "elasticsearchUrl", "https://elasticsearch.gohan.local", "Elasticsearch URL")
	flag.StringVar(&elasticsearchUsername, "elasticsearchUsername", "elastic", "Elasticsearch Username")
	flag.StringVar(&elasticsearchPassword, "elasticsearchPassword", "changeme!", "Elasticsearch Password")

	flag.StringVar(&drsUrl, "drsUrl", "https://drs.gohan.local", "DRS URL")
	flag.StringVar(&drsUsername, "drsUsername", "drsadmin", "DRS Basic Auth Gateway Username")
	flag.StringVar(&drsPassword, "drsPassword", "gohandrspassword123", "DRS Basic Auth Gateway Password")

	flag.StringVar(&isAuthorizationEnabled, "isAuthorizationEnabled", "true", "Authorization On/Off Toggle")
	flag.StringVar(&oidcPublicJwksUrl, "oidcPublicJwksUrl", "http://localhost:8080/auth/realms/bento/protocol/openid-connect/certs", "OIDC Public JWKS URL")
	flag.StringVar(&opaUrl, "opaUrl", "http://localhost:8181/v1/data/permissions/allowed", "OPA URL")
	flag.StringVar(&requiredHeadersCommaSep, "requiredHeadersCommaSep", "X-AUTHN-TOKEN", "Required HTTP Headers")

	flag.Parse()

	fmt.Printf(
		"Using : \n\tVCF Directory Path : %s \n"+
			"\tElasticsearch Url : %s \n"+
			"\tElasticsearch Username : %s\n\n"+

			"\tDRS Url : %s\n"+
			"\tDRS Username : %s\n\n"+

			"\tAuthorization Enabled : %s\n"+
			"\tOIDC Public JWKS Url : %s\n"+
			"\tOPA Url : %s\n"+
			"\tRequired HTTP Headers: %s\n",
		vcfPath,
		elasticsearchUrl, elasticsearchUsername,
		drsUrl, drsUsername,
		isAuthorizationEnabled,
		oidcPublicJwksUrl,
		opaUrl,
		strings.Split(requiredHeadersCommaSep, ","))

	parsedBool, boolParseErr := strconv.ParseBool(isAuthorizationEnabled)
	if boolParseErr != nil {
		log.Fatalf("Could not parse 'isAuthorizationEnabled' \"%s\" as true or false", isAuthorizationEnabled)
	}
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
	az := services.NewAuthzService(bool(parsedBool),
		oidcPublicJwksUrl, opaUrl,
		strings.Split(requiredHeadersCommaSep, ","))

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
