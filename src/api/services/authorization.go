package services

import (
	"api/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

var publicOidcErrorMessage string = "Something went wrong interfacing with the OIDC provider! Please contact the system administrators.."
var publicOpaErrorMessage string = "Something went wrong interfacing with OPA! Please contact the system administrators.."

type (
	AuthzService struct {
		isEnabled       bool
		oidcJwksUrl     string
		opaUrl          string
		requiredHeaders []string
	}
)

func NewAuthzService(cfg *models.Config) *AuthzService {
	return &AuthzService{
		isEnabled:       cfg.AuthX.IsAuthorizationEnabled,
		oidcJwksUrl:     cfg.AuthX.OidcPublicJwksUrl,
		opaUrl:          cfg.AuthX.OpaUrl,
		requiredHeaders: strings.Split(cfg.AuthX.RequiredHeadersCommaSep, ","),
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

func (a *AuthzService) EnsureRepositoryAccessPermittedForUser(authnTokenString string) error {
	//	- retrieve JWKS from OIDC
	oidcResp, oidcErr := http.Get(a.GetOidcJwksUrl())
	if oidcErr != nil {
		fmt.Printf("%s\n", oidcErr.Error())
		return errors.New(publicOidcErrorMessage)
	}
	defer oidcResp.Body.Close()

	//	-- interpret JWKS from response
	jwksBody, readJwksBodyErr := ioutil.ReadAll(oidcResp.Body)
	if readJwksBodyErr != nil {
		fmt.Printf("%s\n", readJwksBodyErr.Error())
		return errors.New(publicOidcErrorMessage)
	}

	//	--- transform body bytes to string
	authnJwksString := string(jwksBody)

	//	-- check for OIDC json error
	var jwksJson map[string]interface{}
	jwksStringJsonUnmarshallingError := json.Unmarshal([]byte(authnJwksString), &jwksJson)
	if jwksStringJsonUnmarshallingError != nil {
		fmt.Printf("%s\n", jwksStringJsonUnmarshallingError.Error())
		return errors.New(publicOidcErrorMessage)
	}

	if jwksJsonError, doesContainErrorKey := jwksJson["error"]; doesContainErrorKey {
		fmt.Printf("Error: %s\n", jwksJsonError)
		return errors.New(publicOidcErrorMessage)
	}

	//	- validate JWKS and AuthN token agains OPA
	opaInputJson := map[string]interface{}{
		"input": map[string]interface{}{
			"authN_token": authnTokenString,
			"authN_jwks":  authnJwksString,
		},
	}

	json_data, jsonMarshallErr := json.Marshal(opaInputJson)
	if jsonMarshallErr != nil {
		fmt.Printf("%w\n", jsonMarshallErr.Error())
		return errors.New(publicOpaErrorMessage)
	}

	opaResp, opaErr := http.Post(a.GetOpaUrl(), "application/json", bytes.NewBuffer(json_data))
	if opaErr != nil {
		fmt.Printf("%s\n", opaErr.Error())
		return errors.New(publicOpaErrorMessage)
	}
	defer opaResp.Body.Close()

	var opaJson map[string]interface{}
	json.NewDecoder(opaResp.Body).Decode(&opaJson)

	if accessPermitted, isMapContainsKey := opaJson["result"]; isMapContainsKey {
		if accessPermitted != true {
			return errors.New("Access denied!")
		}
	} else {
		fmt.Printf("%s\n", "Missing 'result' key from Opa response!")
		return errors.New(publicOpaErrorMessage)
	}

	// Access permitted! Return no error
	return nil
}

func (a *AuthzService) EnsureAllRequiredHeadersArePresent(headers http.Header) error {
	// return error if anything is missing
	for _, rh := range a.GetRequiredHeaders() {
		if headers.Get(rh) == "" {
			return errors.New("Missing " + rh + " HTTP header!")
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
