package contexts

import (
	"gohan/api/models"
	"gohan/api/models/constants"
	"gohan/api/services"
	variantsService "gohan/api/services/variants"

	es7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo"
)

type (
	// "Helper" Context to pass into routes that need
	//  an elasticsearch client and other variables
	GohanContext struct {
		echo.Context
		Es7Client        *es7.Client
		Config           *models.Config
		IngestionService *services.IngestionService
		VariantService   *variantsService.VariantService
		QueryParameters
	}

	// Convenient storage for relevant http context data
	QueryParameters struct {
		AssemblyId constants.AssemblyId
		Alleles    []string
	}
)
