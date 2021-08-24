package drs

import (
	"fmt"
	"net/http"
	"testing"
	common "tests/common"

	"github.com/stretchr/testify/assert"
)

func TestIsDrsRunning(t *testing.T) {
	cfg := common.InitConfig()

	request, requestErr := http.NewRequest("GET", cfg.Drs.Url, nil)
	assert.Nil(t, requestErr)

	request.SetBasicAuth(cfg.Drs.Username, cfg.Drs.Password)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// default response from DRS is 404; consider it a pass
	shouldBe := 404
	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- DRS GET / Status: %s ; Should be %d", response.Status, shouldBe))
}
