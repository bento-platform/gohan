package workflows

import (
	w "gohan/api/workflows"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

// workflowsDir is where WDL files are stored; a variable so tests can override it.
var workflowsDir = "/app/workflows"

func WorkflowsGet(c echo.Context) error {
	return c.JSON(http.StatusOK, w.WORKFLOW_VARIANT_SCHEMA)
}

func WorkflowsServeFile(c echo.Context) error {
	name := c.Param("file")

	// "/workflows/vcf_gz.wdl" -> serve the WDL file
	if strings.HasSuffix(name, ".wdl") {
		if name != filepath.Base(name) {
			return c.JSON(http.StatusBadRequest, "Invalid workflow file name")
		}
		return c.File(filepath.Join(workflowsDir, name))
	}

	// "/workflows/vcf_gz" -> serve the JSON definition
	for _, group := range w.WORKFLOW_VARIANT_SCHEMA {
		if defs, ok := group.(map[string]interface{}); ok {
			if def, found := defs[name]; found {
				return c.JSON(http.StatusOK, def)
			}
		}
	}
	return c.JSON(http.StatusNotFound, "Unknown workflow ID")
}
