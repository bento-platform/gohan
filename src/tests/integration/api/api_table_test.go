package api

import (
	"api/models"
	"api/models/indexes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	common "tests/common"

	"github.com/stretchr/testify/assert"
)

const (
	GetVariantTablesPath            string = "%s/tables?data-type=variant"
	GetTableByIdPathWithPlaceholder string = "%s/tables/%s"
)

func TestVariantGetTables(t *testing.T) {
	cfg := common.InitConfig()

	// get all available 'variant' tables
	allTableDtos := getVariantTables(t, cfg)
	assert.NotNil(t, allTableDtos)
}

func TestCanGetAllTablesById(t *testing.T) {
	cfg := common.InitConfig()

	allTableDtos := getVariantTables(t, cfg)
	assert.NotNil(t, allTableDtos)
	assert.True(t, len(allTableDtos) > 0)

	for _, table := range allTableDtos {

		tableId := table.Id
		getTableByIdUrl := fmt.Sprintf(GetTableByIdPathWithPlaceholder, cfg.Api.Url, tableId)

		// TODO: refactor
		// ================
		request, _ := http.NewRequest("GET", getTableByIdUrl, nil)

		client := &http.Client{}
		response, responseErr := client.Do(request)
		assert.Nil(t, responseErr)

		defer response.Body.Close()

		// this test (at the time of writing) will only work if authorization is disabled
		shouldBe := 200
		assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET %s Status: %s ; Should be %d", getTableByIdUrl, response.Status, shouldBe))

		//	-- interpret array of available tables from response
		tableRespBody, tableRespBodyErr := ioutil.ReadAll(response.Body)
		assert.Nil(t, tableRespBodyErr)

		//	--- transform body bytes to string
		tableRespBodyString := string(tableRespBody)

		//	-- check for json error
		var tablesRespJson indexes.Table
		tableJsonUnmarshallingError := json.Unmarshal([]byte(tableRespBodyString), &tablesRespJson)
		assert.Nil(t, tableJsonUnmarshallingError)

		// ================

		// -- ensure the table ids are the same
		assert.True(t, tablesRespJson.Id == tableId)
	}
}

func getVariantTables(_t *testing.T, _cfg *models.Config) []indexes.Table {
	url := fmt.Sprintf(GetVariantTablesPath, _cfg.Api.Url)
	request, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(_t, responseErr)

	defer response.Body.Close()

	// this test (at the time of writing) will only work if authorization is disabled
	shouldBe := 200
	assert.Equal(_t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET %s Status: %s ; Should be %d", url, response.Status, shouldBe))

	//	-- interpret array of available tables from response
	overviewRespBody, overviewRespBodyErr := ioutil.ReadAll(response.Body)
	assert.Nil(_t, overviewRespBodyErr)

	//	--- transform body bytes to string
	overviewRespBodyString := string(overviewRespBody)

	//	-- check for json error
	var tableDtos []indexes.Table
	overviewJsonUnmarshallingError := json.Unmarshal([]byte(overviewRespBodyString), &tableDtos)
	assert.Nil(_t, overviewJsonUnmarshallingError)

	return tableDtos
}
