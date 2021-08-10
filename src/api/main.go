package main

import (
	"api/contexts"
	gam "api/middleware"
	"api/mvc"
	"api/services"
	"api/utils"
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
	var (
		port    string
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

	port = os.Getenv("GOHAN_API_INTERNAL_PORT")

	vcfPath = os.Getenv("GOHAN_API_VCF_PATH")

	elasticsearchUrl = os.Getenv("GOHAN_ES_URL")
	elasticsearchUsername = os.Getenv("GOHAN_ES_USERNAME")
	elasticsearchPassword = os.Getenv("GOHAN_ES_PASSWORD")

	drsUrl = os.Getenv("GOHAN_DRS_URL")
	drsUsername = os.Getenv("GOHAN_DRS_BASIC_AUTH_USERNAME")
	drsPassword = os.Getenv("GOHAN_DRS_BASIC_AUTH_PASSWORD")

	isAuthorizationEnabled = os.Getenv("GOHAN_AUTHZ_ENABLED")
	oidcPublicJwksUrl = os.Getenv("GOHAN_PUBLIC_AUTHN_JWKS_URL")
	opaUrl = os.Getenv("GOHAN_PRIVATE_AUTHZ_URL")
	requiredHeadersCommaSep = os.Getenv("GOHAN_AUTHZ_REQHEADS")

	fmt.Printf("Using : \n"+
		"\tVCF Directory Path : %s \n"+
		"\tElasticsearch Url : %s \n"+
		"\tElasticsearch Username : %s\n\n"+

		"\tDRS Url : %s\n"+
		"\tDRS Username : %s\n\n"+

		"\tAuthorization Enabled : %s\n"+
		"\tOIDC Public JWKS Url : %s\n"+
		"\tOPA Url : %s\n"+
		"\tRequired HTTP Headers: %s\n\n"+

		"Running on Port : %s\n",

		vcfPath,
		elasticsearchUrl, elasticsearchUsername,
		drsUrl, drsUsername,
		isAuthorizationEnabled,
		oidcPublicJwksUrl,
		opaUrl,
		strings.Split(requiredHeadersCommaSep, ","),
		port)

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

	iz := services.NewIngestionService()

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
			cc := &contexts.GohanContext{c, es, vcfPath, drsUrl, drsUsername, drsPassword, *iz}
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
	e.GET("/variants/get/by/variantId", mvc.VariantsGetByVariantId,
		// middleware
		gam.MandateChromosomeAttribute,
		gam.MandateCalibratedBounds)
	e.GET("/variants/get/by/sampleId", mvc.VariantsGetBySampleId,
		// middleware
		gam.MandateChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateSampleIdsPluralAttribute)

	e.GET("/variants/count/by/variantId", mvc.VariantsCountByVariantId,
		// middleware
		gam.MandateChromosomeAttribute,
		gam.MandateCalibratedBounds)
	e.GET("/variants/count/by/sampleId", mvc.VariantsCountBySampleId,
		// middleware
		gam.MandateChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateSampleIdsSingularAttribute)

	e.GET("/variants/ingestion/run", mvc.VariantsIngest)
	e.GET("/variants/ingestion/requests", mvc.GetAllVariantIngestionRequests)

	// Run
	e.Logger.Fatal(e.Start(":" + port))
}
