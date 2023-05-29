package serviceInfo

import (
	"fmt"
	"gohan/api/contexts"
	serviceInfo "gohan/api/models/constants/service-info"

	"net/http"

	"github.com/labstack/echo"
)

// Spec: https://github.com/ga4gh-discovery/ga4gh-service-info
func GetServiceInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bento": map[string]interface{}{
			"serviceKind": "gohan",
		},
		"id":   serviceInfo.SERVICE_ID,
		"name": serviceInfo.SERVICE_NAME,
		"type": fmt.Sprintf("%s:%s", serviceInfo.SERVICE_TYPE_NO_VER, c.(*contexts.GohanContext).Config.SemVer),

		"description": serviceInfo.SERVICE_DESCRIPTION,
		"organization": map[string]string{
			"name": "C3G",
			"url":  "http://c3g.ca",
		},
		"contactUrl": c.(*contexts.GohanContext).Config.ServiceContact,
		"version":    c.(*contexts.GohanContext).Config.SemVer,
	})
}
