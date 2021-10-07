package mvc

import (
	w "api/workflows"
	"github.com/labstack/echo"
	"net/http"
)

func WorkflowsGet(c echo.Context) error {
	return c.JSON(http.StatusOK, w.WORKFLOW_VARIANT_SCHEMA)
}
