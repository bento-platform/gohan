package api

import (
	"api/models"
	"api/models/dtos"
	"api/models/indexes"
	"api/utils"
	"bytes"
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
	PostCreateTablePath             string = "%s/tables"
)

func TestVariantGetTables(t *testing.T) {
	cfg := common.InitConfig()

	// get all available 'variant' tables
	allTableDtos := getVariantTables(t, cfg)
	assert.NotNil(t, allTableDtos)
}

func TestCanCreateTable(t *testing.T) {
	cfg := common.InitConfig()

	// prepare request
	postCreateTableUrl := fmt.Sprintf(PostCreateTablePath, cfg.Api.Url)
	data := dtos.CreateTableRequestDto{
		Name:     utils.RandomString(32),   // random table name
		DataType: "variant",                // set variant data_type
		Dataset:  utils.RandomString(32),   // random dataset name
		Metadata: map[string]interface{}{}, // TODO : expand ?
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	dataString := string(dataBytes)

	r, _ := http.NewRequest("POST", postCreateTableUrl, bytes.NewBufferString(dataString))
	r.Header.Add("Content-Type", "application/json")

	// perform request
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Table Creation error: %s\n", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Table Creation status: %d\n", resp.StatusCode)

	// obtain the newly created table
	//	-- interpret create-table dto from response
	createTableRespBody, createTableRespBodyErr := ioutil.ReadAll(resp.Body)
	assert.Nil(t, createTableRespBodyErr)

	//	--- transform body bytes to string
	createTableRespBodyString := string(createTableRespBody)

	//	-- check for json error
	var createTablesRespJson dtos.CreateTableResponseDto
	createTableJsonUnmarshallingError := json.Unmarshal([]byte(createTableRespBodyString), &createTablesRespJson)
	assert.Nil(t, createTableJsonUnmarshallingError)

	// -- ensure table was successfully created
	assert.Empty(t, createTablesRespJson.Error)

	assert.NotNil(t, createTablesRespJson.Table)
	assert.NotNil(t, createTablesRespJson.Table.Id)
	assert.NotEmpty(t, createTablesRespJson.Table.Id)

	// test get-by-id with newly created table
	newTableId := createTablesRespJson.Id
	getTableByIdUrl := fmt.Sprintf(GetTableByIdPathWithPlaceholder, cfg.Api.Url, newTableId)

	// TODO: refactor
	// ================
	request, _ := http.NewRequest("GET", getTableByIdUrl, nil)

	client = &http.Client{}
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
	var getTableByIdResp indexes.Table
	getTableByIdRespUnmarshallingError := json.Unmarshal([]byte(tableRespBodyString), &getTableByIdResp)
	assert.Nil(t, getTableByIdRespUnmarshallingError)

	// ================

	// -- ensure the table ids are the same
	assert.True(t, getTableByIdResp.Id == newTableId)

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
