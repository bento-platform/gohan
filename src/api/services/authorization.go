package services

import (
	"errors"
	"net/http"

	"github.com/labstack/echo"
)

type (
	AuthzService struct {
		isEnabled       bool
		oidcJwksUrl     string
		opaUrl          string
		requiredHeaders []string
	}
)

func NewAuthzService(_isEnabled bool, _oidcJwksUrl string, _opaUrl string, _requiredHeaders []string) *AuthzService {
	return &AuthzService{
		isEnabled:       _isEnabled,
		oidcJwksUrl:     _oidcJwksUrl,
		opaUrl:          _opaUrl,
		requiredHeaders: _requiredHeaders,
	}
}

func (a *AuthzService) IsEnabled() bool {
	return a.isEnabled
}

func (a *AuthzService) GetOidcJwksUrl() string {
	return a.oidcJwksUrl
}

func (a *AuthzService) GetOpaUrl() string {
	return a.opaUrl
}

func (a *AuthzService) GetRequiredHeaders() []string {
	return a.requiredHeaders
}

func (a *AuthzService) EnsureRepositoryAccessPermittedForUser(authnToken string) error {
	// TODO:
	//	- retrieve JWKS from OIDC
	//	- validate JWKS and AuthN token agains OPA

	// Simulate permitted access :
	return nil
}

func (a *AuthzService) EnsureAllRequiredHeadersArePresent(headers http.Header) error {
	// return error if anything is missing
	for _, rh := range a.GetRequiredHeaders() {
		if headers.Get(rh) == "" {
			return errors.New("Missing " + rh + " HTTP Header!")
		}
	}
	return nil
}

func (a *AuthzService) MandateAuthorizationTokensMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.IsEnabled() {
			// check headers
			presentHeaders := c.Request().Header

			missingHeadersErr := a.EnsureAllRequiredHeadersArePresent(presentHeaders)
			if missingHeadersErr != nil {
				return echo.NewHTTPError(http.StatusForbidden, missingHeadersErr.Error())
			}

			// TEMP
			authnTokenHeader := "X-AUTHN-TOKEN"
			authnToken := presentHeaders.Get(authnTokenHeader)

			// check user permission
			accessError := a.EnsureRepositoryAccessPermittedForUser(authnToken)
			if accessError != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, accessError.Error())
			}
		}

		// access granted!
		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
