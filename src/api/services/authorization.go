package services

import (
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

func (a *AuthzService) EnsureRepositoryAccessPermittedForUser(authnToken string) bool {
	accessPermitted := true

	// TODO:
	//	- retrieve JWKS from OIDC
	//	- validate JWKS and AuthN token agains OPA

	// Simulate permitted access :
	return accessPermitted
}

func (a *AuthzService) EnsureAllRequiredHeadersArePresent(headers http.Header) bool {
	allRequiredHeadersArePresent := true

	for _, rh := range a.GetRequiredHeaders() {
		if headers.Get(rh) == "" {
			allRequiredHeadersArePresent = false
			break
		}
	}

	return allRequiredHeadersArePresent
}

// AuthZ middleware --
func (a *AuthzService) MandateAuthorizationTokens(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.IsEnabled() {
			presentHeaders := c.Request().Header

			if a.EnsureAllRequiredHeadersArePresent(presentHeaders) {

				// TEMP
				authnTokenHeader := "X-AUTHN-TOKEN"
				authnToken := presentHeaders.Get(authnTokenHeader)

				if a.EnsureRepositoryAccessPermittedForUser(authnToken) {
					// blah
					if err := next(c); err != nil {
						c.Error(err)
					}

				} else {
					return echo.NewHTTPError(http.StatusUnauthorized, "User not permitted here!")
				}

			} else {
				return echo.NewHTTPError(http.StatusForbidden, "Missing required headers! check again")
			}
		} else {
			// AuthZ check disabled -- allow all calls
			if err := next(c); err != nil {
				c.Error(err)
			}
		}

		return nil
	}
}

// --
