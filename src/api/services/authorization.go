package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gohan/api/models"
	authz "gohan/api/models/authorization"
	authzc "gohan/api/models/constants/authorization"
	dtos "gohan/api/models/dtos/authorization"
	e "gohan/api/models/dtos/errors"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

var publicAuthzErrorMessage string = "Something went wrong interfacing with the authorization service! Please contact the system administrators.."

type (
	AuthzService struct {
		isEnabled        bool
		authorizationUrl string
	}
)

func NewAuthzService(cfg *models.Config) *AuthzService {
	return &AuthzService{
		isEnabled:        cfg.AuthX.IsAuthorizationEnabled,
		authorizationUrl: cfg.AuthX.AuthorizationUrl,
	}
}

func (a *AuthzService) IsEnabled() bool {
	return a.isEnabled
}

func (a *AuthzService) GetAuthorizationUrl() string {
	return a.authorizationUrl
}

func (a *AuthzService) EnsureRepositoryAccessPermittedForUser(authnTokenString string) error {
	//	- validate authn token against external authorization service
	permissionRequestJson := dtos.PermissionRequestDto{
		RequestedResource: authz.ResourceEverything{Everything: true},
		RequiredPermissions: authz.PermissionsList{
			List: []authz.Permission{
				{Verb: authzc.QUERY, Noun: authzc.DATA},
			},
		},
	}

	permJsonData, permissionJsonMarshallErr := json.Marshal(&permissionRequestJson)
	if permissionJsonMarshallErr != nil {
		fmt.Printf("%s\n", permissionJsonMarshallErr.Error())
		return errors.New(publicAuthzErrorMessage)
	}

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + authnTokenString

	// Create a new request using http
	evaluateUrl := fmt.Sprintf("%s/%s/%s", a.GetAuthorizationUrl(), "policy", "evaluate")
	permReq, permReqErr := http.NewRequest("POST", evaluateUrl, bytes.NewBuffer(permJsonData))
	if permReqErr != nil {
		fmt.Printf("%s\n", permReqErr.Error())
		return errors.New(publicAuthzErrorMessage)
	}
	// add authorization header to the req
	permReq.Header.Add("Authorization", bearer)
	permReq.Header.Add("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
	permRes, permResErr := client.Do(permReq)
	if permResErr != nil {
		fmt.Printf("%s\n", permResErr.Error())
		return errors.New(publicAuthzErrorMessage)
	}

	defer permRes.Body.Close()

	// check http status code
	if permRes.StatusCode != 200 {
		return errors.New("access denied")
	}

	// fetch response body if permitted
	var permJson map[string]interface{}
	json.NewDecoder(permRes.Body).Decode(&permJson)

	if accessPermitted, isMapContainsKey := permJson["result"]; isMapContainsKey {
		if !(accessPermitted.(bool)) {
			return errors.New("access denied")
		}
	} else {
		fmt.Printf("%s\n", "Missing 'result' key from authorization service response!")
		return errors.New(publicAuthzErrorMessage)
	}

	// Access permitted! Return no error
	return nil
}

func (a *AuthzService) FetchAuthorizationHeader(headers http.Header) (string, error) {
	// return error if the Authorization header is missing
	if headers.Get("Authorization") == "" {
		return "", errors.New("missing 'Authorization' HTTP header")
	}

	authnToken := headers.Get("Authorization")
	// remove "Bearer " if need be, assuming the header is properly formatted
	if strings.Contains(authnToken, "Bearer") {
		authnToken = strings.Split(authnToken, " ")[1]
	}

	return authnToken, nil
}

func (a *AuthzService) MandateAuthorizationTokensMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if a.IsEnabled() {
			// check request headers
			authnToken, missingHeaderErr := a.FetchAuthorizationHeader(c.Request().Header)
			if missingHeaderErr != nil {
				return echo.NewHTTPError(http.StatusForbidden, missingHeaderErr.Error())
			}

			// check user permission
			accessError := a.EnsureRepositoryAccessPermittedForUser(authnToken)
			if accessError != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, e.CreateSimpleUnauthorized(accessError.Error()))
			}
		}

		// access granted!

		// pass context off to the next middleware handler
		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
