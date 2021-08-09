package contexts

import (
	"api/models"

	es7 "github.com/elastic/go-elasticsearch"
	"github.com/labstack/echo"
)

type (
	// "Helper" Context to pass into routes that need
	//  an elasticsearch client and other variables
	GohanContext struct {
		echo.Context
		Es7Client            *es7.Client
		VcfPath              string
		DrsUrl               string
		DrsUsername          string
		DrsPassword          string
		IngestRequestChannel chan *models.IngestRequest
	}
)
