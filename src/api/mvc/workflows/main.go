package workflows

import (
	"fmt"
	w "gohan/api/workflows"
	"net/http"

	"github.com/labstack/echo"
)

func WorkflowsGet(c echo.Context) error {
	return c.JSON(http.StatusOK, w.WORKFLOW_VARIANT_SCHEMA)
}

func WorkflowsServeFile(c echo.Context) error {
	// retrieve wdl from storage and send to client
	if len(c.ParamValues()) > 0 && len(c.ParamValues()) < 2 {
		fileName := c.ParamValues()[0]
		return c.File(fmt.Sprintf("/app/workflows/%s", fileName))
	} else {
		return c.JSON(http.StatusBadRequest, "Invalid Request! Please only specify a filename; example : /workflows/vcf_gz.wdl")
	}
}
