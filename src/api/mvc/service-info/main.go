package serviceInfo

import (
	serviceInfo "api/models/constants/service-info"

	"net/http"

	"github.com/labstack/echo"
)

// Spec: https://github.com/ga4gh-discovery/ga4gh-service-info
func GetServiceInfo(c echo.Context) error {
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
}
