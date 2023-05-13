package middleware

import (
	"gohan/api/contexts"
	authzModels "gohan/api/models/authorization"
	authzConstants "gohan/api/models/constants/authorization"

	"github.com/labstack/echo"
)

func QueryDataPermissionAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)
		gc.RequestedResource = authzModels.ResourceEverything{
			Everything: true,
		}
		gc.RequiredPermissions = []authzModels.Permission{{
			Verb: authzConstants.QUERY,
			Noun: authzConstants.DATA,
		}}
		return next(gc)
	}
}
