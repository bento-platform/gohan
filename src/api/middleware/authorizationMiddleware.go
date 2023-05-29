package middleware

import (
	"gohan/api/contexts"
	authzModels "gohan/api/models/authorization"
	authzConstants "gohan/api/models/constants/authorization"

	"github.com/labstack/echo"
)

func ViewEverythingPermissionAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)
		addResourceEverything(gc)
		addPermissions(gc, authzConstants.VIEW, authzConstants.GOHAN)
		return next(gc)
	}
}
func QueryEverythingPermissionAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)
		addResourceEverything(gc)
		addPermissions(gc, authzConstants.QUERY, authzConstants.GOHAN)
		return next(gc)
	}
}
func CreateEverythingPermissionAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)
		addResourceEverything(gc)
		addPermissions(gc, authzConstants.CREATE, authzConstants.GOHAN)
		return next(gc)
	}
}
func DeleteEverythingPermissionAttribute(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		gc := c.(*contexts.GohanContext)
		addResourceEverything(gc)
		addPermissions(gc, authzConstants.DELETE, authzConstants.GOHAN)
		return next(gc)
	}
}

// -- helper functions
func addResourceEverything(gc *contexts.GohanContext) {
	gc.RequestedResource = authzModels.ResourceEverything{
		Everything: true,
	}
}
func addPermissions(gc *contexts.GohanContext, verb authzConstants.PermissionVerb, noun authzConstants.PermissionNoun) {
	gc.RequiredPermissions = []authzModels.Permission{{
		Verb: verb,
		Noun: noun,
	}}
}
