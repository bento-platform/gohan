package api

import (
	"fmt"
	common "gohan/api/tests/common"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithInvalidAuthenticationToken(t *testing.T) {
	cfg := common.InitConfig()

	request, _ := http.NewRequest("GET", cfg.Api.Url, nil)

	request.Header.Add("X-AUTHN-TOKEN", "gibberish")

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// default response without a valid authentication token is is 401; consider it a pass
	var shouldBe int
	if cfg.AuthX.IsAuthorizationEnabled {
		shouldBe = 401
	} else {
		shouldBe = 200
	}

	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET / Status: %s ; Should be %d", response.Status, shouldBe))
}

// -- Common utility functions for api tests
