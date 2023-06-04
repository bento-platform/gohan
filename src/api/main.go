package main

import (
	"gohan/api/contexts"
	gam "gohan/api/middleware"
	"gohan/api/models"
	serviceInfo "gohan/api/models/constants/service-info"
	dataTypesMvc "gohan/api/mvc/data-types"
	genesMvc "gohan/api/mvc/genes"
	serviceInfoMvc "gohan/api/mvc/service-info"
	tablesMvc "gohan/api/mvc/tables"
	variantsMvc "gohan/api/mvc/variants"
	workflowsMvc "gohan/api/mvc/workflows"
	"gohan/api/services"
	authzServices "gohan/api/services/authorization"
	"gohan/api/services/sanitation"
	variantsService "gohan/api/services/variants"
	"gohan/api/utils"
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

		"\tDebug : %t \n"+
		"\tService Contact : %s \n"+
		"\tSemantic Version : %s \n\n"+

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
		"\tAuthorization Url : %s\n"+
		"\tRequired HTTP Headers: %s\n\n"+

		"Running on Port : %s\n",

		cfg.Debug,
		cfg.ServiceContact,
		cfg.SemVer,
		cfg.Api.VcfPath,
		cfg.Api.GtfPath,
		cfg.Api.BulkIndexingCap,
		cfg.Api.FileProcessingConcurrencyLevel,
		cfg.Api.LineProcessingConcurrencyLevel,
		cfg.Elasticsearch.Url, cfg.Elasticsearch.Username,
		cfg.Api.BridgeDirectory, cfg.Drs.BridgeDirectory,
		cfg.Drs.Url, cfg.Drs.Username,
		cfg.AuthX.IsAuthorizationEnabled,
		cfg.AuthX.AuthorizationUrl,
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
	az := authzServices.NewAuthzService(&cfg)
	iz := services.NewIngestionService(es, &cfg)
	vs := variantsService.NewVariantService(&cfg)

	_ = sanitation.NewSanitationService(es, &cfg)

	// Configure Server
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	// -- Override handlers with "custom Gohan" context
	//		to be able to optimally provide access to global variables/singletons
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &contexts.GohanContext{
				Context:          c,
				Es7Client:        es,
				Config:           &cfg,
				IngestionService: iz,
				VariantService:   vs,
				// SanitationService: ss,
			}
			return h(cc)
		}
	})

	// Begin MVC Routes
	// -- Root
	e.GET("/", func(c echo.Context) error {
		fmt.Printf("[%s] - Root hit!\n", time.Now())
		return c.JSON(http.StatusOK, serviceInfo.SERVICE_WELCOME)
	})

	// -- Service Info
	e.GET("/service-info", serviceInfoMvc.GetServiceInfo)

	// -- Data-Types
	e.GET("/data-types", dataTypesMvc.GetDataTypes)
	e.GET("/data-types/variant", dataTypesMvc.GetVariantDataType)
	e.GET("/data-types/variant/schema", dataTypesMvc.GetVariantDataTypeSchema)
	e.GET("/data-types/variant/metadata_schema", dataTypesMvc.GetVariantDataTypeMetadataSchema)

	// -- Tables
	e.GET("/tables", tablesMvc.GetTables,
		// middleware
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute)
	e.POST("/tables", tablesMvc.CreateTable,
		// middleware
		// - authorization
		gam.CreateEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute)
	e.GET("/tables/:id", tablesMvc.GetTables,
		// middleware
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute)
	e.DELETE("/tables/:id", tablesMvc.DeleteTable,
		// middleware
		// - authorization
		gam.DeleteEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute)
	e.GET("/tables/:id/summary", tablesMvc.GetTableSummary,
		// middleware
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute)

	// -- Variants
	e.GET("/variants/overview", variantsMvc.GetVariantsOverview,
		// middleware
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute)

	e.GET("/variants/get/by/variantId", variantsMvc.VariantsGetByVariantId,
		// middleware
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute,
		// - everything else
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateCalibratedAlleles,
		gam.MandateAssemblyIdAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)
	e.GET("/variants/get/by/sampleId", variantsMvc.VariantsGetBySampleId,
		// middleware
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute,
		// - everything else
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateCalibratedAlleles,
		gam.MandateAssemblyIdAttribute,
		gam.MandateSampleIdsPluralAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)
	e.GET("/variants/get/by/documentId", variantsMvc.VariantsGetByDocumentId)

	e.GET("/variants/count/by/variantId", variantsMvc.VariantsCountByVariantId,
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute,
		// - everything else
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateAssemblyIdAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)
	e.GET("/variants/count/by/sampleId", variantsMvc.VariantsCountBySampleId,
		// - authorization
		gam.QueryEverythingPermissionAttribute,
		az.ValidateTokenPermissionsAttribute,
		// - everything else
		gam.ValidateOptionalChromosomeAttribute,
		gam.MandateCalibratedBounds,
		gam.MandateAssemblyIdAttribute,
		gam.MandateSampleIdsSingularAttribute,
		gam.ValidatePotentialGenotypeQueryParameter)

	// TODO: refactor (deduplicate) --
	e.GET("/variants/ingestion/run", variantsMvc.VariantsIngest,
		// middleware
		gam.MandateAssemblyIdAttribute,
		gam.MandateTableIdAttribute)
	e.GET("/variants/ingestion/requests", variantsMvc.GetAllVariantIngestionRequests)
	e.GET("/variants/ingestion/stats", variantsMvc.VariantsIngestionStats)

	e.GET("/private/variants/ingestion/run", variantsMvc.VariantsIngest,
		// middleware
		gam.MandateAssemblyIdAttribute,
		gam.MandateTableIdAttribute)
	e.GET("/private/variants/ingestion/requests", variantsMvc.GetAllVariantIngestionRequests)
	// --

	// -- Genes
	e.GET("/genes/overview", genesMvc.GetGenesOverview)
	e.GET("/genes/search", genesMvc.GenesGetByNomenclatureWildcard,
		// middleware
		gam.ValidateOptionalChromosomeAttribute)
	e.GET("/genes/ingestion/requests", genesMvc.GetAllGeneIngestionRequests)
	e.GET("/genes/ingestion/run", genesMvc.GenesIngest)
	e.GET("/genes/ingestion/stats", genesMvc.GenesIngestionStats)

	// -- Workflows
	e.GET("/workflows", workflowsMvc.WorkflowsGet)
	e.GET("/workflows/:file", workflowsMvc.WorkflowsServeFile)

	// Run
	e.Logger.Fatal(e.Start(":" + cfg.Api.Port))
}
