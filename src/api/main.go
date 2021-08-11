package main

import (
	"api/contexts"
	gam "api/middleware"
	"api/models"
	"api/mvc"
	"api/services"
	"api/utils"
	"strings"

	"fmt"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Gather environment variables
	var cfg models.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Printf("Using : \n"+
		"\tVCF Directory Path : %s \n"+
		"\tElasticsearch Url : %s \n"+
		"\tElasticsearch Username : %s\n\n"+

		"\tDRS Url : %s\n"+
		"\tDRS Username : %s\n\n"+

		"\tAuthorization Enabled : %b\n"+
		"\tOIDC Public JWKS Url : %s\n"+
		"\tOPA Url : %s\n"+
		"\tRequired HTTP Headers: %s\n\n"+

		"Running on Port : %s\n",

		cfg.Api.VcfPath,
		cfg.Elasticsearch.Url, cfg.Elasticsearch.Username,
		cfg.Drs.Url, cfg.Drs.Username,
		cfg.AuthX.IsAuthorizationEnabled,
		cfg.AuthX.OidcPublicJwksUrl,
		cfg.AuthX.OpaUrl,
		strings.Split(cfg.AuthX.RequiredHeadersCommaSep, ","),
		cfg.Api.Port)
	// --

	// Instantiate Server
	e := echo.New()

	// Service Connections:
	// -- Elasticsearch
	es := utils.CreateEsConnection(&cfg)
	// -- TODO: DRS ?
	// 		(or perhaps just create an http client with credentials when necessary
	//		rather than have one global http client ?)

	// Service Singletons
	az := services.NewAuthzService(&cfg)
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
			cc := &contexts.GohanContext{c, es, &cfg, *iz}
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
	e.Logger.Fatal(e.Start(":" + cfg.Api.Port))
}
