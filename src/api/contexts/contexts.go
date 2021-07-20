package contexts

import (
	es7 "github.com/elastic/go-elasticsearch"
	"github.com/labstack/echo"
)

type (
	// "Helper" Context to pass into routes
	// that need an elasticsearch client
	EsContext struct {
		echo.Context
		Es *es7.Client
	}
)
