package elasticsearch

import (
	"fmt"
	common "gohan/api/tests/common"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityWithoutBasicAuth(t *testing.T) {
	cfg := common.InitConfig()

	request, requestErr := http.NewRequest("GET", cfg.Elasticsearch.Url, nil)
	assert.Nil(t, requestErr)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// response status code from a basic-auth-secured elasticsearch
	// without valid (or any) credentials is 401 ; consider it a pass
	shouldBe := 401
	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Elasticsearch GET / Status: %s ; Should be %d", response.Status, shouldBe))
}

func TestSecurityWithBasicAuth(t *testing.T) {
	cfg := common.InitConfig()

	request, requestErr := http.NewRequest("GET", cfg.Elasticsearch.Url, nil)
	assert.Nil(t, requestErr)

	request.SetBasicAuth(cfg.Elasticsearch.Username, cfg.Elasticsearch.Password)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// response status code from a basic-auth-secured elasticsearch
	// with valid credentials ; consider it a pass
	shouldBe := 200
	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Elasticsearch GET / Status: %s ; Should be %d", response.Status, shouldBe))
}
