package serviceInfo

import (
	"gohan/api/contexts"
	serviceInfo "gohan/api/models/constants/service-info"

	"net/http"

	"github.com/labstack/echo"
)

// Spec: https://github.com/ga4gh-discovery/ga4gh-service-info
func GetServiceInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"bento": map[string]interface{}{
			"dataService": true,
			"serviceKind": serviceInfo.SERVICE_ARTIFACT,
		},
		"type": map[string]interface{}{
			"artifact": serviceInfo.SERVICE_ARTIFACT,
			"group":    serviceInfo.SERVICE_TYPE_NO_VER,
			"version":  c.(*contexts.GohanContext).Config.SemVer,
		},
		"id":          serviceInfo.SERVICE_ID,
		"name":        serviceInfo.SERVICE_NAME,
		"description": serviceInfo.SERVICE_DESCRIPTION,
		"organization": map[string]string{
			"name": "C3G",
			"url":  "http://c3g.ca",
		},
		"contactUrl": c.(*contexts.GohanContext).Config.ServiceContact,
		"version":    c.(*contexts.GohanContext).Config.SemVer,
	})
}
