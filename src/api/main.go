package main

import (
	"api/contexts"
	gam "api/middleware"
	"api/models"
	serviceInfo "api/models/constants/service-info"
	"api/mvc"
	"api/services"
	"api/utils"
	"strings"
	"time"

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

		"\tDebug : %t \n\n"+

		"\tVCF Directory Path : %s \n"+
		"\tGTF Directory Path : %s \n"+
		"\tBulk Indexing Cap : %d\n"+
		"\tFile Processing Concurrency Level : %d\n"+
		"\tLine Processing Concurrency Level : %d\n"+
		"\tElasticsearch Url : %s \n"+
		"\tElasticsearch Username : %s\n\n"+

		"\tAPI's API-DRS Bridge Directory : %s\n"+
		"\tDRS's API-DRS Bridge Directory : %s\n"+
		"\tDRS Url : %s\n"+
		"\tDRS Username : %s\n\n"+

		"\tAuthorization Enabled : %t\n"+
		"\tOIDC Public JWKS Url : %s\n"+
		"\tOPA Url : %s\n"+
		"\tRequired HTTP Headers: %s\n\n"+

		"Running on Port : %s\n",

		cfg.Debug,
		cfg.Api.VcfPath,
		cfg.Api.GtfPath,
		cfg.Api.BulkIndexingCap,
		cfg.Api.FileProcessingConcurrencyLevel,
		cfg.Api.LineProcessingConcurrencyLevel,
		cfg.Elasticsearch.Url, cfg.Elasticsearch.Username,
		cfg.Api.BridgeDirectory, cfg.Drs.BridgeDirectory,
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
	iz := services.NewIngestionService(es, &cfg)

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
			cc := &contexts.GohanContext{
				Context:          c,
				Es7Client:        es,
				Config:           &cfg,
				IngestionService: iz,
			}
			return h(cc)
		}
	})

	// Global Middleware
	e.Use(az.MandateAuthorizationTokensMiddleware)

	// Begin MVC Routes
	// -- Root
	e.GET("/", func(c echo.Context) error {
		fmt.Printf("[%s] - Root hit!\n", time.Now())
		return c.JSON(http.StatusOK, serviceInfo.SERVICE_WELCOME)
	})

	// -- Service Info
	e.GET("/service-info", func(c echo.Context) error {
		// Spec: https://github.com/ga4gh-discovery/ga4gh-service-info
		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":          serviceInfo.SERVICE_ID,
			"name":        serviceInfo.SERVICE_NAME,
			"type":        serviceInfo.SERVICE_TYPE,
			"description": serviceInfo.SERVICE_DESCRIPTION,
			"organization": map[string]string{
				"name": "C3G",
				"url":  "http://c3g.ca",
			},
			"contactUrl": serviceInfo.SERVICE_CONTACT,
			"version":    serviceInfo.SERVICE_VERSION,
		})
	})

	// -- Data-Type
	e.GET("/data-types", mvc.GetDataTypes)
	e.GET("/data-types/variant", mvc.GetVariantDataType)
	e.GET("/data-types/variant/schema", mvc.GetVariantDataTypeSchema)
	e.GET("/data-types/variant/metadata_schema", mvc.GetVariantDataTypeMetadataSchema)

	// -- TEMP (simulating bento_variants)
	e.GET("/tables", mvc.FakeBentoTables)
	e.GET("/tables/fake", mvc.FakeBentoTableSchema)
	// --

	// -- Variants
	e.GET("/variants/overview", mvc.GetVariantsOverview)

	e.GET("/variants/get/by/variantId", mvc.VariantsGetByVariantId,
		// middleware
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateAssemblyIdAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)
	e.GET("/variants/get/by/sampleId", mvc.VariantsGetBySampleId,
		// middleware
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateAssemblyIdAttribute,
		gam.MandateSampleIdsPluralAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)
	e.GET("/variants/get/by/documentId", mvc.VariantsGetByDocumentId)

	e.GET("/variants/count/by/variantId", mvc.VariantsCountByVariantId,
		// middleware
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateAssemblyIdAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)
	e.GET("/variants/count/by/sampleId", mvc.VariantsCountBySampleId,
		// middleware
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateAssemblyIdAttribute,
		gam.MandateSampleIdsSingularAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)

	// TODO: refactor (deduplicate) --
	e.GET("/variants/ingestion/run", mvc.VariantsIngest,
		// middleware
		gam.MandateAssemblyIdAttribute)
	e.GET("/variants/ingestion/requests", mvc.GetAllVariantIngestionRequests)
	e.GET("/variants/ingestion/stats", mvc.VariantsIngestionStats)

	e.GET("/private/variants/ingestion/run", mvc.VariantsIngest,
		// middleware
		gam.MandateAssemblyIdAttribute)
	e.GET("/private/variants/ingestion/requests", mvc.GetAllVariantIngestionRequests)
	// --

	// -- Genes
	e.GET("/genes/overview", mvc.GetGenesOverview)
	e.GET("/genes/search", mvc.GenesGetByNomenclatureWildcard,
		// middleware
		gam.ValidateOptionalChromosomeAttribute)
	e.GET("/genes/ingestion/requests", mvc.GetAllGeneIngestionRequests)
	e.GET("/genes/ingestion/run", mvc.GenesIngest)
	e.GET("/genes/ingestion/stats", mvc.VariantsIngestionStats)

	// -- Workflows
	e.GET("/workflows", mvc.WorkflowsGet)
	e.GET("/workflows/:file", mvc.WorkflowsServeFile)

	// Run
	e.Logger.Fatal(e.Start(":" + cfg.Api.Port))
}
