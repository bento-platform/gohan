package gateway

import (
	"fmt"
	common "gohan/api/tests/common"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoesApiServiceOverHttpRedirectToHttps(t *testing.T) {
	cfg := common.InitConfig()
	makeTheCallAndVerify(t, cfg.Api.Url)
}
func TestDoesDrsServiceOverHttpRedirectToHttps(t *testing.T) {
	cfg := common.InitConfig()
	makeTheCallAndVerify(t, cfg.Drs.Url)
}

func TestDoesElasticsearchServiceOverHttpRedirectToHttps(t *testing.T) {
	cfg := common.InitConfig()
	makeTheCallAndVerify(t, cfg.Elasticsearch.Url)
}

func makeTheCallAndVerify(_t *testing.T, url string) {

	// ** assumes url is already https and port is 443
	replacedHttpsWithHttpAndRemovedPortUrl := strings.Replace(strings.Replace(url, "https", "http", -1), ":443", "", -1)

	request, _ := http.NewRequest("GET", replacedHttpsWithHttpAndRemovedPortUrl, nil)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { // deactivate automatic redirect handling
			return http.ErrUseLastResponse
		}}

	response, responseErr := client.Do(request)
	assert.Nil(_t, responseErr)

	defer response.Body.Close()

	// default response without a valid authentication token is is 401; consider it a pass
	shouldBe := 301

	assert.Equal(_t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- GET / Status: %s ; Should be %d", response.Status, shouldBe))
}
